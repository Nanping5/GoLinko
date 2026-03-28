package chat

import (
	"GoLinko/internal/dao"
	"GoLinko/internal/dto/request"
	"GoLinko/internal/model"
	"GoLinko/internal/service/kafka"
	myredis "GoLinko/internal/service/redis"
	constants "GoLinko/pkg/const"
	"GoLinko/pkg/enum/group_info/group_status_enum"
	"GoLinko/pkg/enum/message/message_status_enum"
	"GoLinko/pkg/enum/message/message_type_enum"
	"GoLinko/pkg/utils"
	"GoLinko/pkg/zlog"
	"encoding/json"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type Server struct {
	Clients  map[string]*Client //key是用户id
	mutex    *sync.RWMutex      // 改为 RWMutex 支持并发读
	Transmit chan []byte        //转发通道
	Login    chan *Client       //登录通道
	Logout   chan *Client       //退出登录

	// 分布式支持
	instanceID string       // 实例唯一标识
	broker     *KafkaBroker // Kafka 消息广播器
}

var ChatServer *Server

func init() {
	if ChatServer == nil {
		// 生成实例唯一ID
		instanceID := os.Getenv("INSTANCE_ID")
		if instanceID == "" {
			instanceID = uuid.New().String()[:8]
		}

		ChatServer = &Server{
			Clients:    make(map[string]*Client),
			mutex:      &sync.RWMutex{},
			Transmit:   make(chan []byte, constants.CHANNEL_SIZE),
			Login:      make(chan *Client, constants.CHANNEL_SIZE),
			Logout:     make(chan *Client, constants.CHANNEL_SIZE),
			instanceID: instanceID,
		}

		// 初始化 Kafka 消息广播器
		chatKafka := kafka.InitChatKafka()
		if chatKafka != nil && chatKafka.ChatWriter != nil && chatKafka.ChatReader != nil {
			ChatServer.broker = NewKafkaBroker(
				instanceID,
				&ChatServer.Clients,
				ChatServer.mutex,
				chatKafka.ChatWriter,
				chatKafka.ChatReader,
				kafka.GetChatKafka().ChatReader.Config().Topic,
			)
			if err := ChatServer.broker.Start(); err != nil {
				zlog.GetLogger().Warn("Kafka Broker 启动失败，降级为单实例模式", zap.Error(err))
			} else {
				zlog.GetLogger().Info("分布式模式已启用（Kafka）", zap.String("instance_id", instanceID))
			}
		} else {
			zlog.GetLogger().Warn("Kafka 未初始化，运行在单实例模式")
		}

		go ChatServer.Start()
	}
}

func (s *Server) Start() {
	for {
		select {
		case client, ok := <-s.Login:
			if !ok {
				return
			}
			s.handleLogin(client)
		case client, ok := <-s.Logout:
			if !ok {
				return
			}
			s.handleLogout(client)

		case data, ok := <-s.Transmit:
			if !ok {
				return
			}
			s.ConsumeMessage(data)
		}
	}
}

func (s *Server) handleLogin(client *Client) {
	s.mutex.Lock()
	// 若同一 UUID 已有旧连接，先关闭旧连接避免 goroutine 泄漏
	// 发送 4001 关闭码，告知旧客户端是被新设备挤下线，不要自动重连
	if old, exists := s.Clients[client.Uuid]; exists && old != client {
		// 使用 SendBack channel 发送关闭消息，避免并发写
		closeMsg := MessageBack{Message: websocket.FormatCloseMessage(4001, "another device logged in")}
		select {
		case old.SendBack <- closeMsg:
		default:
		}
		go func() {
			time.Sleep(100 * time.Millisecond) // 给消息发送时间
			old.Conn.Close()
		}()
	}
	s.Clients[client.Uuid] = client
	s.mutex.Unlock()

	// 注册用户在线状态到 Redis
	if s.broker != nil {
		if err := myredis.SetUserOnline(client.Uuid, s.instanceID); err != nil {
			zlog.GetLogger().Error("注册用户在线状态失败", zap.String("uuid", client.Uuid), zap.Error(err))
		}
	}

	zlog.GetLogger().Info("用户登录", zap.String("uuid", client.Uuid), zap.String("instance_id", s.instanceID))
	// 通过 SendBack channel 发送，由 Write() goroutine 统一写 conn，避免并发写
	loginMsg := MessageBack{Message: []byte("登录成功,欢迎来到GoLinko"), Uuid: ""}
	select {
	case client.SendBack <- loginMsg:
	default:
		zlog.GetLogger().Error("发送登录成功消息失败", zap.String("uuid", client.Uuid))
	}
}

func (s *Server) handleLogout(client *Client) {
	s.mutex.Lock()
	// 只有当 map 中存储的仍是这个 client 时才删除
	// 避免旧连接关闭触发的 Logout 把新连接从 map 中误删
	if current, exists := s.Clients[client.Uuid]; exists && current == client {
		delete(s.Clients, client.Uuid)
	}
	s.mutex.Unlock()

	// 移除用户在线状态
	if s.broker != nil {
		if err := myredis.RemoveUserOnline(client.Uuid); err != nil {
			zlog.GetLogger().Error("移除用户在线状态失败", zap.String("uuid", client.Uuid), zap.Error(err))
		}
	}

	zlog.GetLogger().Info("用户退出登录", zap.String("uuid", client.Uuid), zap.String("instance_id", s.instanceID))
}

// ConsumeMessage 统一处理来自不同消息通道（channel/kafka）的聊天消息。
func (s *Server) ConsumeMessage(data []byte) {
	var msg request.ChatMessageRequest
	if err := json.Unmarshal(data, &msg); err != nil {
		zlog.GetLogger().Error("反序列化消息失败")
		return
	}
	// 先查群聊表确定接收者类型，查到则为群聊，否则视为普通用户
	db := dao.GetDB()
	var isUser bool
	var group model.GroupInfo
	if err := db.Where("uuid=?", msg.ReceiveId).First(&group).Error; err == nil {
		// 群已解散/禁用时，不允许继续发群消息
		if group.Status == group_status_enum.DISABLE {
			zlog.GetLogger().Warn("群已解散，拒绝发送群消息", zap.String("group_id", msg.ReceiveId), zap.String("send_id", msg.SendId))
			return
		}
		isUser = false
	} else {
		isUser = true
	}

	//  持久化消息（Text/Voice/File 入库，AudioOrVideo 为信令不入库）
	msgUuid := utils.GenerateMessageID()
	if msg.Type == message_type_enum.Text || msg.Type == message_type_enum.Voice || msg.Type == message_type_enum.File {
		var fileSize int64
		if msg.FileSize != "" {
			fileSize, _ = strconv.ParseInt(msg.FileSize, 10, 64)
		}
		dbMsg := model.Message{
			Uuid:       msgUuid,
			SessionId:  msg.SessionId,
			Type:       msg.Type,
			Content:    msg.Content,
			Url:        msg.Url,
			SendId:     msg.SendId,
			SendName:   msg.SendName,
			SendAvatar: msg.SendAvatar,
			ReceiveId:  msg.ReceiveId,
			FileType:   msg.FileType,
			FileName:   msg.FileName,
			FileSize:   fileSize,
			AVdata:     msg.AVdata,
			Status:     message_status_enum.Unsent,
		}
		if err := db.Create(&dbMsg).Error; err != nil {
			zlog.GetLogger().Error("消息持久化失败", zap.Error(err))
			return
		}
		// 清理双方的消息列表缓存，防止刷新后新消息消失
		_ = myredis.DelKeyWithPatternIfExist("message_list_" + msg.SendId + "_" + msg.SessionId)
		_ = myredis.DelKeyWithPatternIfExist("message_list_" + msg.ReceiveId + "_" + msg.SessionId)
		_ = myredis.DelKeyWithPatternIfExist("group_message_list_" + msg.ReceiveId)
	}

	messageBack := MessageBack{
		Message: data,
		Uuid:    msgUuid,
	}

	// 消息推送（分布式支持：通过 Kafka 跨实例推送）
	if msg.Type == message_type_enum.Text || msg.Type == message_type_enum.Voice || msg.Type == message_type_enum.File {
		if isUser {
			// 私聊消息
			if s.broker != nil {
				// 分布式模式：通过 Kafka 广播
				if msg.ReceiveId != msg.SendId {
					s.broker.Publish(msg.ReceiveId, data, msgUuid)
				}
				s.broker.Publish(msg.SendId, data, msgUuid)
			} else {
				// 单实例模式：直接推送
				s.mutex.RLock()
				receiveClient := s.Clients[msg.ReceiveId]
				sendClient := s.Clients[msg.SendId]
				s.mutex.RUnlock()
				if receiveClient != nil && msg.ReceiveId != msg.SendId {
					receiveClient.SendBack <- messageBack
				}
				if sendClient != nil {
					sendClient.SendBack <- messageBack
				}
			}
		} else {
			// 群聊消息
			var members []string
			if err := json.Unmarshal(group.Members, &members); err != nil {
				zlog.GetLogger().Error("反序列化群成员失败", zap.String("receive_id", msg.ReceiveId), zap.Error(err))
				return
			}
			// 发送者必须仍在群内
			inGroup := false
			for _, memberId := range members {
				if memberId == msg.SendId {
					inGroup = true
					break
				}
			}
			if !inGroup {
				zlog.GetLogger().Warn("发送者不在群内，拒绝发送群消息", zap.String("group_id", msg.ReceiveId), zap.String("send_id", msg.SendId))
				return
			}

			if s.broker != nil {
				// 分布式模式：通过 Kafka 广播给所有群成员
				s.broker.PublishToUsers(members, data, msgUuid)
			} else {
				// 单实例模式：直接推送
				s.mutex.RLock()
				onlineClients := make([]*Client, 0, len(members))
				for _, memberId := range members {
					if c, ok := s.Clients[memberId]; ok {
						onlineClients = append(onlineClients, c)
					}
				}
				s.mutex.RUnlock()
				for _, c := range onlineClients {
					c.SendBack <- messageBack
				}
			}
		}
	} else if msg.Type == message_type_enum.AudioOrVideo {
		// 音视频信令仅转发给接收方，不入库
		signalingBack := MessageBack{
			Message: data,
			Uuid:    "", // 不入库，UUID留空
		}
		if isUser {
			if s.broker != nil {
				// 分布式模式
				s.broker.Publish(msg.ReceiveId, data, "")
			} else {
				// 单实例模式
				s.mutex.RLock()
				receiveClient := s.Clients[msg.ReceiveId]
				s.mutex.RUnlock()
				if receiveClient != nil {
					receiveClient.SendBack <- signalingBack
				}
			}
		}
	}
}

// 将https://127.0.0.1:8000/static/xxx 转为 /static/xxx
func normalizePath(path string) string {
	if path == "https://cube.elemecdn.com/0/88/03b0d39583f48206768a7534e55bcpng.png" {
		return path
	}
	staticIndex := strings.Index(path, "/static/")
	if staticIndex < 0 {
		log.Println(path)
		zlog.GetLogger().Error("路径不合法")
		return path // 防止越界 panic，原路返回
	}
	// 返回从 "/static/" 开始的部分
	return path[staticIndex:]
}

func (s *Server) Close() {
	if s.broker != nil {
		s.broker.Stop()
	}
	close(s.Login)
	close(s.Logout)
	close(s.Transmit)
}

func (s *Server) SendClientToLogin(client *Client) {
	s.Login <- client
}

func (s *Server) SendClientToLogout(client *Client) {
	s.Logout <- client
}

func (s *Server) SendMessageToTransmit(message []byte) {
	s.Transmit <- message
}

func (s *Server) RemoveClient(uuid string) {
	s.mutex.Lock()
	delete(s.Clients, uuid)
	s.mutex.Unlock()
}

// PushToUser 向在线用户直接推送一条消息（不走持久化流程）。
// 支持分布式模式：如果用户不在本实例，通过 Kafka 转发
func (s *Server) PushToUser(userID string, messageBack MessageBack) bool {
	// 分布式模式：通过 Kafka 广播
	if s.broker != nil {
		return s.broker.Publish(userID, messageBack.Message, messageBack.Uuid) == nil
	}

	// 单实例模式：直接推送
	s.mutex.RLock()
	client := s.Clients[userID]
	s.mutex.RUnlock()
	if client == nil {
		return false
	}
	select {
	case client.SendBack <- messageBack:
		return true
	default:
		zlog.GetLogger().Warn("用户发送队列已满，放弃推送", zap.String("user_id", userID))
		return false
	}
}

// GetInstanceID 获取当前实例ID
func (s *Server) GetInstanceID() string {
	return s.instanceID
}

// IsDistributed 检查是否启用分布式模式
func (s *Server) IsDistributed() bool {
	return s.broker != nil
}

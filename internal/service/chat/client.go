package chat

import (
	"GoLinko/internal/config"
	"GoLinko/internal/dao"
	"GoLinko/internal/dto/request"
	"GoLinko/internal/model"

	constants "GoLinko/pkg/const"
	"GoLinko/pkg/enum/message/message_status_enum"
	"GoLinko/pkg/enum/message/message_type_enum"
	"GoLinko/pkg/zlog"
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type MessageBack struct {
	Message []byte `json:"message"`
	Uuid    string `json:"uuid"`
}

type Client struct {
	Conn     *websocket.Conn
	Uuid     string
	SendTo   chan []byte      //给server
	SendBack chan MessageBack //给前端
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  2048,
	WriteBufferSize: 2048,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var ctx = context.Background()

func (c *Client) Read() {
	zlog.GetLogger().Info("开始读取消息", zap.String("uuid", c.Uuid))
	defer func() {
		zlog.GetLogger().Info("WebSocket连接断开，正在清理资源", zap.String("uuid", c.Uuid))
		ChatServer.SendClientToLogout(c)
		c.Conn.Close()
	}()
	for {
		_, jsonMessage, err := c.Conn.ReadMessage()
		if err != nil {
			zlog.GetLogger().Error("读取WS消息失败", zap.Error(err))
			return //断开Websocket
		} else {
			message := request.ChatMessageRequest{}
			if err := json.Unmarshal(jsonMessage, &message); err != nil {
				zlog.GetLogger().Error("反序列化消息失败")
				continue
			}
			// 安全：强制使用 JWT 鉴权后的用户ID覆盖客户端传入的 SendId，防止伪造
			message.SendId = c.Uuid
			modifiedMessage, err := json.Marshal(message)
			if err != nil {
				zlog.GetLogger().Error("重新序列化消息失败", zap.Error(err))
				continue
			}
			messageMode := strings.ToLower(strings.TrimSpace(config.GetConfig().KafkaConfig.MessageMode))
			// 音视频信令对时延敏感，始终走进程内通道，避免 Kafka 批处理带来的额外延迟。
			if message.Type == message_type_enum.AudioOrVideo {
				if len(ChatServer.Transmit) < constants.CHANNEL_SIZE {
					ChatServer.SendMessageToTransmit(modifiedMessage)
				} else if len(c.SendTo) < constants.CHANNEL_SIZE {
					c.SendTo <- modifiedMessage
				} else {
					if err := c.Conn.WriteMessage(websocket.TextMessage, []byte("服务器繁忙，请稍后再发送")); err != nil {
						zlog.GetLogger().Error(err.Error())
					}
				}
				continue
			}
			if messageMode == "kafka" {
				if err := PublishMessageToKafka(message.SessionId, modifiedMessage); err != nil {
					zlog.GetLogger().Error("发送 Kafka 消息失败", zap.Error(err))
					if err := c.Conn.WriteMessage(websocket.TextMessage, []byte("消息发送失败，请稍后重试")); err != nil {
						zlog.GetLogger().Error(err.Error())
					}
				}
				continue
			}
			if messageMode == "channel" || messageMode == "" {
				//如果server的转发channel没满，先把sendto给transmit
				for len(ChatServer.Transmit) < constants.CHANNEL_SIZE && len(c.SendTo) > 0 {
					sendToMessage := <-c.SendTo
					ChatServer.SendMessageToTransmit(sendToMessage)
				}
				//如果server没满，sendto空了，直接给server的transmit
				if len(ChatServer.Transmit) < constants.CHANNEL_SIZE {
					ChatServer.SendMessageToTransmit(modifiedMessage)
				} else if len(c.SendTo) < constants.CHANNEL_SIZE {
					c.SendTo <- modifiedMessage
				} else {
					if err := c.Conn.WriteMessage(websocket.TextMessage, []byte("服务器繁忙，请稍后再发送")); err != nil {
						zlog.GetLogger().Error(err.Error())
					}
				}
			} else {
				zlog.GetLogger().Warn("未知 message_mode，回退为 channel", zap.String("message_mode", messageMode))
				if len(ChatServer.Transmit) < constants.CHANNEL_SIZE {
					ChatServer.SendMessageToTransmit(modifiedMessage)
				}
			}
		}

	}
}

func (c *Client) Write() {
	zlog.GetLogger().Info("ws Write goroutine Start", zap.String("uuid", c.Uuid))
	defer func() {
		zlog.GetLogger().Info("ws Write goroutine Stop", zap.String("uuid", c.Uuid))
		c.Conn.Close()
	}()
	for messageBack := range c.SendBack {
		err := c.Conn.WriteMessage(websocket.TextMessage, messageBack.Message)
		if err != nil {
			zlog.GetLogger().Error("向客户端发送消息失败", zap.String("uuid", c.Uuid), zap.Error(err))
			return //直接断开ws
		}
		// 仅当消息已入库（有UUID）时才更新发送状态，音视频信令不入库故跳过
		if messageBack.Uuid != "" {
			db := dao.GetDB()
			if err := db.Model(&model.Message{}).Where("uuid=?", messageBack.Uuid).Update("status", message_status_enum.Sent).Error; err != nil {
				zlog.GetLogger().Error("更新消息状态失败", zap.String("msg_uuid", messageBack.Uuid), zap.Error(err))
			}
		}
	}
}

func NewClientInit(c *gin.Context, clientId string) {

	coon, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		zlog.GetLogger().Error("升级WebSocket连接失败", zap.Error(err))
		return
	}
	client := &Client{
		Conn:     coon,
		Uuid:     clientId,
		SendTo:   make(chan []byte, constants.CHANNEL_SIZE),
		SendBack: make(chan MessageBack, constants.CHANNEL_SIZE),
	}
	// 将客户端发送到 Login 通道
	ChatServer.SendClientToLogin(client)

	go client.Read()
	go client.Write()
	zlog.GetLogger().Info("WebSocket连接已建立", zap.String("clientId", clientId))
}

func WsLogout(userId string) (string, int) {

	//kafkaConfig := config.GetConfig().KafkaConfig
	client := ChatServer.Clients[userId]
	if client != nil {
		if err := client.Conn.Close(); err != nil {
			zlog.GetLogger().Error("关闭WebSocket连接失败", zap.Error(err))
			return constants.SYSTEM_ERROR, -1
		}
	}
	return "退出成功", 0
}

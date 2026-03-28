package chat

import (
	"GoLinko/pkg/zlog"
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// WSMessagePayload WebSocket 跨实例消息载体
type WSMessagePayload struct {
	TargetUserID string `json:"target_user_id"` // 目标用户ID
	Data         []byte `json:"data"`           // 原始消息数据
	MessageID    string `json:"message_id"`     // 消息UUID
	Timestamp    int64  `json:"timestamp"`      // 时间戳
}

// KafkaBroker Kafka 消息广播器
// 用于实现 WebSocket 集群化，支持跨实例消息推送
type KafkaBroker struct {
	instanceID   string              // 当前实例ID
	writer       *kafka.Writer       // Kafka 生产者
	reader       *kafka.Reader       // Kafka 消费者
	localClients *map[string]*Client // 引用 Server.Clients（只读）
	clientsMutex *sync.RWMutex       // 引用 Server.mutex
	stopChan     chan struct{}       // 停止信号
	topic        string              // Kafka Topic
	wg           sync.WaitGroup      // 等待组
}

// NewKafkaBroker 创建 Kafka 消息广播器
func NewKafkaBroker(instanceID string, clients *map[string]*Client, mutex *sync.RWMutex, writer *kafka.Writer, reader *kafka.Reader, topic string) *KafkaBroker {
	return &KafkaBroker{
		instanceID:   instanceID,
		writer:       writer,
		reader:       reader,
		localClients: clients,
		clientsMutex: mutex,
		stopChan:     make(chan struct{}),
		topic:        topic,
	}
}

// Start 启动 Kafka 消费者
func (b *KafkaBroker) Start() error {
	if b.reader == nil {
		zlog.GetLogger().Warn("Kafka Reader 未初始化，跨实例消息功能不可用")
		return nil
	}

	zlog.GetLogger().Info("Kafka Broker 启动",
		zap.String("instance_id", b.instanceID),
		zap.String("topic", b.topic))

	// 启动消息消费 goroutine
	b.wg.Add(1)
	go b.consumeMessages()

	return nil
}

// consumeMessages 消费 Kafka 消息
func (b *KafkaBroker) consumeMessages() {
	defer b.wg.Done()

	for {
		select {
		case <-b.stopChan:
			zlog.GetLogger().Info("Kafka Broker 停止", zap.String("instance_id", b.instanceID))
			return
		default:
			// 读取消息
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			msg, err := b.reader.ReadMessage(ctx)
			cancel()

			if err != nil {
				if err == context.DeadlineExceeded {
					continue // 超时是正常的，继续检查停止信号
				}
				if err == context.Canceled {
					return
				}
				// 仅在非超时错误时记录日志
				zlog.GetLogger().Debug("读取 Kafka 消息", zap.Error(err))
				continue
			}

			b.processMessage(msg)
		}
	}
}

// processMessage 处理单条 Kafka 消息
func (b *KafkaBroker) processMessage(msg kafka.Message) {
	var payload WSMessagePayload
	if err := json.Unmarshal(msg.Value, &payload); err != nil {
		zlog.GetLogger().Error("反序列化 Kafka 消息失败", zap.Error(err))
		return
	}

	// 检查目标用户是否在本实例
	b.clientsMutex.RLock()
	client, exists := (*b.localClients)[payload.TargetUserID]
	b.clientsMutex.RUnlock()

	if exists && client != nil {
		// 用户在本实例，推送消息
		messageBack := MessageBack{
			Message: payload.Data,
			Uuid:    payload.MessageID,
		}

		select {
		case client.SendBack <- messageBack:
			zlog.GetLogger().Debug("Kafka 跨实例消息推送成功",
				zap.String("target_user", payload.TargetUserID),
				zap.String("instance_id", b.instanceID))
		default:
			zlog.GetLogger().Warn("用户发送队列已满，放弃推送",
				zap.String("user_id", payload.TargetUserID))
		}
	}
}

// Publish 发布消息到 Kafka（供其他实例消费）
func (b *KafkaBroker) Publish(userID string, data []byte, messageID string) error {
	if b.writer == nil {
		zlog.GetLogger().Warn("Kafka Writer 未初始化，无法发布跨实例消息")
		return nil
	}

	payload := WSMessagePayload{
		TargetUserID: userID,
		Data:         data,
		MessageID:    messageID,
		Timestamp:    time.Now().UnixMilli(),
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	msg := kafka.Message{
		Key:   []byte(userID), // 按用户ID分区，保证同一用户消息顺序
		Value: payloadBytes,
	}

	if err := b.writer.WriteMessages(ctx, msg); err != nil {
		zlog.GetLogger().Error("发布消息到 Kafka 失败", zap.String("user_id", userID), zap.Error(err))
		return err
	}

	return nil
}

// PublishToUsers 批量发布消息给多个用户
func (b *KafkaBroker) PublishToUsers(userIDs []string, data []byte, messageID string) error {
	if b.writer == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	messages := make([]kafka.Message, 0, len(userIDs))
	for _, userID := range userIDs {
		payload := WSMessagePayload{
			TargetUserID: userID,
			Data:         data,
			MessageID:    messageID,
			Timestamp:    time.Now().UnixMilli(),
		}

		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			zlog.GetLogger().Error("序列化消息失败", zap.String("user_id", userID), zap.Error(err))
			continue
		}

		messages = append(messages, kafka.Message{
			Key:   []byte(userID),
			Value: payloadBytes,
		})
	}

	if len(messages) == 0 {
		return nil
	}

	if err := b.writer.WriteMessages(ctx, messages...); err != nil {
		zlog.GetLogger().Error("批量发布消息到 Kafka 失败", zap.Error(err))
		return err
	}

	return nil
}

// Stop 停止 Kafka Broker
func (b *KafkaBroker) Stop() {
	close(b.stopChan)
	b.wg.Wait()
}

// GetInstanceID 获取当前实例ID
func (b *KafkaBroker) GetInstanceID() string {
	return b.instanceID
}

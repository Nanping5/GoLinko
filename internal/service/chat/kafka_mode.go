package chat

import (
	"GoLinko/internal/config"
	kafkasvc "GoLinko/internal/service/kafka"
	"GoLinko/pkg/zlog"
	"context"
	"errors"
	"strings"

	"go.uber.org/zap"
)

// StartKafkaPipeline 在 kafka 模式下初始化并启动聊天消息消费者。
func StartKafkaPipeline() {
	mode := strings.ToLower(strings.TrimSpace(config.GetConfig().KafkaConfig.MessageMode))
	zlog.GetLogger().Info("消息通道模式已生效", zap.String("message_mode", mode))
	if mode != "kafka" {
		zlog.GetLogger().Info("Kafka 未启用，当前使用 channel 模式")
		return
	}
	if err := kafkasvc.ProbeKafkaConnection(context.Background()); err != nil {
		zlog.GetLogger().Error("Kafka 连接状态: 失败", zap.Error(err))
		return
	}
	zlog.GetLogger().Info("Kafka 连接状态: 正常")
	service := kafkasvc.InitChatKafka()
	if service == nil {
		zlog.GetLogger().Error("Kafka 初始化失败，无法启动消费者")
		return
	}
	go service.StartChatConsumer(context.Background(), ChatServer.ConsumeMessage)
	zlog.GetLogger().Info("Kafka 聊天消费者已启动")
}

func PublishMessageToKafka(sessionID string, payload []byte) error {
	service := kafkasvc.GetChatKafka()
	if service == nil {
		service = kafkasvc.InitChatKafka()
	}
	if service == nil {
		return errors.New("kafka service not initialized")
	}
	return service.PublishChatMessage(context.Background(), sessionID, payload)
}

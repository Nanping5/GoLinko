package kafka

import (
	"GoLinko/internal/config"
	"GoLinko/pkg/zlog"
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type KafkaService struct {
	ChatWriter *kafka.Writer
	ChatReader *kafka.Reader
	KafkaConn  *kafka.Conn
}

var (
	chatKafkaService *KafkaService
	kafkaOnce        sync.Once
)

func InitChatKafka() *KafkaService {
	kafkaOnce.Do(func() {
		svc := &KafkaService{}
		svc.KafkaInit()
		svc.CreateTopic()
		chatKafkaService = svc
	})
	return chatKafkaService
}

func GetChatKafka() *KafkaService {
	return chatKafkaService
}

// ProbeKafkaConnection 探测 Kafka broker 是否可连接。
func ProbeKafkaConnection(ctx context.Context) error {
	kafkaConfig := config.GetConfig().KafkaConfig
	dialer := &kafka.Dialer{Timeout: time.Duration(kafkaConfig.TimeOut) * time.Second}
	conn, err := dialer.DialContext(ctx, "tcp", kafkaConfig.HostPort)
	if err != nil {
		return fmt.Errorf("dial kafka %s failed: %w", kafkaConfig.HostPort, err)
	}
	_ = conn.Close()
	return nil
}

func (k *KafkaService) KafkaInit() {
	kafkaConfig := config.GetConfig().KafkaConfig

	k.ChatWriter = &kafka.Writer{
		Addr:                   kafka.TCP(kafkaConfig.HostPort),
		Topic:                  kafkaConfig.ChatTopic,
		Balancer:               &kafka.Hash{},
		WriteTimeout:           time.Duration(kafkaConfig.TimeOut) * time.Second,
		BatchTimeout:           10 * time.Millisecond,
		RequiredAcks:           kafka.RequireOne, //等待至少一个分区副本确认消息已写入
		AllowAutoTopicCreation: true,             //允许自动创建
	}
	k.ChatReader = kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{kafkaConfig.HostPort},
		GroupID:        "GoLinkoGroup",
		Topic:          kafkaConfig.ChatTopic,
		StartOffset:    kafka.LastOffset,
		CommitInterval: time.Duration(kafkaConfig.TimeOut) * time.Second,
	})
}

// KafkaClose 关闭 Kafka 连接和资源
func (k *KafkaService) KafkaClose() {
	if k.ChatWriter != nil {
		if err := k.ChatWriter.Close(); err != nil {
			zlog.GetLogger().Error("关闭 Kafka Writer 失败", zap.Error(err))
		}
	}
	if k.ChatReader != nil {
		if err := k.ChatReader.Close(); err != nil {
			zlog.GetLogger().Error("关闭 Kafka Reader 失败", zap.Error(err))
		}
	}
	if k.KafkaConn != nil {
		if err := k.KafkaConn.Close(); err != nil {
			zlog.GetLogger().Error("关闭 Kafka Conn 失败", zap.Error(err))
		}
	}
	chatKafkaService = nil
}

func (k *KafkaService) PublishChatMessage(ctx context.Context, key string, payload []byte) error {
	if k == nil || k.ChatWriter == nil {
		return errors.New("kafka writer not initialized")
	}
	msg := kafka.Message{Value: payload}
	if key != "" {
		msg.Key = []byte(key)
	}
	if err := k.ChatWriter.WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("write kafka message failed: %w", err)
	}
	return nil
}

func (k *KafkaService) StartChatConsumer(ctx context.Context, handler func([]byte)) {
	if k == nil || k.ChatReader == nil {
		zlog.GetLogger().Error("Kafka Reader 未初始化，无法启动消费者")
		return
	}
	for {
		msg, err := k.ChatReader.ReadMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			zlog.GetLogger().Error("读取 Kafka 消息失败", zap.Error(err))
			continue
		}
		handler(msg.Value)
	}
}

// CreateTopic 创建 Kafka Topic
func (k *KafkaService) CreateTopic() {
	kafkaConfig := config.GetConfig().KafkaConfig
	chatTopic := kafkaConfig.ChatTopic

	var err error
	k.KafkaConn, err = kafka.Dial("tcp", kafkaConfig.HostPort)
	if err != nil {
		zlog.GetLogger().Error("连接 Kafka 失败", zap.Error(err))
		return
	}
	topicConfigs := []kafka.TopicConfig{
		{
			Topic:             chatTopic,
			NumPartitions:     kafkaConfig.Partition,
			ReplicationFactor: 1,
		},
	}

	//创建topic
	if err := k.KafkaConn.CreateTopics(topicConfigs...); err != nil {
		zlog.GetLogger().Error("创建 Kafka topic失败", zap.Error(err))
	}
}

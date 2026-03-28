package kafka

import "context"

// MessageBroker 统一封装不同消息通道（channel/kafka）的发送能力。
type MessageBroker interface {
	PublishChatMessage(ctx context.Context, key string, payload []byte) error
}

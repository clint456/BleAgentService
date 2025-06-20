package interfaces

import "github.com/edgexfoundry/go-mod-messaging/v4/pkg/types"

// MessageBusClient 定义消息总线的通用接口，供 dataparse、mqttbus 等包依赖
// 实现应包含 Publish、Subscribe 等方法

type MessageBusClient interface {
	Publish(topic string, payload []byte) error
	Subscribe(topics []string, handler func(topic string, envelope types.MessageEnvelope) error) error
}

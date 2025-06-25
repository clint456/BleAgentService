package interfaces

import (
	"time"

	"github.com/edgexfoundry/go-mod-messaging/v4/pkg/types"
)

// MessageBusClient 定义消息总线的通用接口，供 dataparse、mqttbus 等包依赖
// 实现应包含 Publish、Subscribe 等方法

type MessageBusClient interface {
	Publish(topic string, data interface{}) error
	Subscribe(topic1 string, handler func(topic2 string, envelope types.MessageEnvelope) error) error
	Request(topic string, data interface{}) (types.MessageEnvelope, error)
	SubscribeResponse(topic string) error
	SetTimeout(timeout time.Duration)
}

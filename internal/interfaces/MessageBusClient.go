package interfaces

import (
	"time"

	"github.com/edgexfoundry/go-mod-messaging/v4/pkg/types"
)

// MessageBusClient 定义消息总线的通用接口。
// 该接口提供了消息的发布、订阅、请求和响应的功能。
type MessageBusClient interface {
	// Publish 发布消息到指定的主题。
	//
	// 参数:
	//   - topic: 消息的主题。
	//   - data: 要发送的数据，可以是任何类型的对象。
	//
	// 返回值:
	//   - error: 如果发布消息过程中出现错误，返回非 nil 错误。
	Publish(topic string, data interface{}) error

	// Subscribe 订阅指定主题的消息。
	// 当接收到消息时，将调用 handler 函数进行处理。
	//
	// 参数:
	//   - topic1: 要订阅的主题。
	//   - handler: 当收到消息时调用的处理函数，参数为接收到的主题和消息内容。
	//
	// 返回值:
	//   - error: 如果订阅过程中出现错误，返回非 nil 错误。
	Subscribe(topic1 string, handler func(topic2 string, envelope types.MessageEnvelope) error) error

	// Request 发送请求消息到指定主题并等待响应。
	//
	// 参数:
	//   - topic: 请求的主题。
	//   - data: 要发送的数据。
	//
	// 返回值:
	//   - types.MessageEnvelope: 响应的消息封装。
	//   - error: 如果请求过程中出现错误，返回非 nil 错误。
	Request(topic string, data interface{}) (types.MessageEnvelope, error)

	// SubscribeResponse 订阅响应消息的主题。
	// 此方法用于响应消息的订阅，通常与 Request 方法配合使用。
	//
	// 参数:
	//   - topic: 响应的主题。
	//
	// 返回值:
	//   - error: 如果订阅响应过程出现错误，返回非 nil 错误。
	SubscribeResponse(topic string) error

	// SetTimeout 设置请求和响应的超时时间。
	//
	// 参数:
	//   - timeout: 设置的超时时间。
	//
	// 返回值:
	//   - none
	SetTimeout(timeout time.Duration)

	// Disconnect 断开与消息总线的连接。
	//
	// 返回值:
	//   - error: 如果断开连接过程中出现错误，返回非 nil 错误。
	Disconnect() error
}

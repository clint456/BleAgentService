package mqttbus

import (
	"fmt"
	"strings"

	messagebus "github.com/clint456/edgex-messagebus-client"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"github.com/edgexfoundry/go-mod-messaging/v4/pkg/types"
)

// 消息总线接口
type MessageBusClient interface {
	Connect() error
	Disconnect() error
	IsConnected() bool
	Publish(topic string, data interface{}) error
	Subscribe(topics []string, handler func(topic string, envelope types.MessageEnvelope) error) error
}

// EdgexMessageBusClient 实现
type EdgexMessageBusClient struct {
	client *messagebus.Client
}

// handler 通过参数传递，logger 也通过参数传递
func NewEdgexMessageBusClient(cfg map[string]interface{}, logger logger.LoggingClient, subscribeTopics []string, handler func(topic string, envelope types.MessageEnvelope) error) (*EdgexMessageBusClient, error) {
	config := messagebus.Config{
		Host:     cfg["Host"].(string),
		Port:     cfg["Port"].(int),
		Protocol: strings.ToLower(cfg["Protocol"].(string)),
		Type:     "mqtt",
		ClientID: cfg["ClientID"].(string),
		QoS:      cfg["QoS"].(int),
		Username: cfg["Username"].(string),
		Password: cfg["Password"].(string),
	}
	client, err := messagebus.NewClient(config, logger)
	if err != nil {
		return nil, fmt.Errorf("创建MessageBus客户端失败: %w", err)
	}
	if err := client.Connect(); err != nil {
		return nil, fmt.Errorf("连接MessageBus失败: %w", err)
	}
	// 包装 handler 以适配 messagebus.Client.Subscribe 的签名
	wrappedHandler := func(topic string, envelope types.MessageEnvelope) error {
		return handler(topic, envelope)
	}
	if err := client.Subscribe(subscribeTopics, wrappedHandler); err != nil {
		client.Disconnect()
		return nil, fmt.Errorf("订阅主题失败: %w", err)
	}
	return &EdgexMessageBusClient{client: client}, nil
}

func (e *EdgexMessageBusClient) Connect() error    { return e.client.Connect() }
func (e *EdgexMessageBusClient) Disconnect() error { return e.client.Disconnect() }
func (e *EdgexMessageBusClient) IsConnected() bool { return e.client.IsConnected() }
func (e *EdgexMessageBusClient) Publish(topic string, data interface{}) error {
	return e.client.Publish(topic, data)
}
func (e *EdgexMessageBusClient) Subscribe(topics []string, handler func(topic string, envelope types.MessageEnvelope) error) error {
	// 包装 handler 以适配 messagebus.Client.Subscribe 的签名
	wrappedHandler := func(topic string, envelope types.MessageEnvelope) error {
		return handler(topic, envelope)
	}
	return e.client.Subscribe(topics, wrappedHandler)
}

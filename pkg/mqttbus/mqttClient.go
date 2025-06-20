package mqttbus

import (
	"fmt"
	"strings"

	messagebus "github.com/clint456/edgex-messagebus-client"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"github.com/edgexfoundry/go-mod-messaging/v4/pkg/types"
)

// EdgexMessageBusClient 实现 interfaces.MessageBusClient
// handler 通过参数传递，logger 也通过参数传递

type EdgexMessageBusClient struct {
	client *messagebus.Client
}

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

func (e *EdgexMessageBusClient) Publish(topic string, payload []byte) error {
	return e.client.Publish(topic, payload)
}

func (e *EdgexMessageBusClient) Subscribe(topic string, handler func([]byte)) error {
	// 这里需要适配 handler 签名，具体实现可根据实际 messagebus 客户端调整
	return nil // TODO: 实现订阅逻辑
}

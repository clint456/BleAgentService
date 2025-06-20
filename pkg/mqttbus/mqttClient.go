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

// NewEdgexMessageBusClient 只负责初始化和连接，不注册 handler
func NewEdgexMessageBusClient(cfg map[string]interface{}, logger logger.LoggingClient) (*EdgexMessageBusClient, error) {
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
	return &EdgexMessageBusClient{client: client}, nil
}

// Subscribe 注册 handler
func (e *EdgexMessageBusClient) Subscribe(topics []string, handler func(topic string, envelope types.MessageEnvelope) error) error {
	wrappedHandler := func(topic string, envelope types.MessageEnvelope) error {
		return handler(topic, envelope)
	}
	if err := e.client.Subscribe(topics, wrappedHandler); err != nil {
		return fmt.Errorf("订阅主题失败: %w", err)
	}
	return nil
}

func (e *EdgexMessageBusClient) Publish(topic string, payload []byte) error {
	return e.client.Publish(topic, payload)
}

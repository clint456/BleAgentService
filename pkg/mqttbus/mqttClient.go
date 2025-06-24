package mqttbus

import (
	internalif "device-ble/internal/interfaces"
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
func NewEdgexMessageBusClient(cfg internalif.MQTTConfig, logger logger.LoggingClient) (*EdgexMessageBusClient, error) {
	config := messagebus.Config{
		Host:     cfg.Host,
		Port:     cfg.Port,
		Protocol: cfg.Protocol,
		Type:     "mqtt",
		ClientID: cfg.ClientID,
		QoS:      cfg.QoS,
		Username: cfg.Username,
		Password: cfg.Password,
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
		// 加日志确认是否被调用
		fmt.Printf("wrappedHandler被调用: topic=%s\n", topic)

		return handler(topic, envelope)
	}
	if err := e.client.Subscribe(topics, wrappedHandler); err != nil {
		return fmt.Errorf("订阅主题失败: %w", err)
	}
	return nil
}

func (e *EdgexMessageBusClient) Publish(topic string, data interface{}) error {
	transmitTopic := strings.Replace(topic, "edgex", "", -1)
	return e.client.Publish("edgex/service/data"+transmitTopic, data)
}

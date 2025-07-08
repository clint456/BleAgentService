package mqttbus

import (
	"device-ble/cmd/config"
	"device-ble/pkg/messagebus"
	"fmt"
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"github.com/edgexfoundry/go-mod-messaging/v4/pkg/types"
)

// EdgexMessageBusClient 实现 interfaces.MessageBusClient
// handler 通过参数传递，logger 也通过参数传递

type EdgexMessageBusClient struct {
	client *messagebus.Client
}

// NewEdgexMessageBusClient 只负责初始化和连接，不注册 handler
func NewEdgexMessageBusClient(cfg *config.MQTTUserClientConfig, logger logger.LoggingClient) (*EdgexMessageBusClient, error) {
	config := messagebus.Config{
		Host:     cfg.MqttUserConfig.Host,
		Port:     cfg.MqttUserConfig.Port,
		Protocol: cfg.MqttUserConfig.Protocol,
		Type:     "mqtt",
		ClientID: cfg.MqttUserConfig.ClientID,
		QoS:      cfg.MqttUserConfig.QoS,
		Username: cfg.MqttUserConfig.Username,
		Password: cfg.MqttUserConfig.Password,
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
func (e *EdgexMessageBusClient) Subscribe(topic1 string, handler func(topic2 string, envelope types.MessageEnvelope) error) error {
	// 装饰器：后期可以在不修改远程包的基础上自定义该函数
	wrappedHandler := func(topic string, envelope types.MessageEnvelope) error {
		// 加日志确认是否被调用
		fmt.Printf("wrappedHandler被调用: topic=%s\n", topic)
		return handler(topic, envelope)
	}
	if err := e.client.Subscribe(topic1, wrappedHandler); err != nil {
		return fmt.Errorf("订阅主题失败: %w", err)
	}
	return nil
}

func (e *EdgexMessageBusClient) Publish(topic string, data interface{}) error {
	return e.client.Publish(topic, data)
}

func (e *EdgexMessageBusClient) SubscribeResponse(topic string) error {
	return e.client.SubscribeResponse(topic)
}

func (e *EdgexMessageBusClient) Request(topic string, data interface{}) (types.MessageEnvelope, error) {
	return e.client.Request(topic, data)
}

func (e *EdgexMessageBusClient) SetTimeout(timeout time.Duration) {
	e.client.SetTimeout(timeout)
}

func (e *EdgexMessageBusClient) Disconnect() error {
	return e.client.Disconnect()
}

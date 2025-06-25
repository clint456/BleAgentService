// Package messagebus 提供简化版的 EdgeX MessageBus 客户端封装，支持请求-响应通信
package messagebus

import (
	"fmt"
	"sync"
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/common"
	"github.com/edgexfoundry/go-mod-messaging/v4/messaging"
	"github.com/edgexfoundry/go-mod-messaging/v4/pkg/types"
	"github.com/google/uuid"
)

// Client 表示一个简化版的 EdgeX MessageBus 客户端，支持基本的发布、订阅和基于 RequestID 的请求-响应机制
type Client struct {
	client        messaging.MessageClient    // 底层消息客户端
	lc            logger.LoggingClient       // 日志客户端
	isConnected   bool                       // 连接状态标志
	mutex         sync.RWMutex               // 并发读写锁
	messageChan   chan types.MessageEnvelope // 主动订阅消息通道
	errorChan     chan error                 // 错误信息通道
	stopChan      chan struct{}              // 停止所有处理器的控制通道
	wg            sync.WaitGroup             // 等待所有 goroutine 正常退出
	responseChMap sync.Map                   // RequestID -> 响应通道，用于匹配响应
	timeout       time.Duration              // 请求-响应超时时间
}

// Config 表示 MessageBus 配置参数
type Config struct {
	Host     string        // 主机地址
	Port     int           // 端口号
	Protocol string        // 协议（mqtt/nats）
	Type     string        // 消息总线类型
	ClientID string        // 客户端 ID
	Username string        // 用户名（可选）
	Password string        // 密码（可选）
	QoS      int           // QoS 级别（可选）
	Timeout  time.Duration // 默认请求超时时间（可选）
}

// MessageHandler 定义处理订阅消息的回调函数签名
type MessageHandler func(topic string, message types.MessageEnvelope) error

// NewClient 创建一个新的 MessageBus 客户端实例
func NewClient(config Config, lc logger.LoggingClient) (*Client, error) {
	messageBusConfig := types.MessageBusConfig{
		Broker: types.HostInfo{
			Host:     config.Host,
			Port:     config.Port,
			Protocol: config.Protocol,
		},
		Type: config.Type,
		Optional: map[string]string{
			"ClientId": config.ClientID,
		},
	}
	if config.Username != "" {
		messageBusConfig.Optional["Username"] = config.Username
	}
	if config.Password != "" {
		messageBusConfig.Optional["Password"] = config.Password
	}
	if config.QoS > 0 {
		messageBusConfig.Optional["Qos"] = fmt.Sprintf("%d", config.QoS)
	}

	client, err := messaging.NewMessageClient(messageBusConfig)
	if err != nil {
		return nil, err
	}

	return &Client{
		client:      client,
		lc:          lc,
		messageChan: make(chan types.MessageEnvelope, 100),
		errorChan:   make(chan error, 10),
		stopChan:    make(chan struct{}),
		timeout:     config.Timeout,
	}, nil
}

// SetTimeout 设置默认请求响应超时时间
func (c *Client) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
}

// Connect 建立与消息总线的连接
func (c *Client) Connect() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.isConnected {
		return nil
	}
	if err := c.client.Connect(); err != nil {
		return err
	}
	c.isConnected = true
	return nil
}

// Disconnect 断开连接并停止所有订阅处理器
func (c *Client) Disconnect() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if !c.isConnected || c.client == nil {
		return nil
	}
	close(c.stopChan)
	c.wg.Wait()
	if err := c.client.Disconnect(); err != nil {
		return err
	}
	c.isConnected = false
	return nil
}

// IsConnected 返回当前连接状态
func (c *Client) IsConnected() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.isConnected
}

// Publish 直接向指定主题发布消息（不带响应能力）
func (c *Client) Publish(topic string, data interface{}) error {
	if !c.IsConnected() {
		return fmt.Errorf("MessageBus 未连接")
	}
	message := types.MessageEnvelope{
		Versionable: commonDTO.NewVersionable(),
		RequestID:   uuid.NewString(),
		Payload:     data,
		ContentType: "application/json",
	}
	return c.client.Publish(message, topic)
}

// Request 发送带 RequestID 的请求并等待响应，适用于 RPC 风格通信
func (c *Client) Request(topic string, data interface{}) (types.MessageEnvelope, error) {
	if !c.IsConnected() {
		return types.MessageEnvelope{}, fmt.Errorf("MessageBus 未连接")
	}

	reqID := uuid.NewString()
	respCh := make(chan types.MessageEnvelope, 1)
	c.responseChMap.Store(reqID, respCh)
	defer c.responseChMap.Delete(reqID)

	message := types.MessageEnvelope{
		Versionable: commonDTO.NewVersionable(),
		RequestID:   reqID,
		Payload:     data,
		ContentType: "application/json",
	}

	if err := c.client.Publish(message, topic); err != nil {
		return types.MessageEnvelope{}, err
	}

	select {
	case resp := <-respCh:
		return resp, nil
	case <-time.After(c.timeout):
		return types.MessageEnvelope{}, fmt.Errorf("请求超时: %s", reqID)
	}
}

// Subscribe 订阅普通主题并提供消息处理函数
func (c *Client) Subscribe(topic string, handler MessageHandler) error {
	if !c.IsConnected() {
		return fmt.Errorf("MessageBus 未连接")
	}
	topicChannel := types.TopicChannel{
		Topic:    topic,
		Messages: make(chan types.MessageEnvelope, 100),
	}
	if err := c.client.Subscribe([]types.TopicChannel{topicChannel}, c.errorChan); err != nil {
		return err
	}
	c.wg.Add(1)
	go c.handleMessages(topic, topicChannel.Messages, handler)
	return nil
}

// SubscribeResponse 订阅响应主题，自动将响应分发到对应请求通道
func (c *Client) SubscribeResponse(topic string) error {
	if !c.IsConnected() {
		return fmt.Errorf("MessageBus 未连接")
	}
	topicChannel := types.TopicChannel{
		Topic:    topic,
		Messages: make(chan types.MessageEnvelope, 100),
	}
	if err := c.client.Subscribe([]types.TopicChannel{topicChannel}, c.errorChan); err != nil {
		return err
	}
	c.wg.Add(1)
	go c.handleResponseMessages(topic, topicChannel.Messages)
	return nil
}

// handleMessages 是普通订阅主题的处理逻辑
func (c *Client) handleMessages(topic string, ch chan types.MessageEnvelope, handler MessageHandler) {
	defer c.wg.Done()
	for {
		select {
		case msg, ok := <-ch:
			if !ok {
				return
			}
			actualTopic := msg.ReceivedTopic
			if actualTopic == "" {
				actualTopic = topic
			}
			msg.ReceivedTopic = actualTopic
			if err := handler(actualTopic, msg); err != nil {
				c.lc.Errorf("处理主题 %s 的消息时出错: %v", actualTopic, err)
			}
		case <-c.stopChan:
			return
		}
	}
}

// handleResponseMessages 专门处理响应类型消息，根据 RequestID 分发到等待通道
func (c *Client) handleResponseMessages(topic string, ch chan types.MessageEnvelope) {
	defer c.wg.Done()
	for {
		select {
		case msg, ok := <-ch:
			if !ok {
				return
			}
			if respChAny, ok := c.responseChMap.Load(msg.RequestID); ok {
				if respCh, ok := respChAny.(chan types.MessageEnvelope); ok {
					respCh <- msg
					continue
				}
			}
			c.lc.Warnf("未匹配的响应: %s", msg.RequestID)
		case <-c.stopChan:
			return
		}
	}
}

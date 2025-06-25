// Package messagebus 提供简化版的 EdgeX MessageBus 客户端封装
package messagebus

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"github.com/edgexfoundry/go-mod-messaging/v4/messaging"
	"github.com/edgexfoundry/go-mod-messaging/v4/pkg/types"
	"github.com/google/uuid"
)

// Client 表示一个简化版的 EdgeX MessageBus 客户端
type Client struct {
	client        messaging.MessageClient               // 底层消息客户端
	lc            logger.LoggingClient                  // 日志客户端
	isConnected   bool                                  // 是否已连接
	mutex         sync.RWMutex                          // 并发读写锁
	subscriptions map[string]chan types.MessageEnvelope // 订阅的主题及其消息通道
	errorChan     chan error                            // 错误通道
	stopChan      chan struct{}                         // 停止通道
	wg            sync.WaitGroup                        // 用于等待所有 goroutine 退出

	responseMap sync.Map // 请求响应映射：requestID -> 响应通道
}

// Config 表示 MessageBus 配置参数
type Config struct {
	Host     string
	Port     int
	Protocol string
	Type     string
	ClientID string
	Username string
	Password string
	QoS      int
}

// MessageHandler 定义处理消息的函数类型
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
		client:        client,
		lc:            lc,
		subscriptions: make(map[string]chan types.MessageEnvelope),
		errorChan:     make(chan error, 10),
		stopChan:      make(chan struct{}),
	}, nil
}

// Connect 连接到 MessageBus
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

// Disconnect 断开与 MessageBus 的连接，并停止所有订阅处理
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

// Publish 发布消息到指定主题
func (c *Client) Publish(topic string, data interface{}) error {
	if !c.IsConnected() {
		return fmt.Errorf("MessageBus未连接")
	}
	payload, err := toPayload(data)
	if err != nil {
		return err
	}
	return c.client.Publish(types.MessageEnvelope{
		CorrelationID: uuid.NewString(),
		Payload:       payload,
		ContentType:   "application/json",
	}, topic)
}

// Subscribe 订阅多个主题，并使用指定处理函数处理接收的消息
func (c *Client) Subscribe(topics []string, handler MessageHandler) error {
	if !c.IsConnected() {
		return fmt.Errorf("MessageBus未连接")
	}
	topicChannels := make([]types.TopicChannel, len(topics))
	for i, topic := range topics {
		ch := make(chan types.MessageEnvelope, 100)
		c.subscriptions[topic] = ch
		topicChannels[i] = types.TopicChannel{Topic: topic, Messages: ch}
	}
	if err := c.client.Subscribe(topicChannels, c.errorChan); err != nil {
		return err
	}
	for _, topic := range topics {
		c.wg.Add(1)
		go c.handleMessages(topic, handler)
	}
	return nil
}

// Request 请求并等待指定 requestID 的响应
func (c *Client) Request(topic string, data interface{}, timeout time.Duration) (*types.MessageEnvelope, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("MessageBus未连接")
	}
	payload, err := toPayload(data)
	if err != nil {
		return nil, err
	}
	requestID := uuid.NewString()
	msg := types.MessageEnvelope{
		CorrelationID: requestID,
		Payload:       payload,
		ContentType:   "application/json",
	}

	respCh := make(chan *types.MessageEnvelope, 1)
	c.responseMap.Store(requestID, respCh)
	defer c.responseMap.Delete(requestID)

	if err := c.client.Publish(msg, topic); err != nil {
		return nil, err
	}

	select {
	case resp := <-respCh:
		return resp, nil
	case <-time.After(timeout):
		return nil, fmt.Errorf("响应超时: %s", requestID)
	}
}

// handleMessages 处理订阅主题的消息循环
func (c *Client) handleMessages(topic string, handler MessageHandler) {
	defer c.wg.Done()
	ch, ok := c.subscriptions[topic]
	if !ok {
		return
	}
	for {
		select {
		case msg, ok := <-ch:
			if !ok {
				return
			}
			// 如果匹配请求响应ID，则直接返回给等待通道
			if v, ok := c.responseMap.Load(msg.CorrelationID); ok {
				respCh := v.(chan *types.MessageEnvelope)
				respCh <- &msg
				continue
			}
			actualTopic := msg.ReceivedTopic
			if actualTopic == "" {
				actualTopic = topic
			}
			_ = handler(actualTopic, msg)
		case <-c.stopChan:
			return
		}
	}
}

// IsConnected 判断当前是否已连接到 MessageBus
func (c *Client) IsConnected() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.isConnected
}

// toPayload 将任意数据转换为字节切片
func toPayload(data interface{}) ([]byte, error) {
	switch v := data.(type) {
	case []byte:
		return v, nil
	case string:
		return []byte(v), nil
	default:
		return json.Marshal(v)
	}
}

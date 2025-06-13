// Package messagebus provides a high-level client for EdgeX MessageBus operations.
//
// This package offers a simplified interface for interacting with EdgeX Foundry's
// MessageBus system, supporting MQTT and NATS protocols. It provides features like
// connection management, message publishing/subscribing, request-response patterns,
// and binary data handling.
//
// Example usage:
//
//	config := messagebus.Config{
//		Host:     "localhost",
//		Port:     1883,
//		Protocol: "tcp",
//		Type:     "mqtt",
//		ClientID: "my-client",
//	}
//
//	client, err := messagebus.NewClient(config, logger)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	err = client.Connect()
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer client.Disconnect()
//
//	// Publish a message
//	data := map[string]interface{}{"temperature": 25.6}
//	err = client.Publish("sensors/temperature", data)
//
//	// Subscribe to messages
//	handler := func(topic string, message types.MessageEnvelope) error {
//		fmt.Printf("Received: %s\n", string(message.Payload.([]byte)))
//		return nil
//	}
//	err = client.SubscribeSingle("sensors/#", handler)
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

// Client EdgeX MessageBus客户端封装
type Client struct {
	client        messaging.MessageClient
	lc            logger.LoggingClient
	isConnected   bool
	mutex         sync.RWMutex
	subscriptions map[string]chan types.MessageEnvelope
	errorChan     chan error
	stopChan      chan struct{}
	wg            sync.WaitGroup
}

// Config MessageBus配置结构
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

// MessageHandler 消息处理函数类型
type MessageHandler func(topic string, message types.MessageEnvelope) error

// NewClient 创建新的MessageBus客户端
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

	messageBus, err := messaging.NewMessageClient(messageBusConfig)
	if err != nil {
		return nil, fmt.Errorf("创建消息客户端失败: %v", err)
	}

	client := &Client{
		client:        messageBus,
		lc:            lc,
		subscriptions: make(map[string]chan types.MessageEnvelope),
		errorChan:     make(chan error, 10),
		stopChan:      make(chan struct{}),
	}

	return client, nil
}

// Connect 连接到MessageBus
func (c *Client) Connect() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.isConnected {
		c.lc.Debug("MessageBus客户端已连接")
		return nil
	}

	c.lc.Info("正在连接到EdgeX MessageBus...")
	if err := c.client.Connect(); err != nil {
		return fmt.Errorf("连接到MessageBus失败: %v", err)
	}

	c.isConnected = true
	c.lc.Info("✅ 成功连接到EdgeX MessageBus")
	return nil
}

// Disconnect 断开MessageBus连接
func (c *Client) Disconnect() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.isConnected || c.client == nil {
		return nil
	}

	// 停止所有订阅
	close(c.stopChan)
	c.wg.Wait()

	// 断开连接
	if err := c.client.Disconnect(); err != nil {
		c.lc.Errorf("断开MessageBus连接时发生错误: %v", err)
		return err
	}

	c.isConnected = false
	c.lc.Info("✅ 已断开EdgeX MessageBus连接")
	return nil
}

// IsConnected 检查连接状态
func (c *Client) IsConnected() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.isConnected
}

// Publish 发布消息到指定主题
func (c *Client) Publish(topic string, data interface{}) error {
	if !c.IsConnected() {
		return fmt.Errorf("MessageBus未连接")
	}

	// 创建消息信封
	var payload []byte
	var err error

	switch v := data.(type) {
	case []byte:
		payload = v
	case string:
		payload = []byte(v)
	default:
		payload, err = json.Marshal(data)
		if err != nil {
			return fmt.Errorf("序列化消息数据失败: %v", err)
		}
	}

	msgEnvelope := types.MessageEnvelope{
		CorrelationID: "MessageBus-" + uuid.New().String(),
		Payload:       payload,
		ContentType:   "application/json",
	}

	// 发布消息
	if err := c.client.Publish(msgEnvelope, topic); err != nil {
		c.lc.Errorf("发布消息到主题 '%s' 失败: %v", topic, err)
		return fmt.Errorf("发布消息失败: %v", err)
	}

	c.lc.Debugf("✅ 成功发布消息到主题: %s", topic)
	return nil
}

// PublishWithCorrelationID 使用指定的CorrelationID发布消息
func (c *Client) PublishWithCorrelationID(topic string, data interface{}, correlationID string) error {
	if !c.IsConnected() {
		return fmt.Errorf("MessageBus未连接")
	}

	var payload []byte
	var err error

	switch v := data.(type) {
	case []byte:
		payload = v
	case string:
		payload = []byte(v)
	default:
		payload, err = json.Marshal(data)
		if err != nil {
			return fmt.Errorf("序列化消息数据失败: %v", err)
		}
	}

	msgEnvelope := types.MessageEnvelope{
		CorrelationID: correlationID,
		Payload:       payload,
		ContentType:   "application/json",
	}

	if err := c.client.Publish(msgEnvelope, topic); err != nil {
		c.lc.Errorf("发布消息到主题 '%s' 失败: %v", topic, err)
		return fmt.Errorf("发布消息失败: %v", err)
	}

	c.lc.Debugf("✅ 成功发布消息到主题: %s (CorrelationID: %s)", topic, correlationID)
	return nil
}

// PublishBinaryData 发布二进制数据
func (c *Client) PublishBinaryData(topic string, data []byte) error {
	if !c.IsConnected() {
		return fmt.Errorf("MessageBus未连接")
	}

	if err := c.client.PublishBinaryData(data, topic); err != nil {
		c.lc.Errorf("发布二进制数据到主题 '%s' 失败: %v", topic, err)
		return fmt.Errorf("发布二进制数据失败: %v", err)
	}

	c.lc.Debugf("✅ 成功发布二进制数据到主题: %s", topic)
	return nil
}

// Subscribe 订阅主题并使用处理函数处理消息
func (c *Client) Subscribe(topics []string, handler MessageHandler) error {
	if !c.IsConnected() {
		return fmt.Errorf("MessageBus未连接")
	}

	// 创建主题通道
	topicChannels := make([]types.TopicChannel, len(topics))
	for i, topic := range topics {
		messageChan := make(chan types.MessageEnvelope, 100)
		c.subscriptions[topic] = messageChan
		topicChannels[i] = types.TopicChannel{
			Topic:    topic,
			Messages: messageChan,
		}
	}

	// 订阅主题
	if err := c.client.Subscribe(topicChannels, c.errorChan); err != nil {
		return fmt.Errorf("订阅主题失败: %v", err)
	}

	// 启动消息处理goroutine
	for _, topic := range topics {
		c.wg.Add(1)
		go c.handleMessages(topic, handler)
	}

	c.lc.Infof("✅ 成功订阅主题: %v", topics)
	return nil
}

// SubscribeSingle 订阅单个主题
func (c *Client) SubscribeSingle(topic string, handler MessageHandler) error {
	return c.Subscribe([]string{topic}, handler)
}

// handleMessages 处理接收到的消息
func (c *Client) handleMessages(topic string, handler MessageHandler) {
	defer c.wg.Done()

	messageChan, exists := c.subscriptions[topic]
	if !exists {
		c.lc.Errorf("主题 '%s' 的消息通道不存在", topic)
		return
	}

	c.lc.Debugf("开始处理主题 '%s' 的消息", topic)

	for {
		select {
		case message, ok := <-messageChan:
			if !ok {
				c.lc.Debugf("主题 '%s' 的消息通道已关闭", topic)
				return
			}

			// 调用用户提供的处理函数
			if err := handler(topic, message); err != nil {
				c.lc.Errorf("处理主题 '%s' 的消息时发生错误: %v", topic, err)
			}

		case err := <-c.errorChan:
			c.lc.Errorf("MessageBus订阅错误: %v", err)

		case <-c.stopChan:
			c.lc.Debugf("停止处理主题 '%s' 的消息", topic)
			return
		}
	}
}

// Unsubscribe 取消订阅主题
func (c *Client) Unsubscribe(topics ...string) error {
	if !c.IsConnected() {
		return fmt.Errorf("MessageBus未连接")
	}

	if err := c.client.Unsubscribe(topics...); err != nil {
		return fmt.Errorf("取消订阅失败: %v", err)
	}

	// 清理订阅记录
	for _, topic := range topics {
		if messageChan, exists := c.subscriptions[topic]; exists {
			close(messageChan)
			delete(c.subscriptions, topic)
		}
	}

	c.lc.Infof("✅ 成功取消订阅主题: %v", topics)
	return nil
}

// Request 发送请求并等待响应
func (c *Client) Request(message types.MessageEnvelope, requestTopic string, responseTopicPrefix string, timeout time.Duration) (*types.MessageEnvelope, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("MessageBus未连接")
	}

	response, err := c.client.Request(message, requestTopic, responseTopicPrefix, timeout)
	if err != nil {
		c.lc.Errorf("请求-响应操作失败: %v", err)
		return nil, fmt.Errorf("请求失败: %v", err)
	}

	c.lc.Debugf("✅ 请求-响应操作成功，主题: %s", requestTopic)
	return response, nil
}

// GetSubscribedTopics 获取当前订阅的主题列表
func (c *Client) GetSubscribedTopics() []string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	topics := make([]string, 0, len(c.subscriptions))
	for topic := range c.subscriptions {
		topics = append(topics, topic)
	}
	return topics
}

// GetErrorChannel 获取错误通道，用于监听MessageBus错误
func (c *Client) GetErrorChannel() <-chan error {
	return c.errorChan
}

// CreateMessageEnvelope 创建标准的消息信封
func (c *Client) CreateMessageEnvelope(data interface{}, correlationID string) (types.MessageEnvelope, error) {
	var payload []byte
	var err error

	switch v := data.(type) {
	case []byte:
		payload = v
	case string:
		payload = []byte(v)
	default:
		payload, err = json.Marshal(data)
		if err != nil {
			return types.MessageEnvelope{}, fmt.Errorf("序列化消息数据失败: %v", err)
		}
	}

	if correlationID == "" {
		correlationID = "MessageBus-" + uuid.New().String()
	}

	return types.MessageEnvelope{
		CorrelationID: correlationID,
		Payload:       payload,
		ContentType:   "application/json",
	}, nil
}

// PublishMessageEnvelope 直接发布消息信封
func (c *Client) PublishMessageEnvelope(envelope types.MessageEnvelope, topic string) error {
	if !c.IsConnected() {
		return fmt.Errorf("MessageBus未连接")
	}

	if err := c.client.Publish(envelope, topic); err != nil {
		c.lc.Errorf("发布消息信封到主题 '%s' 失败: %v", topic, err)
		return fmt.Errorf("发布消息信封失败: %v", err)
	}

	c.lc.Debugf("✅ 成功发布消息信封到主题: %s (CorrelationID: %s)", topic, envelope.CorrelationID)
	return nil
}

// HealthCheck 检查MessageBus客户端健康状态
func (c *Client) HealthCheck() error {
	if !c.IsConnected() {
		return fmt.Errorf("MessageBus未连接")
	}

	// 可以添加更多健康检查逻辑，比如发送ping消息等
	return nil
}

// GetClientInfo 获取客户端信息
func (c *Client) GetClientInfo() map[string]interface{} {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return map[string]interface{}{
		"connected":          c.isConnected,
		"subscribedTopics":   len(c.subscriptions),
		"errorChannelBuffer": len(c.errorChan),
	}
}

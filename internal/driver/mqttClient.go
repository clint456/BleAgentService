package driver

import (
	"encoding/json"
	"fmt"
	"strings"

	messagebus "github.com/clint456/edgex-messagebus-client"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/edgexfoundry/go-mod-messaging/v4/messaging"
	"github.com/edgexfoundry/go-mod-messaging/v4/pkg/types"
)

// createMessageBusClient 创建MessageBus监听客户端
func (d *Driver) createMessageBusClient() (*messagebus.Client, error) {
	// 配置MessageBus客户端
	config := messagebus.Config{
		Host:     d.serviceConfig.MQTTBrokerInfo.Host,
		Port:     d.serviceConfig.MQTTBrokerInfo.Port,
		Protocol: strings.ToLower(d.serviceConfig.MQTTBrokerInfo.Schema),
		Type:     "mqtt",
		ClientID: d.serviceConfig.MQTTBrokerInfo.ClientId,
		QoS:      d.serviceConfig.MQTTBrokerInfo.Qos,
	}

	// 处理认证
	if d.serviceConfig.MQTTBrokerInfo.AuthMode == AuthModeUsernamePassword {
		credentials, err := d.GetCredentials(d.sdk.SecretProvider(), d.serviceConfig.MQTTBrokerInfo.CredentialsName)
		if err != nil {
			return nil, fmt.Errorf("获取MQTT认证信息失败: %w", err)
		}
		config.Username = credentials.Username
		config.Password = credentials.Password
	}

	// 创建客户端
	client, err := messagebus.NewClient(config, d.logger)
	if err != nil {
		return nil, fmt.Errorf("创建MessageBus客户端失败: %w", err)
	}

	// 连接到MessageBus
	d.logger.Infof("连接到MessageBus: %s:%d", config.Host, config.Port)
	if err := client.Connect(); err != nil {
		return nil, fmt.Errorf("连接MessageBus失败: %w", err)
	}

	// 订阅主题
	incomingTopic := d.serviceConfig.MQTTBrokerInfo.IncomingTopic
	subscribeTopics := []string{incomingTopic}

	if err := client.Subscribe(subscribeTopics, d.onMessageBusDataReceived); err != nil {
		client.Disconnect()
		return nil, fmt.Errorf("订阅主题 '%s' 失败: %w", incomingTopic, err)
	}

	d.logger.Infof("成功订阅到 '%s' 用于接收数据", incomingTopic)
	return client, nil
}

// onMessageBusDataReceived 处理通过新MessageBus库接收到的消息
func (d *Driver) onMessageBusDataReceived(topic string, message types.MessageEnvelope) error {
	// 获取接收到的消息主题
	incomingTopic := topic
	// 从消息主题中移除订阅主题部分，提取元数据
	incomingTopic = strings.Replace(incomingTopic, "edgex", "", -1)

	// 解析消息的 payload（JSON 格式）
	asyncData := make(map[string]interface{})

	// 处理不同类型的payload
	var payloadBytes []byte
	switch payload := message.Payload.(type) {
	case []byte:
		payloadBytes = payload
	case string:
		payloadBytes = []byte(payload)
	default:
		// 如果payload已经是map类型，直接使用
		if data, ok := payload.(map[string]interface{}); ok {
			asyncData = data
		} else {
			// 尝试序列化为JSON再解析
			var err error
			payloadBytes, err = json.Marshal(payload)
			if err != nil {
				d.logger.Errorf("序列化payload失败: %v", err)
				return err
			}
		}
	}

	// 如果有payloadBytes，则解析JSON
	if len(payloadBytes) > 0 {
		err := json.Unmarshal(payloadBytes, &asyncData)
		if err != nil {
			d.logger.Errorf("反序列化payload失败: %v", err)
			return err
		}
	}

	// 记录接收到的消息信息
	d.logger.Debugf("收到消息 - 主题: %s, CorrelationID: %s", topic, message.CorrelationID)

	// 转发到 MessageBus
	err := d.publishToMessageBus(asyncData, "edgex/data"+incomingTopic)
	if err != nil {
		d.logger.Errorf("转发到MessageBus失败: %v", err)
		return err
	}

	// 将接收到的数据向蓝牙发送器异步传输数据
	d.sendToBluetoothTransmitter(asyncData)

	return nil
}

/* ============================ 以下是使用go-mod-messaging使用Mqtt ==============================*/

func (d *Driver) InitMessageBusClient(ClientID string, Host string, Port int) (messaging.MessageClient, errors.EdgeX) {
	messageBus, err := messaging.NewMessageClient(types.MessageBusConfig{
		Broker: types.HostInfo{
			Host:     Host,
			Port:     Port,
			Protocol: "tcp",
		},
		Type: "mqtt",
		Optional: map[string]string{
			"ClientId": ClientID,
			"Username": "",
			"Password": ""}})

	if err != nil {
		return nil, errors.NewCommonEdgeXWrapper(fmt.Errorf("⛔️ 消息客户端失败: %v", err))
	}
	if messageBus == nil {
		return nil, errors.NewCommonEdgeXWrapper(fmt.Errorf("⛔️ 消息客户端为 nil"))
	}
	// 连接到 Broker
	if err := messageBus.Connect(); err != nil {
		return nil, errors.NewCommonEdgeXWrapper(fmt.Errorf("⛔️ 连接到 MQTT Broker 失败: %v", err))
	}
	d.logger.Debugf("消息客户端 %s 初始化成功", ClientID)
	return messageBus, nil
}

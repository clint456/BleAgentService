package driver

import (
	"encoding/json"

	"github.com/edgexfoundry/go-mod-messaging/v4/pkg/types"
	"github.com/google/uuid"
)

// publishToMessageBus 发布数据到 MessageBus
func (s *Driver) publishToMessageBus(data map[string]interface{}, topic string) error {
	// 序列化数据为 JSON
	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// 创建 MessageEnvelope
	msgEnvelope := types.MessageEnvelope{
		CorrelationID: "MessageEnvelope-" + uuid.New().String(), // 假设有生成关联 ID 的方法
		Payload:       payload,
		ContentType:   "application/json",
	}

	// 发布消息到 MessageBus
	err = s.transmitClient.Publish(msgEnvelope, topic)
	if err != nil {
		return err
	}

	s.lc.Debugf("📤 [EdgeX %v 服务数据转发] 成功发布到 MessageBus, 主题: %v", s.serviceConfig.MQTTBrokerInfo.IncomingTopic, topic)
	return nil
}

// sendToBluetoothTransmitter 异步传输到蓝牙发送器（占位实现）
func (s *Driver) sendToBluetoothTransmitter(data map[string]interface{}) {
	// 实现蓝牙异步传输逻辑
	s.lc.Debugf("📡 [EdgeX %v 服务数据传输] 正在向蓝牙发送器传输数据", s.serviceConfig.MQTTBrokerInfo.IncomingTopic)
	// 具体蓝牙传输逻辑待实现
	go SendJSONOverUART(s.ble.queue, data)
}

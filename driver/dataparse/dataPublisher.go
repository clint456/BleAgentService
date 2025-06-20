package dataparse

import (
	"device-ble/driver/ble"
	"fmt"

	"github.com/labstack/gommon/log"
)

// MessageBusClient 接口（与 mqttbus 保持一致，彻底解耦）
type MessageBusClient interface {
	IsConnected() bool
	Publish(topic string, data interface{}) error
}

// PublishToMessageBus 发布数据到MessageBus
func PublishToMessageBus(client MessageBusClient, data map[string]interface{}, topic string) error {
	if !client.IsConnected() {
		return fmt.Errorf("MessageBus客户端未连接")
	}
	err := client.Publish(topic, data)
	if err != nil {
		return fmt.Errorf("发布到MessageBus失败: %v", err)
	}
	return nil
}

// SendToBlE 异步传输到蓝牙发送器
func SendToBlE(controller *ble.BLEController, data map[string]interface{}) {
	if controller != nil {
		SendJSONOverUART(controller.Queue, data)
		log.Debug("数据已发送到蓝牙传输器")
	} else {
		log.Warn("BLE控制器未初始化，无法发送数据")
	}
}

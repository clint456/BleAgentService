package dataparse

import (
	"device-ble/internal/interfaces"
	"device-ble/pkg/ble"
	"encoding/json"
	"fmt"

	"github.com/labstack/gommon/log"
)

// PublishToMessageBus 发布数据到MessageBus
func PublishToMessageBus(client interfaces.MessageBusClient, data interface{}, topic string) error {
	// 数据类型断言
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("序列化数据失败: %v", err)
	}
	err = client.Publish(topic, dataBytes)
	if err != nil {
		return fmt.Errorf("发布到MessageBus失败: %v", err)
	}
	return nil
}

// SendToBlE 异步传输到蓝牙发送器
func SendToBlE(controller interfaces.BLEController, data interface{}) {
	if controller != nil {
		queueIface := controller.GetQueue()
		if queue, ok := queueIface.(interfaces.SerialQueue); ok {
			ble.SendJSONOverBLE(queue, data)
			log.Debug("数据已发送到蓝牙传输器")
		} else {
			log.Warn("BLE控制器队列类型断言失败，无法发送数据")
		}
	} else {
		log.Warn("BLE控制器未初始化，无法发送数据")
	}
}

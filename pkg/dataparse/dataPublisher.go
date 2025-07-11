package dataparse

import (
	"device-ble/internal/interfaces"
	"device-ble/pkg/ble"
	"fmt"
)

// PublishToMessageBus 发布数据到MessageBus。

// SendToBlE 异步传输到蓝牙发送器。
func SendToBlE(controller interfaces.BLEController, data interface{}) error {
	if controller != nil {
		queueIface := controller.GetQueue()
		if queue, ok := queueIface.(interfaces.SerialQueueInterface); ok {
			if err := ble.SendJSONOverBLE(queue, data); err == nil {
				return nil
			} else {
				return fmt.Errorf("向BLE控制器发送数据失败")
			}
		} else {
			return fmt.Errorf("BLE控制器队列类型断言失败，无法发送数据")
		}
	} else {
		return fmt.Errorf("BLE控制器未初始化，无法发送数据")
	}
}

package driver

import (
	"fmt"

	"github.com/google/uuid"
)

// publishToMessageBus 发布数据到MessageBus
func (d *Driver) publishToMessageBus(data map[string]interface{}, topic string) error {
	// 检查客户端是否已连接
	if !d.messageBusClient.IsConnected() {
		d.logger.Error("MessageBus客户端未连接")
		return fmt.Errorf("MessageBus客户端未连接")
	}

	// 使用统一的messagebus客户端发布消息
	err := d.messageBusClient.PublishWithCorrelationID(topic, data, "MessageEnvelope-"+uuid.New().String())
	if err != nil {
		d.logger.Errorf("发布到MessageBus失败: %v", err)
		return err
	}

	d.logger.Debugf("成功发布到MessageBus, 主题: %s", topic)
	return nil
}

// sendToBluetoothTransmitter 异步传输到蓝牙发送器
func (d *Driver) sendToBluetoothTransmitter(data map[string]interface{}) {
	d.logger.Debug("正在向蓝牙发送器传输数据")

	// TODO: 实现具体的蓝牙传输逻辑
	// 这里需要通过BLE控制器发送数据
	if d.bleController != nil {
		// 将数据通过串口队列发送
		SendJSONOverUART(d.bleController.queue, data)
		d.logger.Debug("数据已发送到蓝牙传输器")
	} else {
		d.logger.Warn("BLE控制器未初始化，无法发送数据")
	}
}

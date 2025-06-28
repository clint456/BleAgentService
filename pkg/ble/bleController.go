package ble

import (
	"device-ble/internal/interfaces"
	"device-ble/pkg/uart"
	"strings"
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
)

// BLEController 蓝牙低功耗控制器，管理BLE设备的初始化、命令发送和状态控制。
type BLEController struct {
	Port   *uart.SerialPort
	Queue  interfaces.SerialQueueInterface
	logger logger.LoggingClient
}

// NewBLEController 创建新的BLE控制器。
func NewBLEController(port *uart.SerialPort, queue interfaces.SerialQueueInterface, logger logger.LoggingClient) *BLEController {
	return &BLEController{
		Port:   port,
		Queue:  queue,
		logger: logger,
	}
}

// InitializeAsPeripheral 启动初始化BLE设备为外围设备模式。
func (c *BLEController) InitializeAsPeripheral() error {
	initCommands := []BLECommand{
		CommandReset,
		CommandInitPeripheral,
		CommandSetAdvertisingParams,
		CommandCreateGATTService,
		CommandCreateGATTCharacteristic,
		CommandCompleteGATTService,
		CommandSetDeviceName,
		CommandStartAdvertising,
	}
	for _, cmd := range initCommands {
		response, err := c.Queue.SendCommand([]byte(cmd), 2*time.Second, 1*time.Second, 100*time.Millisecond)
		if strings.Contains(response, "OK") {
			c.logger.Infof("✅ 发送 %q 成功, 回显： %v", cmd, response)
		} else if strings.Contains(response, "ERROR") {
			c.logger.Errorf("⛔️  发送 %q 失败: %q , 回显： %v ", cmd, err, response)
		} else if err != nil {
			c.logger.Errorf("❗❓未知回显 :%v, response:%v", err, response)
		} else {
			c.logger.Warnf("❗❓未知回显, response:%v", response)
		}
	}

	c.logger.Info("BLE设备已成功初始化为外围设备")
	return nil
}

// InitializeAsPeripheral 自定义初始化BLE设备。
func (c *BLEController) CustomInitializeBle(cmds []string) error {
	for _, cmd := range cmds {
		response, err := c.Queue.SendCommand([]byte(cmd), 2*time.Second, 1*time.Second, 100*time.Millisecond)
		if strings.Contains(response, "OK") {
			c.logger.Infof("✅ 发送 %q 成功, 回显： %v", cmd, response)
		} else if strings.Contains(response, "ERROR") {
			c.logger.Errorf("⛔️  发送 %q 失败: %q , 回显： %v ", cmd, err, response)
		} else if err != nil {
			c.logger.Errorf("❗❓未知回显 :%v, response:%v", err, response)
		} else {
			c.logger.Warnf("❗❓未知回显, response:%v", response)
		}
	}

	c.logger.Info("BLE设备已成功初始化为外围设备")
	return nil
}

func (c *BLEController) GetQueue() interfaces.SerialQueueInterface {
	return c.Queue
}

func (c *BLEController) Close() error {
	err := c.Queue.Close()
	if err != nil {
		return err
	}
	return nil
}

// 向BLE发送一条数据（MTU小于247)
func (c *BLEController) SendSingle(cmd string) error {
	response, err := c.Queue.SendCommand([]byte(cmd), 2*time.Second, 1*time.Second, 100*time.Millisecond)
	if err != nil {
		c.logger.Errorf("❌发送%v, 出现错误 :%v, response:%v", cmd, err, response)
		return err
	} else {
		if strings.Contains(response, "OK") {
			c.logger.Infof("✅ 发送 %q 成功, 回显： %v", cmd, response)
		} else if strings.Contains(response, "ERROR") {
			c.logger.Errorf("⛔️  发送 %q 失败: %q , 回显： %v ", cmd, err, response)
		} else {
			c.logger.Warnf("❗❓  未知回显, response:%v", response)
		}
	}
	c.logger.Info("BLE 单条指令发送成功")
	return nil
}

// 多条指令（单条指令MTU小于247)
// 如果需要发送JSON数据请使用jsonSender中的方法
func (c *BLEController) SendMulti(cmds []string) error {
	for _, cmd := range cmds {
		response, err := c.Queue.SendCommand([]byte(cmd), 2*time.Second, 1*time.Second, 100*time.Millisecond)
		if err != nil {
			c.logger.Errorf("❌发送%v, 出现错误 :%v, response:%v", cmd, err, response)
			return err
		} else {
			if strings.Contains(response, "OK") {
				c.logger.Infof("✅ 发送 %q 成功, 回显： %v", cmd, response)
			} else if strings.Contains(response, "ERROR") {
				c.logger.Errorf("⛔️  发送 %q 失败: %q , 回显： %v ", cmd, err, response)
			} else {
				c.logger.Warnf("❗❓  未知回显, response:%v", response)
			}
		}
	}

	c.logger.Info("BLE 多条指令发送成功")
	return nil
}

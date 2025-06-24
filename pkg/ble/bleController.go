package ble

import (
	"device-ble/internal/interfaces"
	"device-ble/pkg/uart"
	"log"
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
)

// BLEController 蓝牙低功耗控制器
// 职责：管理BLE设备的初始化、命令发送和状态控制
type BLEController struct {
	Port   *uart.SerialPort
	Queue  interfaces.SerialQueue
	logger logger.LoggingClient
}

// NewBLEController 创建新的BLE控制器
func NewBLEController(port *uart.SerialPort, queue interfaces.SerialQueue, logger logger.LoggingClient) *BLEController {
	return &BLEController{
		Port:   port,
		Queue:  queue,
		logger: logger,
	}
}

// InitializeAsPeripheral 初始化BLE设备为外围设备模式
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
		// 通过串口发送
		response, err := c.Queue.SendCommand([]byte(cmd), time.Second, 2*time.Second)
		if response == "OK\n" {
			log.Printf("✅ 发送 %v 成功, 回显： %v\n", cmd, response)
		} else if response == "ERROR\n" {
			log.Printf("⛔️  发送 %v 失败, 回显： %v\n", cmd, response)
		} else {
			log.Printf("❗❓未知回显 :%s, response:%v\n", err, response)
		}
	}

	c.logger.Info("BLE设备已成功初始化为外围设备")
	return nil
}

func (c *BLEController) GetQueue() interfaces.SerialQueue {
	return c.Queue
}

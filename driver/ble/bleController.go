package ble

import (
	"device-ble/driver/uart"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
)

// BLEController 蓝牙低功耗控制器
// 职责：管理BLE设备的初始化、命令发送和状态控制
type BLEController struct {
	Port   *uart.SerialPort
	Queue  *uart.SerialQueue
	logger logger.LoggingClient
}

// NewBLEController 创建新的BLE控制器
func NewBLEController(port *uart.SerialPort, queue *uart.SerialQueue, logger logger.LoggingClient) *BLEController {
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
		if err := c.executeCommand(cmd); err != nil {
			return fmt.Errorf("执行命令 %s 失败: %w", cmd, err)
		}
		// 等待初始化命令处理
		time.Sleep(1000 * time.Millisecond)
	}

	c.logger.Info("BLE设备已成功初始化为外围设备")
	return nil
}

// executeCommand 执行单个BLE命令
func (c *BLEController) executeCommand(cmd BLECommand) error {
	response, err := c.sendCommandAndWaitResponse(cmd)
	if err != nil {
		return err
	}

	if !c.isSuccessResponse(response) {
		return fmt.Errorf("命令执行失败: %s", response)
	}

	c.logger.Debugf("命令执行成功: %s", cmd)
	return nil
}

// sendCommandAndWaitResponse 发送命令并等待响应
func (c *BLEController) sendCommandAndWaitResponse(cmd BLECommand) (string, error) {
	if err := c.writeCommand(cmd); err != nil {
		return "", fmt.Errorf("写入命令失败: %w", err)
	}

	return c.readResponse()
}

// writeCommand 写入启动命令到串口
func (c *BLEController) writeCommand(cmd BLECommand) error {
	_, err := c.Port.Write([]byte(cmd))
	if err != nil {
		return fmt.Errorf("串口写入失败: %w", err)
	}
	return nil
}

// readResponse 读取命令响应
func (c *BLEController) readResponse() (string, error) {
	const responseTimeout = 3 * time.Second
	const readInterval = 20 * time.Millisecond

	var fullResponse strings.Builder
	startTime := time.Now()

	for time.Since(startTime) < responseTimeout {
		line, err := c.readLine()
		if err != nil {
			if err == io.EOF {
				time.Sleep(readInterval)
				continue
			}
			return "", fmt.Errorf("读取响应失败: %w", err)
		}

		if line == "" {
			continue
		}

		fullResponse.WriteString(line + "\n")
		c.logger.Debugf("收到响应: %s", line)

		if c.isTerminalResponse(line) {
			return fullResponse.String(), nil
		}
	}

	return "", fmt.Errorf("读取响应超时")
}

// readLine 读取一行数据并清理格式
func (c *BLEController) readLine() (string, error) {
	rawLine, err := c.Port.ReadLine()
	if err != nil {
		return "", err
	}

	return strings.Trim(string(rawLine), "\r\n"), nil
}

// isTerminalResponse 检查是否为终端响应
func (c *BLEController) isTerminalResponse(line string) bool {
	return line == "OK" || line == "ERROR" || strings.HasPrefix(line, "+CME ERROR:")
}

// isSuccessResponse 检查响应是否表示成功
func (c *BLEController) isSuccessResponse(response string) bool {
	return strings.Contains(response, "OK") &&
		!strings.Contains(response, "ERROR") &&
		!strings.Contains(response, "+CME ERROR:")
}

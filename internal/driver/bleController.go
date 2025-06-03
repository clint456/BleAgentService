package driver

import (
	"fmt"
	"log"
	"strings"
	"time"
)

type Controller struct {
	serial *SerialPort
}

// 创建控制器
func NewController(serial *SerialPort) *Controller {
	return &Controller{serial: serial}
}

// 发送 AT 指令并返回响应
func (c *Controller) SendATCommand(cmd string) (string, error) {
	fullCmd := fmt.Sprintf("%s\r\n", cmd)
	if err := c.serial.Write(fullCmd); err != nil {
		return "", err
	}
	return c.serial.ReadLine()
}

type BleController struct {
	serial *SerialPort
	debug  bool
}

func NewBleController(sp *SerialPort, debug bool) *BleController {
	return &BleController{serial: sp, debug: debug}
}

func (b *BleController) sendCommand(cmd BleCommand) (string, error) {
	if err := b.serial.Write(string(cmd)); err != nil {
		return "", fmt.Errorf("写入失败: %w", err)
	}
	time.Sleep(1000 * time.Millisecond)
	var fullResponse string
	for {
		line, err := b.serial.ReadLine()
		if err != nil {
			return "", fmt.Errorf("读取失败: %w", err)
		}
		line = trimCRLF(line)

		if line == "" {
			continue // 跳过空行
		}

		if b.debug {
			log.Printf("🧾 收到: %q", line)
		}

		fullResponse += line + "\n"

		// 检查是否是结尾状态
		if line == "OK" {
			return fullResponse, nil
		}
		if line == "ERROR" {
			return fullResponse, fmt.Errorf("命令返回 ERROR")
		}
		if strings.HasPrefix(line, "+CME ERROR:") {
			return fullResponse, fmt.Errorf("模块错误: %s", line)
		}
	}
}

// trimCRLF 去除 AT 响应行首尾 CR/LF 字符
func trimCRLF(s string) string {
	return strings.Trim(s, "\r\n")
}

// 初始化为外围设备并启动广播
func (b *BleController) InitAsPeripheral() error {
	commands := []BleCommand{
		ATRESET,
		ATVERSION,
		ATINIT_2,
		ATADV,
		ATGATTSSRV,
		ATGATTSCHAR,
		ATGATTSSRVDONE,
		ATNAME,
		ATADDR,
		ATADVSTART,
	}

	for _, cmd := range commands {
		_, err := b.sendCommand(cmd)
		if err != nil {
			return err
		}
	}
	return nil
}

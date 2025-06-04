package driver

import (
	"fmt"
	"io"
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

type BleController struct {
	serial *SerialPort
	debug  bool
}

func NewBleController(sp *SerialPort, debug bool) *BleController {
	return &BleController{serial: sp, debug: debug}
}

func (b *BleController) sendCommand(cmd BleCommand) (string, error) {
	if _, err := b.serial.Write([]byte(cmd)); err != nil {
		return "", fmt.Errorf("写入失败: %w", err)
	}
	time.Sleep(1000 * time.Millisecond)
	var fullResponse string
	start := time.Now()
	timeout := 3 * time.Second
	for {
		if time.Since(start) > timeout {
			return "", fmt.Errorf("❌ 读取超时")
		}
		rawLine, err := b.serial.ReadLine()
		line := string(rawLine)
		if err != nil {
			if err == io.EOF {
				time.Sleep(20 * time.Millisecond) // 小延时再读
				continue
			}
			return "", fmt.Errorf("❌ 读取失败: %w", err)
		}

		line = trimCRLF(line) // 注意这里传参

		if line == "" {
			continue // 跳过空行
		}

		if b.debug {
			log.Printf("✳️  命令: %v", cmd)
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
		// ATVERSION,
		ATINIT_2,
		ATADV,
		ATGATTSSRV,
		ATGATTSCHAR,
		ATGATTSSRVDONE,
		ATNAME,
		// ATADDR,
		ATADVSTART,
		// ATQBLETRANMODE,
	}

	for _, cmd := range commands {
		_, err := b.sendCommand(cmd)
		if err != nil {
			return err
		}
	}
	return nil
}

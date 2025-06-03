package driver

import (
	"fmt"
	"log"
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
	err := b.serial.Write(string(cmd))
	if err != nil {
		return "", err
	}
	resp, err := b.serial.ReadLine()
	if b.debug {
		log.Printf("🔄 Cmd: %q\n📥 Resp: %s\n❗Err: %v\n", cmd, resp, err)
	}
	return resp, err
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

	var lastErr error
	for _, cmd := range commands {
		_, err := b.sendCommand(cmd)
		if err != nil {
			lastErr = fmt.Errorf("❌ 命令 %q 执行失败: %v", cmd, err)
			// 继续执行剩下命令，记录最后的错误
		}
	}
	return lastErr
}

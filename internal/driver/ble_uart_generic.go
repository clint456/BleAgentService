package driver

import (
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
	"github.com/tarm/serial"
)

// 串口控制模块结构体,
// config口初始化配置指针,
// conn:端口结构体指针,
// rxbuf:slice动态数组,
// portStatus：当autoevent被执行启用时，该值为true,
type Uart struct {
	Name       string
	config     *serial.Config
	conn       *serial.Port
	rxbuf      []byte
	enable     bool
	portStatus bool
}

// 传入：设备名、波特率
// 返回：Uart结构体指针
func NewUart(dev string, baud int, timeout int, lc logger.LoggingClient) (*Uart, error) {
	config := &serial.Config{
		Name:        dev,
		Baud:        baud,
		ReadTimeout: time.Duration(timeout) * time.Second, // 表示timeout s
	}

	var err error
	//尝试开启串口,成功返回(*serial.Port,nill),失败返回(nill,error)
	conn, err := serial.OpenPort(config)
	if err != nil || conn == nil {
		lc.Errorf("NewUart(): 开启 %s 串口失败 : %v", config.Name, err)
		return nil, err

	}
	//成功创建
	return &Uart{config: config, conn: conn, enable: true, portStatus: false}, nil
}

// 传入：读取最大长度
// 返回： error
func (dev *Uart) UartRead(maxbytes int, ls logger.LoggingClient) error {
	if !dev.enable {
		lc.Errorf("UartRead(): 串口 %s 未开启成功", dev.config.Name)
	}
	var buf []byte

	// serial包方法 一次读取的最大值为16byte
	// 对做取余处理
	readCount := (maxbytes / 16) + 1

	lc.Debugf("UartRead(): 串口 读值长度为: %v", readCount)

	if dev.portStatus {

	}

}

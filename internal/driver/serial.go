package driver

import (
	"bufio"
	"time"

	"github.com/tarm/serial"
)

type SerialPort struct {
	port   *serial.Port
	reader *bufio.Reader
}

// 打开串口
func NewSerialPort(portName string, baud int) (*SerialPort, error) {
	cfg := &serial.Config{
		Name:        portName,
		Baud:        baud,
		ReadTimeout: time.Second * 2,
	}
	p, err := serial.OpenPort(cfg)
	if err != nil {
		return nil, err
	}
	return &SerialPort{
		port:   p,
		reader: bufio.NewReader(p),
	}, nil
}

// 发送命令
func (s *SerialPort) Write(cmd string) error {
	_, err := s.port.Write([]byte(cmd))
	return err
}

// 读取回应（读取到 \n）
func (s *SerialPort) ReadLine() (string, error) {
	return s.reader.ReadString('\n')
}

// 关闭串口
func (s *SerialPort) Close() error {
	return s.port.Close()
}

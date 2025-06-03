package driver

import (
	"bufio"
	"fmt"
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

func (sp *SerialPort) ReadLine() (string, error) {
	if sp.reader == nil {
		return "", fmt.Errorf("串口未打开")
	}
	return sp.reader.ReadString('\n')
}

// 读取指定大小的数据
func (s *SerialPort) ReadExact(size int) ([]byte, error) {
	buf := make([]byte, size)
	total := 0

	for total < size {
		n, err := s.port.Read(buf[total:])
		if err != nil {
			return nil, err
		}
		if n == 0 {
			break
		}
		total += n
	}

	return buf[:total], nil
}

// 关闭串口
func (s *SerialPort) Close() error {
	return s.port.Close()
}

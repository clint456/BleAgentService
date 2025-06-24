package uart

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	internalif "device-ble/internal/interfaces"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"github.com/tarm/serial"
)

// SerialPortConfig 串口配置
type SerialPortConfig struct {
	PortName    string
	BaudRate    int
	ReadTimeout time.Duration
}

// SerialPort 串口通信管理器
// 职责：管理串口的连接、读写操作和生命周期
type SerialPort struct {
	port   *serial.Port
	reader *bufio.Reader
	mutex  sync.RWMutex
	logger logger.LoggingClient
}

// NewSerialPort 创建新的串口实例
func NewSerialPort(config internalif.SerialConfig, logger logger.LoggingClient) (*SerialPort, error) {
	serialConfig := &serial.Config{
		Name:        config.PortName,
		Baud:        config.BaudRate,
		ReadTimeout: time.Duration(config.ReadTimeout),
	}

	port, err := serial.OpenPort(serialConfig)
	if err != nil {
		return nil, fmt.Errorf("打开串口失败: %w", err)
	}

	sp := &SerialPort{
		port:   port,
		reader: bufio.NewReader(port),
		logger: logger,
	}

	logger.Infof("串口已打开: %s, 波特率: %d", config.PortName, config.BaudRate)
	return sp, nil
}

// Write 写入数据到串口
func (sp *SerialPort) Write(data []byte) (int, error) {
	sp.mutex.Lock()
	defer sp.mutex.Unlock()

	if len(data) == 0 {
		return 0, fmt.Errorf("写入数据不能为空")
	}

	bytesWritten, err := sp.port.Write(data)
	if err != nil {
		sp.logger.Errorf("串口写入失败: %v", err)
		return bytesWritten, fmt.Errorf("串口写入失败: %w", err)
	}

	sp.logger.Debugf("串口写入成功: %d 字节", bytesWritten)
	return bytesWritten, nil
}

// ReadLine 从串口读取一行数据
func (sp *SerialPort) ReadLine() ([]byte, error) {
	sp.mutex.RLock()
	defer sp.mutex.RUnlock()

	line, err := sp.reader.ReadBytes('\n')
	if err != nil && err != io.EOF {
		sp.logger.Errorf("串口读取失败: %v", err)
		return nil, fmt.Errorf("串口读取失败: %w", err)
	}

	cleanLine := strings.TrimRight(string(line), "\r\n")
	sp.logger.Debugf("串口读取: %d 字节, 内容: %s", len(line), cleanLine)

	return line, err
}

// Close 关闭串口连接
func (sp *SerialPort) Close() error {
	sp.mutex.Lock()
	defer sp.mutex.Unlock()

	if sp.port == nil {
		return nil
	}

	err := sp.port.Close()
	if err != nil {
		sp.logger.Errorf("关闭串口失败: %v", err)
		return fmt.Errorf("关闭串口失败: %w", err)
	}

	sp.logger.Info("串口已关闭")
	return nil
}

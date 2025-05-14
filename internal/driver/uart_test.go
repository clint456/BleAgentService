package driver

import (
	"io"
	"testing"
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/tarm/serial"
)

// 创建 mock 日志客户端
type mockLogger struct {
	logger.LoggingClient
}

func (m *mockLogger) Debugf(format string, args ...interface{}) {}
func (m *mockLogger) Errorf(format string, args ...interface{}) {}

func TestNewUart(t *testing.T) {
	// 创建 mock 串口
	mockPort := &MockSerialPort{}
	mockPort.On("Close").Return(nil)

	// 模拟 serial.OpenPort 返回 mockPort
	// 注意：由于 serial.OpenPort 是全局函数，这里我们直接构造 Uart
	config := &serial.Config{
		Name:        "COM1",
		Baud:        9600,
		ReadTimeout: time.Second,
	}
	uart := &Uart{
		config:     config,
		conn:       mockPort,
		enable:     true,
		portStatus: false,
	}

	// 测试初始化
	assert.Equal(t, "COM1", uart.config.Name)
	assert.Equal(t, true, uart.enable)
	assert.Equal(t, false, uart.portStatus)
	assert.NotNil(t, uart.conn)
}

func TestUartRead_Success(t *testing.T) {
	// 创建 mock 串口和日志
	mockPort := &MockSerialPort{}
	mockLog := &mockLogger{}

	// 设置 mock 行为：模拟读取 5 字节数据
	mockPort.On("Read", mock.Anything).Return(5, nil).Once()
	mockPort.On("Flush").Return(nil)

	// 创建 Uart 实例
	uart := &Uart{
		config:     &serial.Config{Name: "COM1"},
		conn:       mockPort,
		enable:     true,
		portStatus: false,
	}

	// 执行读取
	err := uart.UartRead(16, mockLog)

	// 验证结果
	assert.NoError(t, err)
	assert.Equal(t, 5, len(uart.rxbuf))
	mockPort.AssertExpectations(t)
}

func TestUartRead_EOF(t *testing.T) {
	// 创建 mock 串口和日志
	mockPort := &MockSerialPort{}
	mockLog := &mockLogger{}

	// 设置 mock 行为：模拟 EOF
	mockPort.On("Read", mock.Anything).Return(0, io.EOF).Once()
	mockPort.On("Flush").Return(nil)

	// 创建 Uart 实例
	uart := &Uart{
		config:     &serial.Config{Name: "COM1"},
		conn:       mockPort,
		enable:     true,
		portStatus: false,
	}

	// 执行读取
	err := uart.UartRead(16, mockLog)

	// 验证结果
	assert.NoError(t, err)
	assert.Empty(t, uart.rxbuf)
	mockPort.AssertExpectations(t)
}

func TestUartRead_Error(t *testing.T) {
	// 创建 mock 串口和日志
	mockPort := &MockSerialPort{}
	mockLog := &mockLogger{}

	// 设置 mock 行为：模拟错误
	mockPort.On("Read", mock.Anything).Return(0, io.ErrUnexpectedEOF).Once()
	mockPort.On("Flush").Return(nil)

	// 创建 Uart 实例
	uart := &Uart{
		config:     &serial.Config{Name: "COM1"},
		conn:       mockPort,
		enable:     true,
		portStatus: false,
	}

	// 执行读取
	err := uart.UartRead(16, mockLog)

	// 验证结果
	assert.Error(t, err)
	assert.Equal(t, io.ErrUnexpectedEOF, err)
	assert.False(t, uart.portStatus)
	mockPort.AssertExpectations(t)
}

func TestUartWrite_Success(t *testing.T) {
	// 创建 mock 串口和日志
	mockPort := &MockSerialPort{}
	mockLog := &mockLogger{}

	// 设置 mock 行为：模拟写入 10 字节
	data := []byte("testdata12")
	mockPort.On("Write", data).Return(10, nil).Once()
	mockPort.On("Flush").Return(nil)

	// 创建 Uart 实例
	uart := &Uart{
		config:     &serial.Config{Name: "COM1"},
		conn:       mockPort,
		enable:     true,
		portStatus: false,
	}

	// 执行写入
	length, err := uart.UartWrite(data, mockLog)

	// 验证结果
	assert.NoError(t, err)
	assert.Equal(t, 10, length)
	mockPort.AssertExpectations(t)
}

func TestUartWrite_Error(t *testing.T) {
	// 创建 mock 串口和日志
	mockPort := &MockSerialPort{}
	mockLog := &mockLogger{}

	// 设置 mock 行为：模拟写入错误
	data := []byte("testdata12")
	mockPort.On("Write", data).Return(0, io.ErrUnexpectedEOF).Once()
	mockPort.On("Flush").Return(nil)

	// 创建 Uart 实例
	uart := &Uart{
		config:     &serial.Config{Name: "COM1"},
		conn:       mockPort,
		enable:     true,
		portStatus: false,
	}

	// 执行写入
	length, err := uart.UartWrite(data, mockLog)

	// 验证结果
	assert.Error(t, err)
	assert.Equal(t, io.ErrUnexpectedEOF, err)
	assert.Equal(t, 0, length)
	mockPort.AssertExpectations(t)
}

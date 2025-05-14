package driver

import (
	"github.com/stretchr/testify/mock"
)

// MockSerialPort 是 serial.Port 接口 伪串口接口的 mock 实现
type MockSerialPort struct {
	mock.Mock
}

// Read 模拟 Read 方法
func (m *MockSerialPort) Read(b []byte) (int, error) {
	args := m.Called(b)
	n := args.Int(0)
	err := args.Error(1)
	if n > 0 && len(b) >= n {
		// 模拟写入数据到 b
		for i := 0; i < n; i++ {
			b[i] = byte(i) // 填充假数据
		}
		return n, err
	}
	return n, err
}

// Write 模拟 Write 方法
func (m *MockSerialPort) Write(b []byte) (int, error) {
	args := m.Called(b)
	return args.Int(0), args.Error(1)
}

// Flush 模拟 Flush 方法
func (m *MockSerialPort) Flush() error {
	args := m.Called()
	return args.Error(0)
}

// Close 模拟 Close 方法（如果需要）
func (m *MockSerialPort) Close() error {
	args := m.Called()
	return args.Error(0)
}

package mocks

import (
	"time"
)

// MockSerialQueue 实现 interfaces.SerialQueue

type MockSerialQueue struct {
	SendCommandFunc func(command []byte, timeout time.Duration) (string, error)
	CloseFunc       func()
}

func (m *MockSerialQueue) SendCommand(command []byte, timeout time.Duration) (string, error) {
	if m.SendCommandFunc != nil {
		return m.SendCommandFunc(command, timeout)
	}
	return "", nil
}

func (m *MockSerialQueue) Close() {
	if m.CloseFunc != nil {
		m.CloseFunc()
	}
}

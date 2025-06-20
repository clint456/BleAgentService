package mocks

import "device-ble/internal/interfaces"

// MockConfigProvider 实现 interfaces.ConfigProvider

type MockConfigProvider struct {
	GetConfigFunc       func(key string) (interface{}, error)
	GetSerialConfigFunc func() interfaces.SerialConfig
	GetMQTTConfigFunc   func() interfaces.MQTTConfig
}

func (m *MockConfigProvider) GetConfig(key string) (interface{}, error) {
	if m.GetConfigFunc != nil {
		return m.GetConfigFunc(key)
	}
	return nil, nil
}

func (m *MockConfigProvider) GetSerialConfig() interfaces.SerialConfig {
	if m.GetSerialConfigFunc != nil {
		return m.GetSerialConfigFunc()
	}
	return interfaces.SerialConfig{}
}

func (m *MockConfigProvider) GetMQTTConfig() interfaces.MQTTConfig {
	if m.GetMQTTConfigFunc != nil {
		return m.GetMQTTConfigFunc()
	}
	return interfaces.MQTTConfig{}
}

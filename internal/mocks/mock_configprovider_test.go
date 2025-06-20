package mocks

import (
	"device-ble/internal/interfaces"
	"testing"
)

func TestMockConfigProvider_All(t *testing.T) {
	mock := &MockConfigProvider{
		GetConfigFunc: func(key string) (interface{}, error) {
			if key != "foo" {
				t.Errorf("unexpected key: %s", key)
			}
			return 42, nil
		},
		GetSerialConfigFunc: func() interfaces.SerialConfig {
			return interfaces.SerialConfig{PortName: "mock"}
		},
		GetMQTTConfigFunc: func() interfaces.MQTTConfig {
			return interfaces.MQTTConfig{Host: "mock"}
		},
	}
	if v, _ := mock.GetConfig("foo"); v != 42 {
		t.Error("GetConfig did not return expected value")
	}
	if mock.GetSerialConfig().PortName != "mock" {
		t.Error("GetSerialConfig did not return expected value")
	}
	if mock.GetMQTTConfig().Host != "mock" {
		t.Error("GetMQTTConfig did not return expected value")
	}
}

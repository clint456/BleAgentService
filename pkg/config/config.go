package config

import (
	"device-ble/internal/interfaces"
	"fmt"
)

// Config结构体
// Config将作为ConfigProvider接口的具体实现
type Config struct {
	Serial interfaces.SerialConfig `yaml:"serial"`
	MQTT   interfaces.MQTTConfig   `yaml:"mqtt"`
}

// 实现ConfigProvider的三个方法
func (c *Config) GetSerialConfig() interfaces.SerialConfig {
	return c.Serial
}

func (c *Config) GetMQTTConfig() interfaces.MQTTConfig {
	return c.MQTT
}

func (c *Config) GetConfig(key string) (interface{}, error) {
	switch key {
	case "serial":
		return c.Serial, nil
	case "mqtt":
		return c.MQTT, nil
	default:
		return nil, fmt.Errorf("unsupported config key: %s", key)
	}
}
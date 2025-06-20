package driver

import (
	"device-ble/internal/interfaces"
	"time"
)

// 配置接口
type ConfigProvider interface {
	GetSerialConfig() interfaces.SerialConfig
	GetMQTTConfig() interfaces.MQTTConfig
	GetConfig(key string) (interface{}, error)
}

// 具体实现
type Config struct {
	Serial interfaces.SerialConfig
	MQTT   interfaces.MQTTConfig
}

func (c *Config) GetSerialConfig() interfaces.SerialConfig { return c.Serial }
func (c *Config) GetMQTTConfig() interfaces.MQTTConfig     { return c.MQTT }

func (c *Config) GetConfig(key string) (interface{}, error) {
	return nil, nil // 可根据需要实现具体逻辑
}

// 串口配置结构体
type SerialConfig struct {
	PortName    string
	BaudRate    int
	ReadTimeout time.Duration
}

// MQTT 配置结构体
type MQTTConfig struct {
	Host     string
	Port     int
	Protocol string
	ClientID string
	QoS      int
	Username string
	Password string
}

package driver

import (
	"time"
)

// 配置接口
type ConfigProvider interface {
	GetSerialConfig() SerialConfig
	GetMQTTConfig() MQTTConfig
}

// 具体实现
type Config struct {
	Serial SerialConfig
	MQTT   MQTTConfig
}

func (c *Config) GetSerialConfig() SerialConfig { return c.Serial }
func (c *Config) GetMQTTConfig() MQTTConfig     { return c.MQTT }

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

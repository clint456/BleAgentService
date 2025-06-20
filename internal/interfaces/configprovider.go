package interfaces

// ConfigProvider 定义配置获取的通用接口，供 driver、mqttbus 等包依赖

type SerialConfig struct {
	PortName    string
	BaudRate    int
	ReadTimeout int // 或 time.Duration，视实现而定
}

type MQTTConfig struct {
	Host     string
	Port     int
	Protocol string
	ClientID string
	QoS      int
	Username string
	Password string
}

type ConfigProvider interface {
	GetConfig(key string) (interface{}, error)
	GetSerialConfig() SerialConfig
	GetMQTTConfig() MQTTConfig
}

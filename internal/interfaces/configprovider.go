package interfaces

// ConfigProvider 定义配置获取的通用接口，供 driver、mqttbus 等包依赖

type SerialConfig struct {
	PortName    string `yaml:"portName"`
	BaudRate    int    `yaml:"baudRate"`
	ReadTimeout int    `yaml:"readTimeout"` // 单位毫秒
}

type MQTTConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Protocol string `yaml:"protocol"`
	ClientID string `yaml:"clientID"`
	QoS      int    `yaml:"qos"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// 设计ConfigProvider接口
// 使用者需要实现其包含的三个方法
type ConfigProvider interface {
	GetConfig(key string) (interface{}, error)
	GetSerialConfig() SerialConfig
	GetMQTTConfig() MQTTConfig
}

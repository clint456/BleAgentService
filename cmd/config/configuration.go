package config

import (
	"errors"
	"fmt"
	"os"

	yaml "gopkg.in/yaml.v2"
)

// MQTTUserClientConfig 定义了MQTT配置结构体
type MQTTUserClientConfig struct {
	MqttUserConfig MQTTUserConfig `yaml:"MQTTUserClient"`
}

type MQTTUserConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Protocol string `yaml:"protocol"`
	ClientID string `yaml:"clientID"`
	QoS      int    `yaml:"qos"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// LoadConfig 从指定的文件加载配置
func LoadConfig(filePath string) (*MQTTUserClientConfig, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("unable to open config file: %v", err)
	}
	defer file.Close()

	var config MQTTUserClientConfig
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("unable to decode yaml into config: %v", err)
	}

	// 验证配置
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return &config, nil
}

// Validate 验证配置是否有效
func (config *MQTTUserClientConfig) Validate() error {
	if len(config.MqttUserConfig.Host) == 0 {
		return errors.New("MQTTUserClientConfig.Host cannot be empty")
	}
	if config.MqttUserConfig.Port <= 0 || config.MqttUserConfig.Port > 65535 {
		return errors.New("MQTTUserClientConfig.Port must be a valid port number between 1 and 65535")
	}
	if len(config.MqttUserConfig.Protocol) == 0 {
		return errors.New("MQTTUserClientConfig.Protocol cannot be empty")
	}
	if len(config.MqttUserConfig.ClientID) == 0 {
		return errors.New("MQTTUserClientConfig.ClientID cannot be empty")
	}
	if config.MqttUserConfig.QoS < 0 || config.MqttUserConfig.QoS > 2 {
		return errors.New("MQTTUserClientConfig.QoS must be 0, 1, or 2")
	}
	return nil
}

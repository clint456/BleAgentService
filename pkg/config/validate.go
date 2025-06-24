package config

import "fmt"

func validate(c *Config) error {
	if c.Serial.PortName == "" {
		return fmt.Errorf("串口 portName 不能为空")
	}
	if c.MQTT.Host == "" {
		return fmt.Errorf("MQTT host 未配置")
	}
	if c.MQTT.QoS < 0 || c.MQTT.QoS > 2 {
		return fmt.Errorf("MQTT QoS 取值范围应为 0~2")
	}
	return nil
}

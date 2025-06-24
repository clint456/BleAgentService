package config

func applyDefaults(c *Config) {
	if c.Serial.BaudRate == 0 {
		c.Serial.BaudRate = 115200
	}
	if c.Serial.ReadTimeout == 0 {
		c.Serial.ReadTimeout = 100
	}
	if c.MQTT.Protocol == "" {
		c.MQTT.Protocol = "tcp"
	}
	if c.MQTT.Port == 0 {
		c.MQTT.Port = 1883
	}
}

package config

import "os"

func applyEnvOverrides(c *Config) {
	if v := os.Getenv("MQTT_HOST"); v != "" {
		c.MQTT.Host = v
	}
	if v := os.Getenv("MQTT_USERNAME"); v != "" {
		c.MQTT.Username = v
	}
	if v := os.Getenv("SERIAL_PORT"); v != "" {
		c.Serial.PortName = v
	}
}

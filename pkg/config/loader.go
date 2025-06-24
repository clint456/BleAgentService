package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取配置失败: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("YAML解析失败: %w", err)
	}

	applyDefaults(&cfg)
	applyEnvOverrides(&cfg)

	if err := validate(&cfg); err != nil {
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}

	return &cfg, nil
}

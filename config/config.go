package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	PlayerRateLimit int    `yaml:"playerRateLimit"`
	WebSocketPort   int    `yaml:"websocketPort"`
	GameBinary      string `yaml:"gameBinary"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal yaml: %w", err)
	}

	return &cfg, nil
}

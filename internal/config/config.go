package config

import (
	"errors"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Name         string        `yaml:"name"`
	Env          []string      `yaml:"env"`
	ProtectedRPC []string      `yaml:"protected_rpc"`
	JwtExp       time.Duration `yaml:"jwt_exp"`
	SessionExp   time.Duration `yaml:"session_exp"`
}

func LoadConfig() (*Config, error) {
	f, err := os.ReadFile("config/dev.yaml")
	if err != nil {
		return nil, errors.New("failed to read config file from root config folder")
	}

	cfg := &Config{}
	err = yaml.Unmarshal(f, &cfg)
	if err != nil {
		return nil, errors.New("failed to unmarshal config file")
	}

	return cfg, nil
}

package config

import (
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	App  AppConfig  `yaml:"app"`
	HTTP HTTPConfig `yaml:"http"`
	Log  LogConfig  `yaml:"logger"`
	JWT  JWTConfig  `yaml:"jwt"`
}

type AppConfig struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}

type HTTPConfig struct {
	Port string `yaml:"port"`
}

type LogConfig struct {
	Level      string `yaml:"log_level"`
	RollbarEnv string `yaml:"rollbar_env"`
}

type JWTConfig struct {
	SecretKey     string        `yaml:"secret_key"`
	AccessExpiry  time.Duration `yaml:"access_expiry"`
	RefreshExpiry time.Duration `yaml:"refresh_expiry"`
}

func NewConfig() (*Config, error) {
	cfg := &Config{}
	err := cleanenv.ReadConfig("config/config.yaml", cfg)
	if err != nil {
		return nil, err
	}

	// Set default JWT values if not provided
	if cfg.JWT.AccessExpiry == 0 {
		cfg.JWT.AccessExpiry = 15 * time.Minute
	}
	if cfg.JWT.RefreshExpiry == 0 {
		cfg.JWT.RefreshExpiry = 7 * 24 * time.Hour // 7 dias
	}

	return cfg, nil
}

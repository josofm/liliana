package config

import (
	"fmt"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	App  AppConfig  `yaml:"app"`
	HTTP HTTPConfig `yaml:"http"`
	Log  LogConfig  `yaml:"logger"`
	JWT  JWTConfig  `yaml:"jwt"`
	DB   DBConfig   `yaml:"database"`
}

type AppConfig struct {
	Name        string `yaml:"name" env:"APP_NAME"`
	Version     string `yaml:"version" env:"APP_VERSION"`
	Environment string `yaml:"environment" env:"APP_ENV"`
}

type HTTPConfig struct {
	Port string `yaml:"port" env:"HTTP_PORT"`
}

type LogConfig struct {
	Level      string `yaml:"log_level" env:"LOG_LEVEL"`
	RollbarEnv string `yaml:"rollbar_env" env:"ROLLBAR_ENV"`
}

type JWTConfig struct {
	SecretKey     string        `yaml:"secret_key" env:"JWT_SECRET_KEY"`
	AccessExpiry  time.Duration `yaml:"access_expiry" env:"JWT_ACCESS_EXPIRY"`
	RefreshExpiry time.Duration `yaml:"refresh_expiry" env:"JWT_REFRESH_EXPIRY"`
}

type DBConfig struct {
	URL         string `yaml:"url" env:"DATABASE_URL"`
	AutoMigrate bool   `yaml:"auto_migrate" env:"AUTO_MIGRATE"`
}

func NewConfig() (*Config, error) {
	cfg := &Config{}
	if path, ok := configPath(); ok {
		err := cleanenv.ReadConfig(path, cfg)
		if err != nil {
			return nil, err
		}
	}
	if err := cleanenv.ReadEnv(cfg); err != nil {
		return nil, err
	}

	// Set default JWT values if not provided
	if cfg.App.Environment == "" {
		cfg.App.Environment = "development"
	}
	if cfg.HTTP.Port == "" {
		cfg.HTTP.Port = "8080"
	}
	if port := os.Getenv("PORT"); port != "" && os.Getenv("HTTP_PORT") == "" {
		cfg.HTTP.Port = port
	}
	if cfg.Log.Level == "" {
		cfg.Log.Level = "info"
	}
	if cfg.JWT.AccessExpiry == 0 {
		cfg.JWT.AccessExpiry = 15 * time.Minute
	}
	if cfg.JWT.RefreshExpiry == 0 {
		cfg.JWT.RefreshExpiry = 7 * 24 * time.Hour // 7 dias
	}
	if databaseURL := os.Getenv("DATABASE_URL"); databaseURL != "" {
		cfg.DB.URL = databaseURL
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func configPath() (string, bool) {
	if _, err := os.Stat("config/config.yaml"); err == nil {
		return "config/config.yaml", true
	}
	if _, err := os.Stat("config.yaml"); err == nil {
		return "config.yaml", true
	}

	return "", false
}

func (c *Config) Validate() error {
	if c.App.Environment != "production" {
		return nil
	}

	if c.DB.URL == "" {
		return fmt.Errorf("DATABASE_URL is required in production")
	}
	if c.JWT.SecretKey == "" || c.JWT.SecretKey == "dev-secret-change-me" || c.JWT.SecretKey == "your-super-secret-jwt-key-change-in-production" {
		return fmt.Errorf("JWT_SECRET_KEY must be set to a production secret")
	}

	return nil
}

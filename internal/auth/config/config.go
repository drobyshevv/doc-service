// Package config отвечает за загрузку,
// валидацию и предоставление конфигурации auth-сервиса.
package config

import (
	"fmt"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

// Config содержит конфигурацию auth-сервиса.
type Config struct {
	HTTP  HTTPConfig  `yaml:"http"`
	DB    DBConfig    `yaml:"database"`
	Redis RedisConfig `yaml:"redis"`
	JWT   JWTConfig   `yaml:"jwt"`
}

// HTTPConfig содержит настройки HTTP-сервера.
type HTTPConfig struct {
	Port int `yaml:"port" env:"HTTP_PORT" env-default:"8082"`
}

// DBConfig содержит настройки подключения к PostgreSQL.
type DBConfig struct {
	Host     string `yaml:"host" env:"DB_HOST" env-default:"localhost"`
	Port     int    `yaml:"port" env:"DB_PORT" env-default:"5432"`
	User     string `yaml:"user" env:"DB_USER" env-default:"postgres"`
	Password string `yaml:"password" env:"DB_PASSWORD"`
	Name     string `yaml:"name" env:"DB_NAME"`
	SSLMode  string `yaml:"sslmode" env:"DB_SSLMODE" env-default:"disable"`
}

// RedisConfig содержит настройки Redis.
type RedisConfig struct {
	Host     string `yaml:"host" env:"REDIS_HOST" env-default:"localhost"`
	Port     int    `yaml:"port" env:"REDIS_PORT" env-default:"6379"`
	Password string `yaml:"password" env:"REDIS_PASSWORD"`
	DB       int    `yaml:"db" env:"REDIS_DB" env-default:"1"`
}

// JWTConfig содержит настройки JWT токенов.
type JWTConfig struct {
	AccessSecret string `yaml:"access_secret" env:"JWT_ACCESS_SECRET"`
	//RefreshSecret    string `yaml:"refresh_secret" env:"JWT_REFRESH_SECRET"`
	AccessTTLMinutes int `yaml:"access_ttl_minutes" env:"JWT_ACCESS_TTL_MINUTES" env-default:"15"`
	RefreshTTLDays   int `yaml:"refresh_ttl_days" env:"JWT_REFRESH_TTL_DAYS" env-default:"30"`
}

// LoadConfig загружает конфигурацию из YAML,
// применяет env-переменные и валидирует.
func LoadConfig(path string) (*Config, error) {
	var cfg Config

	if _, err := os.Stat(path); err != nil {
		return nil, fmt.Errorf("config file not found: %w", err)
	}

	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	if err := validate(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// DBConnStr возвращает строку подключения к PostgreSQL.
func (c *Config) DBConnStr() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.DB.User,
		c.DB.Password,
		c.DB.Host,
		c.DB.Port,
		c.DB.Name,
		c.DB.SSLMode,
	)
}

// RedisAddr возвращает адрес Redis.
func (c *Config) RedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Redis.Host, c.Redis.Port)
}

// HTTPAddr возвращает адрес HTTP сервера.
func (c *Config) HTTPAddr() string {
	return fmt.Sprintf(":%d", c.HTTP.Port)
}

func validate(cfg *Config) error {
	if cfg.HTTP.Port <= 0 {
		return fmt.Errorf("invalid http port")
	}

	if cfg.DB.Name == "" {
		return fmt.Errorf("db name is required")
	}

	if cfg.DB.User == "" {
		return fmt.Errorf("db user is required")
	}

	if cfg.JWT.AccessSecret == "" {
		return fmt.Errorf("jwt access_secret is required")
	}

	/*
		if cfg.JWT.RefreshSecret == "" {
			return fmt.Errorf("jwt refresh_secret is required")
		}
	*/

	if cfg.JWT.AccessTTLMinutes <= 0 {
		return fmt.Errorf("jwt access_ttl_minutes must be > 0")
	}

	if cfg.JWT.RefreshTTLDays <= 0 {
		return fmt.Errorf("jwt refresh_ttl_days must be > 0")
	}

	return nil
}

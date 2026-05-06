// Package config отвечает за загрузку,
// валидацию и предоставление конфигурации приложения.
package config

import (
	"fmt"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

// Config содержит конфигурацию всех модулей приложения.
type Config struct {
	HTTP   HTTPConfig   `yaml:"http"`
	DB     DBConfig     `yaml:"database"`
	Redis  RedisConfig  `yaml:"redis"`
	S3     S3Config     `yaml:"s3"`
	Search SearchConfig `yaml:"search"`
}

// HTTPConfig содержит настройки HTTP-сервера.
type HTTPConfig struct {
	Port int `yaml:"port" env:"HTTP_PORT" env-default:"8081"`
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

// RedisConfig содержит настройки подключения к Redis.
type RedisConfig struct {
	Host     string `yaml:"host" env:"REDIS_HOST" env-default:"localhost"`
	Port     int    `yaml:"port" env:"REDIS_PORT" env-default:"6379"`
	Password string `yaml:"password" env:"REDIS_PASSWORD"`
	DB       int    `yaml:"db" env:"REDIS_DB" env-default:"0"`
}

// S3Config содержит настройки S3-совместимого хранилища.
type S3Config struct {
	Endpoint  string `yaml:"endpoint" env:"S3_ENDPOINT"`
	AccessKey string `yaml:"access_key" env:"S3_ACCESS_KEY"`
	SecretKey string `yaml:"secret_key" env:"S3_SECRET_KEY"`
	Bucket    string `yaml:"bucket" env:"S3_BUCKET"`
	UseSSL    bool   `yaml:"use_ssl" env:"S3_USE_SSL" env-default:"false"`
}

// SearchConfig содержит настройки поискового движка.
type SearchConfig struct {
	SnippetLength int `yaml:"snippet_length" env:"SEARCH_SNIPPET_LENGTH" env-default:"200"`
	MaxResults    int `yaml:"max_results" env:"SEARCH_MAX_RESULTS" env-default:"20"`
	MinTermLength int `yaml:"min_term_length" env:"SEARCH_MIN_TERM_LENGTH" env-default:"2"`
}

// LoadConfig загружает конфигурацию из YAML-файла,
// применяет переменные окружения и валидирует результат.
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

// DBConnStr возвращает PostgreSQL connection string.
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

// RedisAddr возвращает адрес Redis в формате host:port.
func (c *Config) RedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Redis.Host, c.Redis.Port)
}

// HTTPAddr возвращает адрес HTTP-сервера.
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

	if cfg.S3.Endpoint == "" {
		return fmt.Errorf("s3 endpoint is required")
	}

	if cfg.S3.Bucket == "" {
		return fmt.Errorf("s3 bucket is required")
	}

	if cfg.Search.MaxResults <= 0 {
		return fmt.Errorf("search max_results must be greater than 0")
	}

	if cfg.Search.MinTermLength <= 0 {
		return fmt.Errorf("search min_term_length must be greater than 0")
	}

	return nil
}

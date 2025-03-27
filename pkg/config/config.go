// Package config provides functionality for loading and accessing application configuration.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

// Config represents the application configuration loaded from environment variables.
type Config struct {
	HTTP     HTTP
	PG       PG
	Redis    Redis
	RabbitMQ RabbitMQ
	JWT      JWT
	Log      Log
}

// HTTP holds the HTTP server related configuration.
type HTTP struct {
	Host         string        `env:"HTTP_HOST" envDefault:"0.0.0.0"`
	Port         int           `env:"HTTP_PORT" envDefault:"8080"`
	Mode         string        `env:"HTTP_MODE" envDefault:"release"`
	ReadTimeout  time.Duration `env:"HTTP_READ_TIMEOUT" envDefault:"15s"`
	WriteTimeout time.Duration `env:"HTTP_WRITE_TIMEOUT" envDefault:"15s"`
	IdleTimeout  time.Duration `env:"HTTP_IDLE_TIMEOUT" envDefault:"60s"`
}

// PG holds the PostgreSQL database configuration.
type PG struct {
	URL             string        `env:"PG_URL" envDefault:"postgres://postgres:postgres@localhost:5432/go_layout?sslmode=disable"`
	MaxOpenConns    int           `env:"PG_MAX_OPEN_CONNS" envDefault:"25"`
	MaxIdleConns    int           `env:"PG_MAX_IDLE_CONNS" envDefault:"5"`
	ConnMaxLifetime time.Duration `env:"PG_CONN_MAX_LIFETIME" envDefault:"5m"`
}

// Redis holds the Redis configuration.
type Redis struct {
	URL      string `env:"REDIS_URL" envDefault:"redis://:@localhost:6379/0"`
	PoolSize int    `env:"REDIS_POOL_SIZE" envDefault:"10"`
}

// RabbitMQ holds the RabbitMQ configuration.
type RabbitMQ struct {
	URL      string `env:"RABBITMQ_URL" envDefault:"amqp://guest:guest@localhost:5672/"`
	Exchange string `env:"RABBITMQ_EXCHANGE" envDefault:"go-layout"`
}

// JWT holds the JWT related configuration.
type JWT struct {
	Secret     string        `env:"JWT_SECRET,required"`
	Expiration time.Duration `env:"JWT_EXPIRATION" envDefault:"24h"`
}

// Log holds the logging related configuration.
type Log struct {
	Level  string `env:"LOG_LEVEL" envDefault:"info"`
	Format string `env:"LOG_FORMAT" envDefault:"json"` // Available options: json, console
}

// Load loads the configuration from environment variables.
func Load() (*Config, error) {
	// Load environment variables from file if in development
	loadEnvFile()

	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("config error: %w", err)
	}

	return cfg, nil
}

// EnvFileLoader loads environment variables from a file if the file exists.
// This is primarily used for development environments.
func loadEnvFile() {
	// Check for environment indication
	environment := os.Getenv("GO_ENV")
	if environment == "" {
		environment = "dev" // Default to development environment
	}

	// For production/docker environments, rely on system environment variables
	if environment == "prod" || environment == "production" {
		return
	}

	// Try to find .env.local or .env file in config directory
	basePaths := []string{
		"config", // When running from project root
		".." + string(os.PathSeparator) + "config",                                   // When running from cmd/api
		".." + string(os.PathSeparator) + ".." + string(os.PathSeparator) + "config", // When running from a subdirectory
	}

	// First try to load .env.local files (prioritized for local development)
	for _, basePath := range basePaths {
		localEnvPath := filepath.Join(basePath, environment, ".env.local")
		if _, err := os.Stat(localEnvPath); err == nil {
			// Load .env.local file using godotenv
			if err := godotenv.Load(localEnvPath); err == nil {
				return // Successfully loaded .env.local
			}
		}
	}

	// If no .env.local file was found, try to load .env files
	for _, basePath := range basePaths {
		defaultEnvPath := filepath.Join(basePath, environment, ".env")
		if _, err := os.Stat(defaultEnvPath); err == nil {
			// Load .env file using godotenv
			if err := godotenv.Load(defaultEnvPath); err == nil {
				return // Successfully loaded .env
			}
		}
	}
}

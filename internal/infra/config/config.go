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
	// RequestTimeout is the application-level timeout for request processing
	RequestTimeout time.Duration `env:"HTTP_REQUEST_TIMEOUT" envDefault:"10s"`
	// AllowOrigins is a comma-separated list of origins allowed for CORS.
	// Set to "*" to allow all origins (not recommended for production).
	// Example: "https://example.com,https://api.example.com"
	AllowOrigins []string `env:"HTTP_ALLOW_ORIGINS,required"`
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
	// Secret key used to sign JWT tokens
	Secret string `env:"JWT_SECRET,required"`

	// TokenExpiration is the duration a token is valid (access token lifetime)
	// After this time, the token will be considered expired and will no longer be valid
	// Recommended: short duration (15-30 minutes) for security
	TokenExpiration time.Duration `env:"JWT_TOKEN_EXPIRATION" envDefault:"30m"`

	// RefreshExpiration is the maximum duration during which a token can be refreshed
	// Even if token has expired, user won't need to re-login if within this window
	// Recommended: longer duration (1-7 days) for good UX
	RefreshExpiration time.Duration `env:"JWT_REFRESH_EXPIRATION" envDefault:"168h"` // 7 days
}

// Log holds the logging related configuration.
type Log struct {
	Development bool   `env:"LOG_DEVELOPMENT" envDefault:"false"` // Enable development mode
	Level       string `env:"LOG_LEVEL" envDefault:"info"`        // Available options: debug, info, warn, error
	OutputFile  string `env:"LOG_OUTPUT_FILE"`                    // Log file path, empty for console output only
	MaxSize     int    `env:"LOG_MAX_SIZE" envDefault:"100"`      // Maximum size of log files in MB before rotation
	MaxAge      int    `env:"LOG_MAX_AGE" envDefault:"7"`         // Maximum number of days to retain old log files
	MaxBackups  int    `env:"LOG_MAX_BACKUPS" envDefault:"5"`     // Maximum number of old log files to retain
	Compress    bool   `env:"LOG_COMPRESS" envDefault:"true"`     // Compress rotated files with gzip
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
		".." + string(os.PathSeparator) + "config",                                   // When running from cmd/xxx
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

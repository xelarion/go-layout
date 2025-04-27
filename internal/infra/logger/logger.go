// Package logger provides logging functionality for the application.
package logger

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/xelarion/go-layout/internal/infra/config"
)

// Logger wraps zap.Logger to provide a consistent logging interface.
type Logger struct {
	*zap.Logger
}

// New creates a logger instance from the provided configuration.
func New(cfg *config.Log) (*Logger, error) {
	// Register lumberjack sink if file output is configured
	if err := RegisterLumberjackSink(cfg); err != nil {
		return nil, err
	}

	// Parse log level
	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		level = zapcore.InfoLevel // Default to info if parsing fails
	}

	// Configure logger based on environment
	var zapCfg zap.Config
	if cfg.Development {
		zapCfg = zap.NewDevelopmentConfig()
		if cfg.OutputFile == "" {
			zapCfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		}
	} else {
		zapCfg = zap.NewProductionConfig()
		zapCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	// Apply configuration
	zapCfg.Level = zap.NewAtomicLevelAt(level)

	// Configure output paths
	if cfg.OutputFile != "" {
		fileURL := fmt.Sprintf("%s://%s", lumberjackScheme, cfg.OutputFile)
		zapCfg.OutputPaths = []string{fileURL}
		zapCfg.ErrorOutputPaths = []string{fileURL}
	} else {
		zapCfg.OutputPaths = []string{"stdout"}
		zapCfg.ErrorOutputPaths = []string{"stderr"}
	}

	// Build the logger
	logger, err := zapCfg.Build(
		zap.AddCallerSkip(1), // Skip the wrapper logger calls
	)
	if err != nil {
		return nil, fmt.Errorf("failed to build logger: %w", err)
	}

	return &Logger{logger}, nil
}

// With creates a child logger with the provided fields.
func (l *Logger) With(fields ...zapcore.Field) *Logger {
	return &Logger{l.Logger.With(fields...)}
}

// Named creates a logger with the provided name.
func (l *Logger) Named(name string) *Logger {
	return &Logger{l.Logger.Named(name)}
}

// Sugar returns a sugared logger.
func (l *Logger) Sugar() *zap.SugaredLogger {
	return l.Logger.Sugar()
}

// Sync flushes any buffered log entries.
func (l *Logger) Sync() error {
	return l.Logger.Sync()
}

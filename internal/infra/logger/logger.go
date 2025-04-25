// Package logger provides logging functionality for the application.
package logger

import (
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
	// Parse log level
	var level zapcore.Level
	switch cfg.Level {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	default:
		level = zapcore.InfoLevel
	}

	// Choose encoder based on format
	var zapCfg zap.Config
	if cfg.Format == "json" {
		zapCfg = zap.NewProductionConfig()
		zapCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		// Production mode - disable caller for performance
		zapCfg.DisableCaller = true
	} else {
		zapCfg = zap.NewDevelopmentConfig()
		zapCfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		// Development mode - enable caller for better debugging
		zapCfg.DisableCaller = false
	}

	zapCfg.Level = zap.NewAtomicLevelAt(level)
	zapCfg.OutputPaths = []string{"stdout"}
	zapCfg.ErrorOutputPaths = []string{"stderr"}

	// In json format, caller adds overhead, so we control it specifically above
	// But stacktrace should be disabled by default for both formats
	zapCfg.DisableStacktrace = true

	// Configure log sampling to reduce log volume in production environments
	if cfg.EnableSampling {
		zapCfg.Sampling = &zap.SamplingConfig{
			Initial:    cfg.SamplingInitial,
			Thereafter: cfg.SamplingAfter,
		}
	}

	logger, err := zapCfg.Build()
	if err != nil {
		return nil, err
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

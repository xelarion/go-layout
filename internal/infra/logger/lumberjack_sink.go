package logger

import (
	"fmt"
	"net/url"
	"sync"

	"go.uber.org/zap"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/xelarion/go-layout/internal/infra/config"
)

// Custom sink scheme for log rotation
const lumberjackScheme = "lumberjack"

// lumberjackSinkRegisterOnce ensures that the sink is only registered once
var lumberjackSinkRegisterOnce sync.Once

// lumberjackSink wraps lumberjack to implement zap.Sink interface
type lumberjackSink struct {
	*lumberjack.Logger
}

// Close method required by zap.Sink interface
func (s *lumberjackSink) Close() error {
	return nil
}

// Sync implements zap.Sink
func (s *lumberjackSink) Sync() error {
	return s.Logger.Close()
}

// RegisterLumberjackSink registers the lumberjack sink for file-based log rotation.
// It is safe to call this function multiple times as the actual registration
// will only happen once thanks to sync.Once.
func RegisterLumberjackSink(cfg *config.Log) error {
	if cfg.OutputFile == "" {
		return nil // No file output configured
	}

	var registerErr error
	lumberjackSinkRegisterOnce.Do(func() {
		registerErr = zap.RegisterSink(lumberjackScheme, func(*url.URL) (zap.Sink, error) {
			return &lumberjackSink{
				Logger: &lumberjack.Logger{
					Filename:   cfg.OutputFile,
					MaxSize:    cfg.MaxSize,
					MaxBackups: cfg.MaxBackups,
					MaxAge:     cfg.MaxAge,
					Compress:   cfg.Compress,
				},
			}, nil
		})

		if registerErr != nil {
			registerErr = fmt.Errorf("failed to register lumberjack sink: %w", registerErr)
		}
	})

	return registerErr
}

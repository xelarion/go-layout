// Package main provides a command line program for running various task types.
package main

import (
	"context"
	"log"

	"go.uber.org/zap"

	"github.com/xelarion/go-layout/internal/infra/config"
	"github.com/xelarion/go-layout/internal/infra/logger"
	"github.com/xelarion/go-layout/pkg/app"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Version information set by build flags
	Version = "0.0.1"
)

// newApp creates a new application instance with specified configuration
func newApp(logger *zap.Logger, ts *taskServer) *app.App {
	return app.New(
		logger,
		app.WithID("task"),
		app.WithName("Task Service"),
		app.WithVersion(Version),
		app.WithServers(ts),
		app.WithAfterStart(func(ctx context.Context) error {
			logger.Info("Task service started",
				zap.String("version", Version))
			return nil
		}),
	)
}

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Panicf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	zapLogger, err := logger.New(&cfg.Log)
	if err != nil {
		log.Panicf("Failed to initialize logger: %v", err)
	}
	defer zapLogger.Sync()

	// Initialize the application with all features enabled
	a, cleanup, err := initApp(&cfg.PG, &cfg.Redis, &cfg.RabbitMQ, zapLogger.Logger)
	if err != nil {
		zapLogger.Panic("Failed to initialize application", zap.Error(err))
	}

	// Ensure cleanup is called even if app fails to start
	defer cleanup()

	// Start the application
	if err := a.Run(); err != nil {
		zapLogger.Panic("Error during application lifecycle", zap.Error(err))
	}
}

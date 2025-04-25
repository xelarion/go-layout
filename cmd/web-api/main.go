package main

import (
	"context"
	"log"

	"go.uber.org/zap"

	"github.com/xelarion/go-layout/internal/infra/config"
	"github.com/xelarion/go-layout/internal/infra/logger"
	httpServer "github.com/xelarion/go-layout/internal/infra/server/http"
	"github.com/xelarion/go-layout/pkg/app"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Version information set by build flags
	Version = "0.0.1"
)

// newApp creates a new application instance with specified configuration
func newApp(logger *zap.Logger, hs *httpServer.Server) *app.App {
	return app.New(
		logger,
		app.WithID("web-api"),
		app.WithName("Web API Service"),
		app.WithVersion(Version),
		app.WithServers(hs),
		app.WithAfterStart(func(ctx context.Context) error {
			logger.Info("Web API service started",
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

	// Initialize the application
	a, cleanup, err := initApp(&cfg.PG, &cfg.Redis, &cfg.RabbitMQ, &cfg.HTTP, &cfg.JWT, zapLogger.Logger)
	if err != nil {
		zapLogger.Panic("Failed to initialize application", zap.Error(err))
	}

	// Ensure cleanup is called even if app fails to start
	defer cleanup()

	// Start the application
	if err := a.Run(); err != nil {
		zapLogger.Panic("Error during Web API application lifecycle", zap.Error(err))
	}
}

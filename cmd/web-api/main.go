package main

import (
	"context"
	"log"
	"os"
	"time"

	"go.uber.org/zap"

	"github.com/xelarion/go-layout/pkg/app"
	"github.com/xelarion/go-layout/pkg/config"
	"github.com/xelarion/go-layout/pkg/logger"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		// Use standard log before we have a proper logger
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	zapLogger, err := logger.New(&cfg.Log)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer zapLogger.Sync()

	// Initialize the application
	apiApp, err := initApp(cfg, zapLogger.Logger)
	if err != nil {
		zapLogger.Logger.Fatal("Failed to initialize Web API application", zap.Error(err))
	}

	// Start the application with signal handling
	ctx := context.Background()
	if err := app.RunWithSignalHandling(ctx, apiApp, 10*time.Second); err != nil {
		zapLogger.Logger.Error("Error during Web API application lifecycle", zap.Error(err))
		os.Exit(1)
	}
}

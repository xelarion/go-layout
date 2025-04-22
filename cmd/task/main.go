// Package main provides a command line program for running various task types.
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"time"

	"go.uber.org/zap"

	"github.com/xelarion/go-layout/internal/config"
	"github.com/xelarion/go-layout/internal/logger"
	"github.com/xelarion/go-layout/pkg/app"
)

// Command line flags for controlling which task types are enabled
var (
	enableScheduler = flag.Bool("scheduler", true, "Enable scheduler tasks")
	enablePoller    = flag.Bool("poller", true, "Enable poller tasks")
	enableQueue     = flag.Bool("queue", true, "Enable queue processing")
)

func main() {
	flag.Parse()

	// Initialize configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	zapLogger, err := logger.New(&cfg.Log)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer zapLogger.Sync()

	// Initialize the application
	taskApp, err := initApp(cfg, zapLogger.Logger, enableScheduler, enablePoller, enableQueue)
	if err != nil {
		zapLogger.Logger.Fatal("Failed to initialize application", zap.Error(err))
	}

	// Start the application with signal handling
	ctx := context.Background()
	if err := app.RunWithSignalHandling(ctx, taskApp, 10*time.Second); err != nil {
		zapLogger.Logger.Error("Error during application lifecycle", zap.Error(err))
		os.Exit(1)
	}
}

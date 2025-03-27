package server

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
)

// Server defines the interface for any server type.
type Server interface {
	Start() error
	Shutdown(ctx context.Context) error
}

// GracefulShutdown handles graceful shutdown for multiple servers.
func GracefulShutdown(ctx context.Context, timeout time.Duration, logger *zap.Logger, servers ...Server) {
	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// Block until a signal is received
	<-quit
	logger.Info("Shutting down servers...")

	// Create a context with timeout for shutdown
	shutdownCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Create a channel to signal when all servers are shut down
	done := make(chan struct{})

	go func() {
		for _, srv := range servers {
			if err := srv.Shutdown(shutdownCtx); err != nil {
				logger.Error("Server shutdown error", zap.Error(err))
			}
		}
		close(done)
	}()

	// Wait for either the shutdown to complete or the context to timeout
	select {
	case <-shutdownCtx.Done():
		if errors.Is(shutdownCtx.Err(), context.DeadlineExceeded) {
			logger.Fatal("Graceful shutdown timed out.. forcing exit.")
		}
	case <-done:
		logger.Info("All servers gracefully stopped")
	}
}

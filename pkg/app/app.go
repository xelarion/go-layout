// Package app provides application lifecycle management utilities.
package app

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"go.uber.org/zap"
)

// AppInfo defines the interface that app implementations must satisfy.
type AppInfo interface {
	// ID returns the unique identifier of the application.
	ID() string

	// Name returns the name of the application.
	Name() string

	// Version returns the version of the application.
	Version() string

	// Start starts the application and blocks until the context is done or an error occurs.
	Start(ctx context.Context) error

	// Stop stops the application gracefully.
	Stop(ctx context.Context) error
}

// Option is a function that configures an App
type Option func(*App)

// App represents a running application instance.
type App struct {
	id      string
	name    string
	version string
	logger  *zap.Logger

	// StartFunc function that will be called when the application is started
	StartFunc func(ctx context.Context) error

	// StopFunc function that will be called when the application is stopped
	StopFunc func(ctx context.Context) error
}

// WithStartFunc sets the start function for the application
func WithStartFunc(fn func(ctx context.Context) error) Option {
	return func(a *App) {
		a.StartFunc = fn
	}
}

// WithStopFunc sets the stop function for the application
func WithStopFunc(fn func(ctx context.Context) error) Option {
	return func(a *App) {
		a.StopFunc = fn
	}
}

// NewApp creates a new application instance with the provided options.
func NewApp(id, name, version string, logger *zap.Logger, opts ...Option) *App {
	app := &App{
		id:      id,
		name:    name,
		version: version,
		logger:  logger,
		// Default implementations do nothing
		StartFunc: func(ctx context.Context) error { return nil },
		StopFunc:  func(ctx context.Context) error { return nil },
	}

	// Apply options
	for _, opt := range opts {
		opt(app)
	}

	return app
}

// ID returns the application identifier.
func (a *App) ID() string {
	return a.id
}

// Name returns the application name.
func (a *App) Name() string {
	return a.name
}

// Version returns the application version.
func (a *App) Version() string {
	return a.version
}

// Start starts the application and blocks until the context is done or an error occurs.
func (a *App) Start(ctx context.Context) error {
	a.logger.Info("Starting application",
		zap.String("id", a.id),
		zap.String("name", a.name),
		zap.String("version", a.version))

	return a.StartFunc(ctx)
}

// Stop stops the application gracefully.
func (a *App) Stop(ctx context.Context) error {
	a.logger.Info("Stopping application", zap.String("id", a.id))
	return a.StopFunc(ctx)
}

// RunWithSignalHandling starts the application and handles OS signals for graceful shutdown.
// It will block until the application exits or a signal is received.
func RunWithSignalHandling(ctx context.Context, app AppInfo, shutdownTimeout time.Duration) error {
	// Create a cancellable context
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Channel to receive application errors
	errCh := make(chan error, 1)

	// Channel to receive OS signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Start the application in a goroutine
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		errCh <- app.Start(ctx)
	}()

	// Wait for application to exit or signal
	select {
	case err := <-errCh:
		// Application exited on its own
		return err
	case sig := <-sigCh:
		// Signal received, initiate shutdown
		logger := zap.L()
		logger.Info("Signal received, shutting down", zap.String("signal", sig.String()))

		// Create a context with timeout for shutdown
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer shutdownCancel()

		// Stop the application
		err := app.Stop(shutdownCtx)
		if err != nil {
			logger.Error("Error during shutdown", zap.Error(err))
		}

		// Wait for the application to exit
		select {
		case appErr := <-errCh:
			return appErr
		case <-shutdownCtx.Done():
			logger.Error("Shutdown timed out")
			return shutdownCtx.Err()
		}
	}
}

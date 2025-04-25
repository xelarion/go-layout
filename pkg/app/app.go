// Package app provides application lifecycle management utilities.
package app

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

// Server represents a server component that can be started and stopped.
type Server interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

// AppInfo is application context value.
type AppInfo interface {
	ID() string
	Name() string
	Version() string
	Metadata() map[string]string
}

// App is an application components lifecycle manager.
type App struct {
	opts   options
	ctx    context.Context
	cancel context.CancelFunc
	logger *zap.Logger
}

// New creates a new application lifecycle manager.
func New(logger *zap.Logger, opts ...Option) *App {
	o := options{
		ctx:         context.Background(),
		sigs:        []os.Signal{syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT},
		stopTimeout: 10 * time.Second,
		metadata:    make(map[string]string),
	}

	// Generate a unique ID if not provided
	if id, err := uuid.NewUUID(); err == nil {
		o.id = id.String()
	}

	// Apply options
	for _, opt := range opts {
		opt(&o)
	}

	ctx, cancel := context.WithCancel(o.ctx)
	return &App{
		ctx:    ctx,
		cancel: cancel,
		opts:   o,
		logger: logger,
	}
}

// ID returns the app instance ID.
func (a *App) ID() string {
	return a.opts.id
}

// Name returns the app name.
func (a *App) Name() string {
	return a.opts.name
}

// Version returns the app version.
func (a *App) Version() string {
	return a.opts.version
}

// Metadata returns the app metadata.
func (a *App) Metadata() map[string]string {
	return a.opts.metadata
}

// Run executes all registered servers and blocks until interrupted or error.
func (a *App) Run() error {
	// Create application context
	appCtx := NewContext(a.ctx, a)

	// Execute beforeStart hooks
	for _, fn := range a.opts.beforeStart {
		if err := fn(appCtx); err != nil {
			return err
		}
	}

	// Create error group for managing goroutines
	eg, ctx := errgroup.WithContext(appCtx)
	wg := sync.WaitGroup{}

	// Start all servers
	for _, srv := range a.opts.servers {
		server := srv
		eg.Go(func() error {
			<-ctx.Done() // Wait for stop signal
			stopCtx := appCtx
			stopCtx, cancel := context.WithTimeout(NewContext(a.opts.ctx, a), a.opts.stopTimeout)
			defer cancel()
			return server.Stop(stopCtx)
		})

		wg.Add(1)
		eg.Go(func() error {
			wg.Done() // here is to ensure server start has begun running, so defer is not needed
			return server.Start(ctx)
		})
	}

	// Wait for all servers to start
	wg.Wait()

	// Execute afterStart hooks
	for _, fn := range a.opts.afterStart {
		if err := fn(appCtx); err != nil {
			return err
		}
	}

	// Setup signal handler
	c := make(chan os.Signal, 1)
	signal.Notify(c, a.opts.sigs...)

	// Wait for signal or error from any server
	eg.Go(func() error {
		select {
		case <-ctx.Done():
			return nil
		case <-c:
			return a.Stop()
		}
	})

	// Wait for all goroutines to complete
	if err := eg.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		return err
	}

	// Execute afterStop hooks
	var err error
	for _, fn := range a.opts.afterStop {
		err = fn(appCtx)
	}

	return err
}

// Stop gracefully stops the application.
func (a *App) Stop() error {
	// Create application context
	appCtx := NewContext(a.ctx, a)

	var err error
	// Execute beforeStop hooks
	for _, fn := range a.opts.beforeStop {
		err = fn(appCtx)
	}

	// Cancel the context to signal all servers to stop
	if a.cancel != nil {
		a.cancel()
	}

	return err
}

// Context returns the app context.
func (a *App) Context() context.Context {
	return a.ctx
}

// appInfoKey is the key type for app info in context.
type appInfoKey struct{}

// NewContext returns a new Context that carries value.
func NewContext(ctx context.Context, s AppInfo) context.Context {
	return context.WithValue(ctx, appInfoKey{}, s)
}

// FromContext returns the AppInfo from context.
func FromContext(ctx context.Context) (AppInfo, bool) {
	app, ok := ctx.Value(appInfoKey{}).(AppInfo)
	return app, ok
}

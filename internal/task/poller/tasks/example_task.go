// Package tasks provides poller task implementations.
package tasks

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/xelarion/go-layout/internal/task"
	"github.com/xelarion/go-layout/internal/task/poller"
)

// ExampleHandler is an example poller task handler.
type ExampleHandler struct {
	deps   *task.Dependencies
	logger *zap.Logger
}

// NewExampleHandler creates a new example handler.
func NewExampleHandler(deps *task.Dependencies, logger *zap.Logger) *ExampleHandler {
	return &ExampleHandler{
		deps:   deps,
		logger: logger.Named("example-handler"),
	}
}

// Register registers this handler with the poller.
func (t *ExampleHandler) Register(p *poller.Poller) error {
	// Runs every 1 hour
	if err := p.Register("example-task", time.Hour, t.Execute); err != nil {
		return fmt.Errorf("failed to register example task handler: %w", err)
	}
	return nil
}

// Execute runs the example task handler.
func (t *ExampleHandler) Execute(ctx context.Context) error {
	// Simple example of using dependencies
	_, count, err := t.deps.UserRepo.List(ctx, map[string]any{"enabled": true}, 10, 0, "")
	if err != nil {
		return err
	}

	t.logger.Info("Retrieved enabled users", zap.Int("count", count))
	return nil
}

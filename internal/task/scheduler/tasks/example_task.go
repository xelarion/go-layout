// Package tasks provides scheduler task implementations.
package tasks

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/xelarion/go-layout/internal/task"
	"github.com/xelarion/go-layout/internal/task/scheduler"
)

// ExampleHandler is an example scheduler task handler.
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

// Register registers this handler with the scheduler.
func (t *ExampleHandler) Register(s *scheduler.Scheduler) error {
	// Runs every 5 minutes
	if err := s.Register("example-task", "0 */5 * * * *", t.Execute); err != nil {
		return fmt.Errorf("failed to register example task handler: %w", err)
	}
	return nil
}

// Execute runs the example task handler.
func (t *ExampleHandler) Execute(ctx context.Context) error {
	// Simple example of using dependencies
	user, err := t.deps.UserRepo.FindByID(ctx, 1)
	if err != nil {
		t.logger.Error("Failed to find user", zap.Error(err))
		return err
	}

	t.logger.Info("Found user", zap.String("username", user.Username))
	return nil
}

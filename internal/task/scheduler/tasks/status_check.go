// Package tasks provides scheduled task implementations.
package tasks

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/xelarion/go-layout/internal/task/scheduler"
)

// StatusCheckHandler simulates checking system status.
type StatusCheckHandler struct {
	logger *zap.Logger
}

// NewStatusCheckHandler creates a new status check handler.
func NewStatusCheckHandler(logger *zap.Logger) *StatusCheckHandler {
	return &StatusCheckHandler{
		logger: logger.Named("status-check-handler"),
	}
}

// Register registers this handler with the scheduler.
func (t *StatusCheckHandler) Register(s *scheduler.Scheduler) error {
	// Runs every 5 minutes
	if err := s.Register("status-check", "0 */5 * * * *", t.Execute); err != nil {
		return fmt.Errorf("failed to register status-check handler: %w", err)
	}
	return nil
}

// Execute runs the status check handler.
func (t *StatusCheckHandler) Execute(ctx context.Context) error {
	t.logger.Info("System status check", zap.String("status", "healthy"))

	// Simulate some work
	select {
	case <-time.After(2 * time.Second):
		t.logger.Info("Status check completed")
	case <-ctx.Done():
		return fmt.Errorf("status check interrupted: %w", ctx.Err())
	}

	return nil
}

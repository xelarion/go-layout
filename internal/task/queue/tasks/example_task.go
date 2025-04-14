// Package tasks provides queue task implementations.
package tasks

import (
	"context"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"

	"github.com/xelarion/go-layout/internal/task"
	"github.com/xelarion/go-layout/internal/task/queue"
)

// ExamplePayload is the payload for the example task.
type ExamplePayload struct {
	UserID uint `json:"user_id"`
}

// ExampleHandler is an example queue task handler.
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

// Register registers this handler with the queue.
func (t *ExampleHandler) Register(q *queue.Manager) error {
	if err := q.RegisterConsumer("example-handler", "example-queue", "example.task", queue.ConvertHandlerFunc(t.Execute)); err != nil {
		return fmt.Errorf("failed to register example task handler: %w", err)
	}
	return nil
}

// Execute runs the example task handler.
func (t *ExampleHandler) Execute(ctx context.Context, rawPayload []byte) error {
	var payload ExamplePayload
	if err := json.Unmarshal(rawPayload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	t.logger.Info("Example queue task starting", zap.Uint("user_id", payload.UserID))

	// Simple example of using dependencies
	user, err := t.deps.UserRepo.FindByID(ctx, payload.UserID)
	if err != nil {
		t.logger.Error("Failed to find user", zap.Error(err), zap.Uint("user_id", payload.UserID))
		return err
	}

	t.logger.Info("Successfully processed user", zap.String("username", user.Username))
	t.logger.Info("Example queue task completed")
	return nil
}

// Package tasks provides queue task implementations.
package tasks

import (
	"context"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"

	"github.com/xelarion/go-layout/internal/task"
	"github.com/xelarion/go-layout/internal/task/queue"
	"github.com/xelarion/go-layout/pkg/errs"
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
	if err := q.RegisterConsumer(
		"example-handler",
		"example-queue",
		t.Execute,
		queue.WithConsumerOptionsConcurrency(3),
		//queue.WithConsumerOptionsDurable(true),
		//queue.WithConsumerOptionsRoutingKey("example.task"),
	); err != nil {
		return fmt.Errorf("failed to register example task handler: %w", err)
	}
	return nil
}

// Execute runs the example task handler.
func (t *ExampleHandler) Execute(ctx context.Context, msg queue.Message) (queue.Action, error) {
	var payload ExamplePayload
	if err := json.Unmarshal(msg.Body, &payload); err != nil {
		return queue.NackDiscard, errs.WrapInternal(err, "failed to unmarshal payload")
	}

	// Simple example of using dependencies
	user, err := t.deps.UserRepo.FindByID(ctx, payload.UserID)
	if err != nil {
		if errs.IsReason(err, errs.ReasonNotFound) {
			return queue.NackDiscard, err
		}
		return queue.NackRequeue, err
	}

	t.logger.Info("Successfully processed user", zap.String("username", user.Username))
	return queue.Ack, nil
}

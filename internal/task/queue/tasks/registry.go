// Package tasks provides queue task implementations.
package tasks

import (
	"go.uber.org/zap"

	"github.com/xelarion/go-layout/internal/task"
	"github.com/xelarion/go-layout/internal/task/queue"
)

// TaskHandler defines the interface that all queue task handlers must implement
type TaskHandler interface {
	Register(q *queue.Manager) error
}

// TaskHandlerConstructor defines a function type for creating task handlers
type TaskHandlerConstructor func(*task.Dependencies, *zap.Logger) TaskHandler

// taskRegistry holds all registered task handlers
var taskRegistry = map[string]TaskHandlerConstructor{
	"example-task": func(deps *task.Dependencies, logger *zap.Logger) TaskHandler {
		return NewExampleHandler(deps, logger)
	},
	// Add new task constructors here
}

// RegisterAll registers all queue tasks with the provided queue manager
func RegisterAll(qm *queue.Manager, deps *task.Dependencies, logger *zap.Logger) error {
	for taskName, constructorFn := range taskRegistry {
		taskHandler := constructorFn(deps, logger)
		if err := taskHandler.Register(qm); err != nil {
			logger.Error("Failed to register queue handler",
				zap.String("name", taskName),
				zap.Error(err))
			return err
		}
		logger.Info("Registered queue handler", zap.String("name", taskName))
	}
	return nil
}

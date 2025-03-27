// Package tasks provides scheduler task implementations.
package tasks

import (
	"go.uber.org/zap"

	"github.com/xelarion/go-layout/internal/task/scheduler"
)

// TaskHandler defines the interface that all scheduler task handlers must implement
type TaskHandler interface {
	Register(s *scheduler.Scheduler) error
}

// TaskHandlerConstructor defines a function type for creating task handlers
type TaskHandlerConstructor func(*zap.Logger) TaskHandler

// taskRegistry holds all registered task handlers
var taskRegistry = map[string]TaskHandlerConstructor{
	"status-check": func(logger *zap.Logger) TaskHandler {
		return NewStatusCheckHandler(logger)
	},
	// Add new task constructors here
}

// RegisterAll registers all scheduler tasks with the provided scheduler
func RegisterAll(s *scheduler.Scheduler, logger *zap.Logger) error {
	for taskName, constructorFn := range taskRegistry {
		taskHandler := constructorFn(logger)
		if err := taskHandler.Register(s); err != nil {
			logger.Error("Failed to register scheduler handler",
				zap.String("name", taskName),
				zap.Error(err))
			return err
		}
		logger.Info("Registered scheduler handler", zap.String("name", taskName))
	}
	return nil
}

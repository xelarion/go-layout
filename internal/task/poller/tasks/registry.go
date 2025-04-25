// Package tasks provides poller task implementations.
package tasks

import (
	"fmt"

	"go.uber.org/zap"

	"github.com/xelarion/go-layout/internal/task"
	"github.com/xelarion/go-layout/internal/task/poller"
)

// TaskHandler defines the interface that all poller task handlers must implement
type TaskHandler interface {
	Register(p *poller.Poller) error
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

// RegisterAll registers all poller tasks with the provided poller
func RegisterAll(p *poller.Poller, deps *task.Dependencies, logger *zap.Logger) error {
	for taskName, constructorFn := range taskRegistry {
		taskHandler := constructorFn(deps, logger)
		if err := taskHandler.Register(p); err != nil {
			return fmt.Errorf("failed to register poller handler %s: %w", taskName, err)
		}
	}
	return nil
}

//go:build wireinject
// +build wireinject

// Package main contains the wire implementation for dependency injection.
package main

import (
	"github.com/google/wire"
	"go.uber.org/zap"

	"github.com/xelarion/go-layout/internal/infra/config"
	"github.com/xelarion/go-layout/internal/infra/mq"
	"github.com/xelarion/go-layout/internal/infra/storage"
	"github.com/xelarion/go-layout/internal/infra/wsocket"
	"github.com/xelarion/go-layout/internal/repository"
	"github.com/xelarion/go-layout/internal/task"
	"github.com/xelarion/go-layout/internal/task/poller"
	pollerTasks "github.com/xelarion/go-layout/internal/task/poller/tasks"
	"github.com/xelarion/go-layout/internal/task/queue"
	queueTasks "github.com/xelarion/go-layout/internal/task/queue/tasks"
	"github.com/xelarion/go-layout/internal/task/scheduler"
	schedulerTasks "github.com/xelarion/go-layout/internal/task/scheduler/tasks"
	"github.com/xelarion/go-layout/pkg/app"
)

// initApp initializes the Task application with all needed resources.
func initApp(cfgPG *config.PG, cfgRedis *config.Redis, cfgRabbitMQ *config.RabbitMQ, cfgExport *config.Export, cfgStorage *config.Storage, logger *zap.Logger) (*app.App, func(), error) {
	wire.Build(
		task.ProviderSet,
		storage.ProvideStorage,
		mq.NewRabbitMQ,
		provideRabbitMQExchange,
		provideBroadcaster,
		provideTaskServer,
		newApp,
	)
	return nil, nil, nil
}

// provideRabbitMQExchange provides RabbitMQ exchange name from config.
func provideRabbitMQExchange(cfg *config.RabbitMQ) string {
	return cfg.Exchange
}

// provideBroadcaster provides Broadcaster instance (publisher only for task).
func provideBroadcaster(data *repository.Data, logger *zap.Logger) wsocket.Broadcaster {
	return wsocket.PublisherOnlyBroadcasterProvider(data.Redis(), logger)
}

// provideTaskServer provides the task server instance with all task runners.
func provideTaskServer(deps *task.Dependencies, cfgRabbitMQ *config.RabbitMQ, logger *zap.Logger) (*taskServer, error) {
	// Initialize scheduler
	taskScheduler := scheduler.NewScheduler(scheduler.Config{
		Logger: logger,
	})
	if err := schedulerTasks.RegisterAll(taskScheduler, deps, logger); err != nil {
		return nil, err
	}

	// Initialize poller
	taskPoller := poller.NewPoller(logger)
	if err := pollerTasks.RegisterAll(taskPoller, deps, logger); err != nil {
		return nil, err
	}

	// Initialize queue manager using Data from deps
	queueManager := queue.NewManager(deps.Data().RabbitMQ(), cfgRabbitMQ, logger)
	if err := queueTasks.RegisterAll(queueManager, deps, logger); err != nil {
		return nil, err
	}

	return newTaskServer(taskScheduler, taskPoller, logger), nil
}

// Package main contains the entry point for the Task service.
package main

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/xelarion/go-layout/internal/infra/config"
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
// It sets up database connections, message queues, and all task runners.
func initApp(cfgPG *config.PG, cfgRedis *config.Redis, cfgRabbitMQ *config.RabbitMQ, logger *zap.Logger) (*app.App, func(), error) {
	// Initialize data with connections
	data, dataCleanup, err := repository.NewData(cfgPG, cfgRedis, cfgRabbitMQ, logger)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize data: %w", err)
	}

	// Initialize dependencies - this manages all connections and their cleanup
	dependencies, err := task.NewDependencies(data, logger)
	if err != nil {
		dataCleanup()
		return nil, nil, fmt.Errorf("failed to initialize dependencies: %w", err)
	}

	// Initialize task components
	var (
		taskScheduler *scheduler.Scheduler
		taskPoller    *poller.Poller
		queueManager  *queue.Manager
	)

	// Initialize scheduler
	taskScheduler = scheduler.NewScheduler(scheduler.Config{
		Logger: logger,
	})
	// Register all scheduler tasks with dependencies
	if err := schedulerTasks.RegisterAll(taskScheduler, dependencies, logger); err != nil {
		dataCleanup()
		return nil, nil, err
	}

	// Initialize poller
	taskPoller = poller.NewPoller(logger)
	// Register all poller tasks with dependencies
	if err := pollerTasks.RegisterAll(taskPoller, dependencies, logger); err != nil {
		dataCleanup()
		return nil, nil, err
	}

	// Initialize queue manager using RabbitMQ from dependencies
	queueManager = queue.NewManager(data.RabbitMQ(), cfgRabbitMQ, logger)
	// Register all queue tasks with dependencies
	if err := queueTasks.RegisterAll(queueManager, dependencies, logger); err != nil {
		dataCleanup()
		return nil, nil, err
	}

	// Create a custom server that implements app.Server
	ts := newTaskServer(
		taskScheduler,
		taskPoller,
		logger,
	)

	// Create application using the newApp function
	appInstance := newApp(logger, ts)

	// Create a combined cleanup function
	cleanup := func() {
		dataCleanup()
	}

	return appInstance, cleanup, nil
}

// taskServer implements app.Server for Task service
type taskServer struct {
	scheduler *scheduler.Scheduler
	poller    *poller.Poller
	logger    *zap.Logger
	ctx       context.Context
	cancel    context.CancelFunc
}

func newTaskServer(
	scheduler *scheduler.Scheduler,
	poller *poller.Poller,
	logger *zap.Logger,
) *taskServer {
	ctx, cancel := context.WithCancel(context.Background())
	return &taskServer{
		scheduler: scheduler,
		poller:    poller,
		logger:    logger,
		ctx:       ctx,
		cancel:    cancel,
	}
}

// Start implements app.Server interface
func (s *taskServer) Start(ctx context.Context) error {
	// Start scheduler
	if s.scheduler != nil {
		s.scheduler.Start()
	}

	// Start poller
	if s.poller != nil {
		s.poller.Start()
	}

	// Block until context is done
	<-s.ctx.Done()
	return nil
}

// Stop implements app.Server interface
func (s *taskServer) Stop(ctx context.Context) error {
	// Stop poller
	if s.poller != nil {
		s.poller.Stop()
	}

	// Stop scheduler
	if s.scheduler != nil {
		s.scheduler.Stop()
	}

	// Signal the task server to stop
	if s.cancel != nil {
		s.cancel()
	}

	return nil
}

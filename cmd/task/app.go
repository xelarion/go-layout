// Package main contains the entry point for the Task service.
package main

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/xelarion/go-layout/internal/task"
	"github.com/xelarion/go-layout/internal/task/poller"
	pollerTasks "github.com/xelarion/go-layout/internal/task/poller/tasks"
	"github.com/xelarion/go-layout/internal/task/queue"
	queueTasks "github.com/xelarion/go-layout/internal/task/queue/tasks"
	"github.com/xelarion/go-layout/internal/task/scheduler"
	schedulerTasks "github.com/xelarion/go-layout/internal/task/scheduler/tasks"
	"github.com/xelarion/go-layout/pkg/app"
	"github.com/xelarion/go-layout/pkg/cache"
	"github.com/xelarion/go-layout/pkg/config"
	"github.com/xelarion/go-layout/pkg/database"
	"github.com/xelarion/go-layout/pkg/mq"
)

// initApp initializes the Task application with all needed resources.
// It sets up database connections, message queues, and all task runners based on
// the provided flag values.
func initApp(cfg *config.Config, logger *zap.Logger, enableScheduler, enablePoller, enableQueue *bool) (*app.App, error) {
	logger.Info("Initializing Task application")

	// Connect to database
	db, err := database.NewPostgres(&cfg.PG, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	logger.Info("Connected to database successfully")

	// Initialize redis connection
	redis, err := cache.NewRedis(&cfg.Redis, logger)
	if err != nil {
		// Clean up database before returning
		if closeErr := db.Close(); closeErr != nil {
			logger.Error("Failed to close database connection during error handling", zap.Error(closeErr))
		}
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}
	logger.Info("Connected to redis successfully")

	// Create task dependencies
	dependencies := task.NewDependencies(db.DB, redis.Client, logger)

	// Initialize task components
	var (
		taskScheduler *scheduler.Scheduler
		taskPoller    *poller.Poller
		queueManager  *queue.Manager
		rabbitMQ      *mq.RabbitMQ
	)

	// Initialize RabbitMQ if queue is enabled
	if *enableQueue {
		rabbitMQ, err = mq.NewRabbitMQ(&cfg.RabbitMQ, logger)
		if err != nil {
			// Clean up database before returning
			if closeErr := db.Close(); closeErr != nil {
				logger.Error("Failed to close database connection during error handling", zap.Error(closeErr))
			}
			return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
		}
		logger.Info("Connected to RabbitMQ successfully")

		// Initialize queue manager
		queueManager = queue.NewManager(rabbitMQ, &cfg.RabbitMQ, logger)
	}

	// Initialize scheduler if enabled
	if *enableScheduler {
		taskScheduler = scheduler.NewScheduler(scheduler.Config{
			Logger: logger,
		})

		// Register all scheduler tasks with dependencies
		if err := schedulerTasks.RegisterAll(taskScheduler, dependencies, logger); err != nil {
			logger.Error("Failed to register scheduler tasks", zap.Error(err))
		}
	}

	// Initialize poller if enabled
	if *enablePoller {
		taskPoller = poller.NewPoller(logger)

		// Register all poller tasks with dependencies
		if err := pollerTasks.RegisterAll(taskPoller, dependencies, logger); err != nil {
			logger.Error("Failed to register poller tasks", zap.Error(err))
		}
	}

	// Initialize queue processors if enabled
	if *enableQueue && queueManager != nil {
		// Register all queue tasks with dependencies
		if err := queueTasks.RegisterAll(queueManager, dependencies, logger); err != nil {
			logger.Error("Failed to register queue tasks", zap.Error(err))
		}
	}

	// Create the application with start and stop functions
	logger.Info("Creating Task application")
	taskApp := app.NewApp(
		"task",
		"Task Service",
		"0.1.0",
		logger,
		app.WithStartFunc(func(ctx context.Context) error {
			logger.Info("Starting Task service")

			// Start scheduler if enabled
			if *enableScheduler && taskScheduler != nil {
				logger.Info("Starting scheduler")
				taskScheduler.Start()
			}

			// Start poller if enabled
			if *enablePoller && taskPoller != nil {
				logger.Info("Starting poller")
				taskPoller.Start()
			}

			// Block until context is done
			<-ctx.Done()

			return nil
		}),
		app.WithStopFunc(func(ctx context.Context) error {
			logger.Info("Stopping Task service")

			// Stop poller if enabled
			if *enablePoller && taskPoller != nil {
				logger.Info("Stopping poller")
				taskPoller.Stop()
			}

			// Stop scheduler if enabled
			if *enableScheduler && taskScheduler != nil {
				logger.Info("Stopping scheduler")
				taskScheduler.Stop()
			}

			// Close RabbitMQ connection if initialized
			if rabbitMQ != nil {
				logger.Info("Closing RabbitMQ connection")
				rabbitMQ.Close()
			}

			// Close database connection
			if err := db.Close(); err != nil {
				logger.Error("Error closing database connection", zap.Error(err))
			} else {
				logger.Info("Database connection closed successfully")
			}

			// Close redis connection
			if err := redis.Close(); err != nil {
				logger.Error("Error closing redis connection", zap.Error(err))
			} else {
				logger.Info("Redis connection closed successfully")
			}

			logger.Info("Task service stopped successfully")
			return nil
		}),
	)

	logger.Info("Task application initialized successfully")
	return taskApp, nil
}

// Package scheduler provides scheduling capabilities for running tasks at fixed times, dates, or intervals.
package scheduler

import (
	"context"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"

	"github.com/xelarion/go-layout/pkg/errs"
)

// Task represents a function that can be scheduled.
type Task func(ctx context.Context) error

// Scheduler manages scheduled tasks.
type Scheduler struct {
	cron   *cron.Cron
	logger *zap.Logger
	tasks  map[string]cron.EntryID
	mu     sync.RWMutex
}

// Config holds the configuration for the scheduler.
type Config struct {
	Location *time.Location
	Logger   *zap.Logger
}

// NewScheduler creates a new task scheduler.
func NewScheduler(cfg Config) *Scheduler {
	if cfg.Location == nil {
		cfg.Location = time.UTC
	}

	if cfg.Logger == nil {
		cfg.Logger = zap.NewNop()
	}

	cronOptions := []cron.Option{
		cron.WithLocation(cfg.Location),
		cron.WithSeconds(), // Enable seconds level precision
	}

	return &Scheduler{
		cron:   cron.New(cronOptions...),
		logger: cfg.Logger.Named("scheduler"),
		tasks:  make(map[string]cron.EntryID),
	}
}

// Register adds a task to the scheduler with the given cron expression.
// Returns an error if the task name already exists or if the cron expression is invalid.
func (s *Scheduler) Register(name, cronExpr string, task Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tasks[name]; exists {
		return ErrTaskAlreadyExists
	}

	id, err := s.cron.AddFunc(cronExpr, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		s.logger.Info("Starting scheduled task", zap.String("task", name))
		start := time.Now()

		if err := task(ctx); err != nil {
			s.logger.WithOptions(
				zap.WithCaller(false),
				zap.AddStacktrace(zap.FatalLevel),
			).Error("Scheduled task failed",
				zap.String("task", name),
				zap.Error(err),
				zap.String("stack_trace", errs.GetStack(err)),
				zap.Duration("duration", time.Since(start)))
			return
		}

		s.logger.Info("Scheduled task completed",
			zap.String("task", name),
			zap.Duration("duration", time.Since(start)))
	})

	if err != nil {
		return err
	}

	s.tasks[name] = id
	return nil
}

// Unregister removes a task from the scheduler.
func (s *Scheduler) Unregister(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if id, exists := s.tasks[name]; exists {
		s.cron.Remove(id)
		delete(s.tasks, name)
		s.logger.Info("Task unregistered", zap.String("task", name))
	}
}

// Start starts the scheduler.
func (s *Scheduler) Start() {
	s.cron.Start()
	s.logger.Info("Scheduler started")
}

// Stop stops the scheduler.
func (s *Scheduler) Stop() {
	ctx := s.cron.Stop()
	<-ctx.Done()
	s.logger.Info("Scheduler stopped")
}

// ListTasks returns a list of all registered tasks.
func (s *Scheduler) ListTasks() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var tasks []string
	for task := range s.tasks {
		tasks = append(tasks, task)
	}
	return tasks
}

// GetEntries returns all the cron entries.
func (s *Scheduler) GetEntries() []cron.Entry {
	return s.cron.Entries()
}

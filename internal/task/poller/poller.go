// Package poller provides functionality for running tasks at fixed intervals.
package poller

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/xelarion/go-layout/pkg/errs"
)

// Task represents a polling task function.
type Task func(ctx context.Context) error

// Poller manages polling tasks that run at fixed intervals.
type Poller struct {
	tasks   map[string]*pollingTask
	mu      sync.RWMutex
	logger  *zap.Logger
	wg      sync.WaitGroup
	ctx     context.Context
	cancel  context.CancelFunc
	started bool
}

// pollingTask represents a task that runs on a polling schedule.
type pollingTask struct {
	name     string
	interval time.Duration
	task     Task
	ctx      context.Context
	cancel   context.CancelFunc
	logger   *zap.Logger
	done     chan struct{}
}

// NewPoller creates a new Poller instance.
func NewPoller(logger *zap.Logger) *Poller {
	ctx, cancel := context.WithCancel(context.Background())

	if logger == nil {
		logger = zap.NewNop()
	}

	return &Poller{
		tasks:   make(map[string]*pollingTask),
		logger:  logger.Named("poller"),
		ctx:     ctx,
		cancel:  cancel,
		started: false,
	}
}

// Register adds a new polling task with the specified interval.
// Returns an error if a task with the same name already exists.
func (p *Poller) Register(name string, interval time.Duration, task Task) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, exists := p.tasks[name]; exists {
		return fmt.Errorf("task '%s' already exists", name)
	}

	taskCtx, taskCancel := context.WithCancel(p.ctx)
	pt := &pollingTask{
		name:     name,
		interval: interval,
		task:     task,
		ctx:      taskCtx,
		cancel:   taskCancel,
		logger:   p.logger.With(zap.String("task", name)),
		done:     make(chan struct{}),
	}

	p.tasks[name] = pt

	// If the poller is already started, start this task immediately
	if p.started {
		p.wg.Add(1)
		go p.runTask(pt)
	}

	return nil
}

// Unregister removes a task from the poller.
func (p *Poller) Unregister(name string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if task, exists := p.tasks[name]; exists {
		task.cancel() // Cancel the task's context
		if p.started {
			<-task.done // Wait for the task to complete
		}
		delete(p.tasks, name)
	}
}

// Start starts all registered tasks.
func (p *Poller) Start() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.started {
		return
	}

	for _, task := range p.tasks {
		p.wg.Add(1)
		go p.runTask(task)
	}

	p.started = true
}

// runTask runs a polling task at the specified interval.
func (p *Poller) runTask(task *pollingTask) {
	defer func() {
		p.wg.Done()
		close(task.done) // Signal that the task has stopped
	}()

	ticker := time.NewTicker(task.interval)
	defer ticker.Stop()

	// Execute task immediately on start
	p.executeTask(task)

	for {
		select {
		case <-ticker.C:
			p.executeTask(task)
		case <-task.ctx.Done():
			return
		}
	}
}

// executeTask runs a single execution of the task with proper logging.
func (p *Poller) executeTask(task *pollingTask) {
	// Use a timeout context for each execution, but make it less than the interval to avoid overlaps
	timeout := task.interval * 80 / 100 // 80% of the interval as timeout
	execCtx, cancel := context.WithTimeout(task.ctx, timeout)
	defer cancel()

	start := time.Now()

	if err := task.task(execCtx); err != nil {
		task.logger.WithOptions(
			zap.WithCaller(false),
			zap.AddStacktrace(zap.FatalLevel),
		).Error("Polling task failed"+errs.GetStack(err),
			zap.String("name", task.name),
			zap.Error(err),
			zap.Duration("duration", time.Since(start)))
		return
	}

	task.logger.Info("Polling task completed",
		zap.Duration("duration", time.Since(start)))
}

// Stop stops all polling tasks and waits for them to complete.
func (p *Poller) Stop() {
	p.mu.Lock()
	if !p.started {
		p.mu.Unlock()
		return
	}
	p.started = false
	p.mu.Unlock()

	p.cancel()  // Cancel the parent context
	p.wg.Wait() // Wait for all tasks to complete
}

// ListTasks returns a list of all registered task names.
func (p *Poller) ListTasks() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	tasks := make([]string, 0, len(p.tasks))
	for name := range p.tasks {
		tasks = append(tasks, name)
	}
	return tasks
}

// GetTaskInterval returns the interval of a specific task.
func (p *Poller) GetTaskInterval(name string) (time.Duration, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if task, exists := p.tasks[name]; exists {
		return task.interval, true
	}
	return 0, false
}

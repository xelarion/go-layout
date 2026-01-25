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
		cfg.Location = time.Local
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
// Returns an error if the task name already exists or if the cron expression is invalid./* cron AddFunc spec
/* cronExpr
	c.AddFunc("30 * * * *", func() { fmt.Println("Every hour on the half hour") })
	c.AddFunc("30 3-6,20-23 * * *", func() { fmt.Println(".. in the range 3-6am, 8-11pm") })
	c.AddFunc("CRON_TZ=Asia/Tokyo 30 04 * * *", func() { fmt.Println("Runs at 04:30 Tokyo time every day") })
	c.AddFunc("@hourly", func() { fmt.Println("Every hour, starting an hour from now") })
	c.AddFunc("@daily", func() { fmt.Println("Every day") })
	c.AddFunc("@every 1h30m", func() { fmt.Println("Every hour thirty, starting an hour thirty from now") })
	c.AddFunc("@every 10s", func() { fmt.Println("Every 10 seconds, starting 10 seconds from now") })

┌───────────── minute (0-59)
│ ┌───────────── hour (0-23)
│ │ ┌───────────── day of the month (1-31)
│ │ │ ┌───────────── month (1-12 or JAN-DEC)
│ │ │ │ ┌───────────── day of the week (0-6 or SUN-SAT)
│ │ │ │ │
│ │ │ │ │
* * * * *

You may use one of several pre-defined schedules in place of a cron expression.

	Entry                  | Description                                | Equivalent To
	-----                  | -----------                                | -------------
	@yearly (or @annually) | Run once a year, midnight, Jan. 1st        | 0 0 1 1 *
	@monthly               | Run once a month, midnight, first of month | 0 0 1 * *
	@weekly                | Run once a week, midnight between Sat/Sun  | 0 0 * * 0
	@daily (or @midnight)  | Run once a day, midnight                   | 0 0 * * *
	@hourly                | Run once an hour, beginning of hour        | 0 * * * *

You may also schedule a job to execute at fixed intervals, starting at the time it's added
or cron is run. This is supported by formatting the cron spec like this:

    @every <duration>

where "duration" is a string accepted by time.ParseDuration
(http://golang.org/pkg/time/#ParseDuration).

*/
func (s *Scheduler) Register(name, cronExpr string, task Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tasks[name]; exists {
		return ErrTaskAlreadyExists
	}

	id, err := s.cron.AddFunc(cronExpr, func() {
		ctx := context.Background()

		if err := task(ctx); err != nil {
			s.logger.WithOptions(
				zap.WithCaller(false),
				zap.AddStacktrace(zap.FatalLevel),
			).Error("Scheduled task failed"+errs.GetStack(err),
				zap.String("name", name),
				zap.Error(err))
		}
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
	}
}

// Start starts the scheduler.
func (s *Scheduler) Start() {
	s.cron.Start()
}

// Stop stops the scheduler.
func (s *Scheduler) Stop() {
	ctx := s.cron.Stop()
	<-ctx.Done()
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

// Package main contains the entry point for the Task service.
package main

import (
	"context"

	"go.uber.org/zap"

	"github.com/xelarion/go-layout/internal/task/poller"
	"github.com/xelarion/go-layout/internal/task/scheduler"
)

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

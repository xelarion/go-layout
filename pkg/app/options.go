package app

import (
	"context"
	"os"
	"time"
)

// Option configures App options.
type Option func(*options)

// options holds App configuration options
type options struct {
	id       string
	name     string
	version  string
	metadata map[string]string

	ctx  context.Context
	sigs []os.Signal

	stopTimeout time.Duration
	servers     []Server

	beforeStart []func(context.Context) error
	afterStart  []func(context.Context) error
	beforeStop  []func(context.Context) error
	afterStop   []func(context.Context) error
}

// WithID sets the app instance ID.
func WithID(id string) Option {
	return func(o *options) {
		o.id = id
	}
}

// WithName sets the app name.
func WithName(name string) Option {
	return func(o *options) {
		o.name = name
	}
}

// WithVersion sets the app version.
func WithVersion(version string) Option {
	return func(o *options) {
		o.version = version
	}
}

// WithMetadata sets the app metadata.
func WithMetadata(md map[string]string) Option {
	return func(o *options) {
		o.metadata = md
	}
}

// WithContext sets the app context.
func WithContext(ctx context.Context) Option {
	return func(o *options) {
		o.ctx = ctx
	}
}

// WithSignals sets the signals for graceful shutdown.
func WithSignals(sigs ...os.Signal) Option {
	return func(o *options) {
		o.sigs = sigs
	}
}

// WithStopTimeout sets the timeout for graceful shutdown.
func WithStopTimeout(timeout time.Duration) Option {
	return func(o *options) {
		o.stopTimeout = timeout
	}
}

// WithServers appends a server to the app.
func WithServers(srv ...Server) Option {
	return func(o *options) {
		o.servers = srv
	}
}

// WithBeforeStart appends a hook function to be run before app starts.
func WithBeforeStart(fn func(context.Context) error) Option {
	return func(o *options) {
		o.beforeStart = append(o.beforeStart, fn)
	}
}

// WithAfterStart appends a hook function to be run after app starts.
func WithAfterStart(fn func(context.Context) error) Option {
	return func(o *options) {
		o.afterStart = append(o.afterStart, fn)
	}
}

// WithBeforeStop appends a hook function to be run before app stops.
func WithBeforeStop(fn func(context.Context) error) Option {
	return func(o *options) {
		o.beforeStop = append(o.beforeStop, fn)
	}
}

// WithAfterStop appends a hook function to be run after app stops.
func WithAfterStop(fn func(context.Context) error) Option {
	return func(o *options) {
		o.afterStop = append(o.afterStop, fn)
	}
}

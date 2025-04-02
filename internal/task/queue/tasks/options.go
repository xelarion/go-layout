// Package tasks provides queue task implementations.
package tasks

import (
	"github.com/xelarion/go-layout/internal/task/queue"
)

// WithHighPriority sets the task priority to high (9).
func WithHighPriority() func(*queue.PublishOptions) {
	return func(opts *queue.PublishOptions) {
		opts.Priority = 9
	}
}

// WithMediumPriority sets the task priority to medium (5).
func WithMediumPriority() func(*queue.PublishOptions) {
	return func(opts *queue.PublishOptions) {
		opts.Priority = 5
	}
}

// WithLowPriority sets the task priority to low (1).
func WithLowPriority() func(*queue.PublishOptions) {
	return func(opts *queue.PublishOptions) {
		opts.Priority = 1
	}
}

// WithExpirationMinutes sets the expiration time in minutes.
func WithExpirationMinutes(minutes int) func(*queue.PublishOptions) {
	return func(opts *queue.PublishOptions) {
		// Convert minutes to milliseconds
		milliseconds := minutes * 60 * 1000
		opts.Expiration = queue.FormatExpiration(milliseconds)
	}
}

// WithExpirationHours sets the expiration time in hours.
func WithExpirationHours(hours int) func(*queue.PublishOptions) {
	return WithExpirationMinutes(hours * 60)
}

// WithPersistentDelivery ensures the task is persisted to disk.
func WithPersistentDelivery() func(*queue.PublishOptions) {
	return func(opts *queue.PublishOptions) {
		opts.DeliveryMode = 2 // Persistent
	}
}

// WithTransientDelivery marks the task as transient (non-persistent).
func WithTransientDelivery() func(*queue.PublishOptions) {
	return func(opts *queue.PublishOptions) {
		opts.DeliveryMode = 1 // Transient
	}
}

// WithTasksExchange uses the default tasks exchange for this message.
func WithTasksExchange(manager *queue.Manager) func(*queue.PublishOptions) {
	return func(opts *queue.PublishOptions) {
		opts.Exchange = manager.GetTasksExchange()
	}
}

// WithCorrelationID sets the correlation ID for the message.
func WithCorrelationID(correlationID string) func(*queue.PublishOptions) {
	return func(opts *queue.PublishOptions) {
		opts.CorrelationID = correlationID
	}
}

// WithReplyTo sets the reply-to address for request-reply pattern.
func WithReplyTo(replyTo string) func(*queue.PublishOptions) {
	return func(opts *queue.PublishOptions) {
		opts.ReplyTo = replyTo
	}
}

// WithMessageID sets a custom message ID.
func WithMessageID(messageID string) func(*queue.PublishOptions) {
	return func(opts *queue.PublishOptions) {
		opts.MessageID = messageID
	}
}

// WithCustomHeader adds a custom header to the message.
func WithCustomHeader(key string, value any) func(*queue.PublishOptions) {
	return func(opts *queue.PublishOptions) {
		if opts.Headers == nil {
			opts.Headers = make(map[string]any)
		}
		opts.Headers[key] = value
	}
}

// Package queue provides functionality for handling asynchronous queue-based tasks using RabbitMQ.
package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/wagslane/go-rabbitmq"
	"go.uber.org/zap"

	"github.com/xelarion/go-layout/internal/config"
	"github.com/xelarion/go-layout/internal/mq"
	"github.com/xelarion/go-layout/pkg/errs"
)

// Action represents the action to take after processing a message.
type Action int

const (
	// Ack acknowledges the message, removing it from the queue.
	Ack Action = iota
	// NackDiscard negatively acknowledges the message without requeuing.
	NackDiscard
	// NackRequeue negatively acknowledges the message and requeues it.
	NackRequeue
	// Manual leaves acknowledgement to the user.
	Manual
)

// Message represents a message received from the queue.
type Message struct {
	// Body contains the message payload.
	Body []byte
	// DeliveryTag is the delivery tag of the message.
	DeliveryTag uint64
	// MessageID is the message ID if set by the publisher.
	MessageID string
	// CorrelationID is the correlation ID if set by the publisher.
	CorrelationID string
	// ReplyTo is the reply-to address if set by the publisher.
	ReplyTo string
	// Timestamp is the timestamp of the message.
	Timestamp int64
}

// Handler defines a function that processes a queue message.
type Handler func(ctx context.Context, msg Message) (Action, error)

// Manager manages the registration and execution of queue-based tasks.
type Manager struct {
	rmq       *mq.RabbitMQ
	logger    *zap.Logger
	consumers map[string]string
	config    *config.RabbitMQ
	exchange  string
}

// NewManager creates a new queue task manager.
func NewManager(rmq *mq.RabbitMQ, config *config.RabbitMQ, logger *zap.Logger) *Manager {
	if logger == nil {
		logger = zap.NewNop()
	}

	if config == nil {
		// Default exchange names if no config provided
		return &Manager{
			rmq:       rmq,
			logger:    logger.Named("queue-manager"),
			consumers: make(map[string]string),
			exchange:  "go-layout",
		}
	}

	return &Manager{
		rmq:       rmq,
		logger:    logger.Named("queue-manager"),
		consumers: make(map[string]string),
		config:    config,
		exchange:  config.Exchange,
	}
}

// ConsumerOptions represents simplified configuration options for message consumers.
type ConsumerOptions struct {
	// Exchange is the name of the exchange to bind the queue to
	Exchange string
	// RoutingKey is the routing key to bind the queue with
	RoutingKey string
	// QueueName is the name of the queue to consume from
	QueueName string
	// Durable determines if the queue survives broker restarts
	Durable bool
	// AutoDelete determines if the queue is deleted when no consumers
	AutoDelete bool
	// Concurrency is the number of goroutines processing messages
	Concurrency int
	// QOSPrefetch is the number of messages to prefetch
	QOSPrefetch int
	// RequeueDelay is the delay duration before retrying a failed message, defaults to 1 second
	// Set to 0 to disable delay
	RequeueDelay time.Duration
}

// DefaultConsumerOptions returns the default options for a consumer.
func DefaultConsumerOptions(queueName string) *ConsumerOptions {
	return &ConsumerOptions{
		QueueName:    queueName,
		RoutingKey:   queueName, // Default routing key is the same as queue name
		Durable:      true,
		AutoDelete:   false,
		Concurrency:  1,
		QOSPrefetch:  10,
		RequeueDelay: time.Second, // Default 1-second delay
	}
}

// WithConsumerOptionsExchange sets the exchange for the consumer.
func WithConsumerOptionsExchange(exchange string) func(*ConsumerOptions) {
	return func(opts *ConsumerOptions) {
		opts.Exchange = exchange
	}
}

// WithConsumerOptionsRoutingKey sets the routing key for the consumer.
func WithConsumerOptionsRoutingKey(routingKey string) func(*ConsumerOptions) {
	return func(opts *ConsumerOptions) {
		opts.RoutingKey = routingKey
	}
}

// WithConsumerOptionsDurable sets whether the queue is durable.
func WithConsumerOptionsDurable(durable bool) func(*ConsumerOptions) {
	return func(opts *ConsumerOptions) {
		opts.Durable = durable
	}
}

// WithConsumerOptionsAutoDelete sets whether the queue is auto-deleted.
func WithConsumerOptionsAutoDelete(autoDelete bool) func(*ConsumerOptions) {
	return func(opts *ConsumerOptions) {
		opts.AutoDelete = autoDelete
	}
}

// WithConsumerOptionsConcurrency sets the concurrency level for the consumer.
func WithConsumerOptionsConcurrency(concurrency int) func(*ConsumerOptions) {
	return func(opts *ConsumerOptions) {
		opts.Concurrency = concurrency
	}
}

// WithConsumerOptionsQOSPrefetch sets the prefetch count for the consumer.
func WithConsumerOptionsQOSPrefetch(prefetch int) func(*ConsumerOptions) {
	return func(opts *ConsumerOptions) {
		opts.QOSPrefetch = prefetch
	}
}

// WithConsumerOptionsRequeueDelay sets the requeue delay duration.
func WithConsumerOptionsRequeueDelay(delay time.Duration) func(*ConsumerOptions) {
	return func(opts *ConsumerOptions) {
		opts.RequeueDelay = delay
	}
}

// RegisterConsumer registers a new queue consumer with the specified configuration.
func (qm *Manager) RegisterConsumer(
	name string,
	queueName string,
	handler Handler,
	options ...func(*ConsumerOptions),
) error {
	// Validate input
	if name == "" {
		return fmt.Errorf("consumer name cannot be empty")
	}
	if queueName == "" {
		return fmt.Errorf("queue name cannot be empty")
	}
	if handler == nil {
		return fmt.Errorf("handler cannot be nil")
	}

	// Set up options
	opts := DefaultConsumerOptions(queueName)

	// Create logger with consumer context
	consumerLogger := qm.logger.WithOptions(
		zap.WithCaller(false),
		zap.AddStacktrace(zap.FatalLevel),
	).With(
		zap.String("consumer", name),
		zap.String("queue", queueName),
		zap.String("routingKey", opts.RoutingKey),
	)

	// Apply custom options
	for _, option := range options {
		option(opts)
	}

	// Use default exchange if not specified
	if opts.Exchange == "" {
		opts.Exchange = qm.exchange
	}

	// Convert our ConsumerOptions to rabbitmq.ConsumerOptions functions
	rabbitOpts := []func(*rabbitmq.ConsumerOptions){
		// Basic configuration
		rabbitmq.WithConsumerOptionsConcurrency(opts.Concurrency),
		rabbitmq.WithConsumerOptionsQOSPrefetch(opts.QOSPrefetch),
		// Use zap logger
		func(options *rabbitmq.ConsumerOptions) {
			options.Logger = qm.logger.Sugar()
		},
		// Configure exchange
		func(options *rabbitmq.ConsumerOptions) {
			options.ExchangeOptions = append(options.ExchangeOptions, rabbitmq.ExchangeOptions{
				Name:       opts.Exchange,
				Kind:       "direct",
				Durable:    true,
				AutoDelete: false,
				Declare:    true,
				Bindings: []rabbitmq.Binding{
					{
						RoutingKey: opts.RoutingKey,
						BindingOptions: rabbitmq.BindingOptions{
							Declare: true,
						},
					},
				},
			})
		},
		// Configure queue
		func(options *rabbitmq.ConsumerOptions) {
			options.QueueOptions.Name = queueName
			options.QueueOptions.Durable = opts.Durable
			options.QueueOptions.AutoDelete = opts.AutoDelete
			options.QueueOptions.Declare = true
		},
	}

	// Wrap our handler
	consumerHandler := func(d rabbitmq.Delivery) rabbitmq.Action {
		ctx := context.Background()

		// Create message from delivery
		msg := Message{
			Body:          d.Body,
			DeliveryTag:   d.DeliveryTag,
			MessageID:     d.MessageId,
			CorrelationID: d.CorrelationId,
			ReplyTo:       d.ReplyTo,
			Timestamp:     d.Timestamp.UnixNano(),
		}

		// Call handler
		action, err := handler(ctx, msg)

		// Log error if any
		if err != nil {
			consumerLogger.Error("Failed to process message",
				zap.String("error", err.Error()),
				zap.String("stack_trace", errs.GetStack(err)),
			)
		}

		// Determine rabbitmq action from our action
		switch action {
		case Ack:
			return rabbitmq.Ack
		case NackRequeue:
			// Apply delay to all messages that need to be requeued (if delay is configured)
			if opts.RequeueDelay > 0 {
				// Use sleep to delay processing
				time.Sleep(opts.RequeueDelay)
			}
			return rabbitmq.NackRequeue
		case NackDiscard:
			return rabbitmq.NackDiscard
		case Manual:
			return rabbitmq.Manual
		default:
			consumerLogger.Warn("Unknown action, will ack",
				zap.Int("action", int(action)))
			return rabbitmq.Ack
		}
	}

	// Start the consumer
	_, err := qm.rmq.StartConsumer(
		queueName,
		consumerHandler,
		rabbitOpts...,
	)
	if err != nil {
		return fmt.Errorf("failed to start consumer: %w", err)
	}

	qm.consumers[name] = queueName
	return nil
}

// PublishOptions represents simplified configuration options for publishing messages.
type PublishOptions struct {
	// Exchange defines the exchange to publish to
	Exchange string
	// ContentType specifies the MIME content type (default: application/json)
	ContentType string
	// DeliveryMode: Transient (0 or 1) or Persistent (2)
	DeliveryMode uint8
	// Priority from 0 to 9
	Priority uint8
	// Expiration time in ms that a message will expire from a queue
	Expiration string
	// CorrelationID for message correlation
	CorrelationID string
	// ReplyTo address for request-reply pattern
	ReplyTo string
	// MessageID for message identification
	MessageID string
	// Headers for application-specific metadata
	Headers map[string]any
}

// DefaultPublishOptions returns the default options for publishing messages.
func DefaultPublishOptions() *PublishOptions {
	return &PublishOptions{
		Exchange:      "",
		ContentType:   "application/json",
		DeliveryMode:  2, // Persistent
		Priority:      0,
		Expiration:    "",
		CorrelationID: "",
		ReplyTo:       "",
		MessageID:     "",
		Headers:       make(map[string]any),
	}
}

// WithExchange sets the exchange name for publishing.
func WithExchange(exchange string) func(*PublishOptions) {
	return func(opts *PublishOptions) {
		opts.Exchange = exchange
	}
}

// WithContentType sets the content type for publishing.
func WithContentType(contentType string) func(*PublishOptions) {
	return func(opts *PublishOptions) {
		opts.ContentType = contentType
	}
}

// WithPersistentDelivery sets the delivery mode to persistent.
func WithPersistentDelivery() func(*PublishOptions) {
	return func(opts *PublishOptions) {
		opts.DeliveryMode = 2 // Persistent
	}
}

// WithTransientDelivery sets the delivery mode to transient.
func WithTransientDelivery() func(*PublishOptions) {
	return func(opts *PublishOptions) {
		opts.DeliveryMode = 1 // Transient
	}
}

// WithPriority sets the priority for the message.
// Priority should be between 0 and 9.
func WithPriority(priority uint8) func(*PublishOptions) {
	return func(opts *PublishOptions) {
		if priority > 9 {
			priority = 9
		}
		opts.Priority = priority
	}
}

// WithExpiration sets the expiration time of the message.
func WithExpiration(expiration string) func(*PublishOptions) {
	return func(opts *PublishOptions) {
		opts.Expiration = expiration
	}
}

// WithExpirationMilliseconds sets the expiration time in milliseconds.
func WithExpirationMilliseconds(milliseconds int) func(*PublishOptions) {
	return func(opts *PublishOptions) {
		opts.Expiration = FormatExpiration(milliseconds)
	}
}

// WithCorrelationID sets the correlation ID of the message.
func WithCorrelationID(correlationID string) func(*PublishOptions) {
	return func(opts *PublishOptions) {
		opts.CorrelationID = correlationID
	}
}

// WithReplyTo sets the reply-to address of the message.
func WithReplyTo(replyTo string) func(*PublishOptions) {
	return func(opts *PublishOptions) {
		opts.ReplyTo = replyTo
	}
}

// WithMessageID sets the message ID.
func WithMessageID(messageID string) func(*PublishOptions) {
	return func(opts *PublishOptions) {
		opts.MessageID = messageID
	}
}

// WithHeaders sets additional headers for the message.
func WithHeaders(headers map[string]any) func(*PublishOptions) {
	return func(opts *PublishOptions) {
		opts.Headers = headers
	}
}

// FormatExpiration formats milliseconds to the expiration string format expected by RabbitMQ.
func FormatExpiration(milliseconds int) string {
	return strconv.Itoa(milliseconds)
}

// PublishTask publishes a task message to the specified routing key.
// The exchange will be determined by the following precedence:
// 1. Options.Exchange if set in options
// 2. Manager's default exchange from configuration
func (qm *Manager) PublishTask(
	ctx context.Context,
	routingKey string,
	payload any,
	options ...func(*PublishOptions),
) error {
	// Convert payload to JSON
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Apply options to our PublishOptions
	opts := DefaultPublishOptions()
	for _, option := range options {
		option(opts)
	}

	// Use default tasks exchange if not specified in options
	if opts.Exchange == "" {
		opts.Exchange = qm.exchange
	}

	publishLogger := qm.logger.With(
		zap.String("exchange", opts.Exchange),
		zap.String("routing_key", routingKey),
	)

	// Convert our PublishOptions to rabbitmq.PublishOptions
	rabbitOpts := []func(*rabbitmq.PublishOptions){
		rabbitmq.WithPublishOptionsExchange(opts.Exchange),
		rabbitmq.WithPublishOptionsContentType(opts.ContentType),
	}

	// Add additional options if specified
	if opts.DeliveryMode > 0 {
		rabbitOpts = append(rabbitOpts, rabbitmq.WithPublishOptionsPersistentDelivery)
	}

	if opts.CorrelationID != "" {
		rabbitOpts = append(rabbitOpts, rabbitmq.WithPublishOptionsCorrelationID(opts.CorrelationID))
	}

	if opts.ReplyTo != "" {
		rabbitOpts = append(rabbitOpts, rabbitmq.WithPublishOptionsReplyTo(opts.ReplyTo))
	}

	if opts.Expiration != "" {
		rabbitOpts = append(rabbitOpts, rabbitmq.WithPublishOptionsExpiration(opts.Expiration))
	}

	if opts.MessageID != "" {
		rabbitOpts = append(rabbitOpts, rabbitmq.WithPublishOptionsMessageID(opts.MessageID))
	}

	if len(opts.Headers) > 0 {
		rabbitOpts = append(rabbitOpts, rabbitmq.WithPublishOptionsHeaders(opts.Headers))
	}

	// Publish the message
	err = qm.rmq.Publish(
		ctx,
		data,
		opts.Exchange,
		[]string{routingKey},
		rabbitOpts...,
	)

	if err != nil {
		publishLogger.Error("Failed to publish task", zap.Error(err))
		return fmt.Errorf("failed to publish task: %w", err)
	}

	publishLogger.Info("Task published successfully")
	return nil
}

// GetTasksExchange returns the configured tasks exchange name.
func (qm *Manager) GetTasksExchange() string {
	return qm.exchange
}

// StopConsumer stops a specific consumer.
func (qm *Manager) StopConsumer(name string) {
	if queueName, exists := qm.consumers[name]; exists {
		qm.rmq.StopConsumer(queueName)
		delete(qm.consumers, name)
		qm.logger.Info("Consumer stopped", zap.String("consumer", name))
	}
}

// StopAllConsumers stops all registered consumers.
func (qm *Manager) StopAllConsumers() {
	for name := range qm.consumers {
		qm.StopConsumer(name)
	}
	qm.logger.Info("All consumers stopped")
}

// ListConsumers returns a list of all registered consumer names.
func (qm *Manager) ListConsumers() []string {
	var consumers []string
	for name := range qm.consumers {
		consumers = append(consumers, name)
	}
	return consumers
}

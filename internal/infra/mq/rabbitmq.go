// Package mq provides message queue functionality.
package mq

import (
	"context"
	"fmt"
	"sync"

	"github.com/wagslane/go-rabbitmq"
	"go.uber.org/zap"

	"github.com/xelarion/go-layout/internal/infra/config"
)

// RabbitMQ represents a RabbitMQ client instance.
type RabbitMQ struct {
	conn       *rabbitmq.Conn
	publisher  *rabbitmq.Publisher
	consumers  map[string]*rabbitmq.Consumer
	consumerMu sync.RWMutex
	logger     *zap.Logger
}

// NewRabbitMQ creates a new RabbitMQ client instance.
func NewRabbitMQ(cfg *config.RabbitMQ, logger *zap.Logger) (*RabbitMQ, error) {
	// Create a connection with reconnection capability
	conn, err := rabbitmq.NewConn(
		cfg.URL,
		rabbitmq.WithConnectionOptionsLogger(logger.Sugar()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create RabbitMQ connection: %w", err)
	}

	// Create a publisher with reconnection capability
	// Note: We don't declare any exchanges here, as exchanges should be declared
	// when specific exchanges are needed in Publish
	publisher, err := rabbitmq.NewPublisher(
		conn,
		rabbitmq.WithPublisherOptionsLogger(logger.Sugar()),
	)
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("failed to create RabbitMQ publisher: %w", err)
	}

	return &RabbitMQ{
		conn:      conn,
		publisher: publisher,
		consumers: make(map[string]*rabbitmq.Consumer),
		logger:    logger.Named("rabbitmq"),
	}, nil
}

// Publish publishes a message to the specified routing keys.
func (r *RabbitMQ) Publish(ctx context.Context, data []byte, exchange string, routingKeys []string, options ...func(*rabbitmq.PublishOptions)) error {
	// Check for context cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Add exchange to options
	options = append([]func(*rabbitmq.PublishOptions){
		rabbitmq.WithPublishOptionsExchange(exchange),
	}, options...)

	// Publish message
	if err := r.publisher.Publish(data, routingKeys, options...); err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}
	return nil
}

// StartConsumer registers a consumer for the given queue and starts consuming messages.
func (r *RabbitMQ) StartConsumer(
	queueName string,
	handler func(d rabbitmq.Delivery) rabbitmq.Action,
	options ...func(*rabbitmq.ConsumerOptions),
) (*rabbitmq.Consumer, error) {
	// Create a consumer specific for this queue with reconnection capability
	consumer, err := rabbitmq.NewConsumer(
		r.conn,
		queueName,
		options...,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer for queue %s: %w", queueName, err)
	}

	// Store the consumer for lifecycle management before running
	r.consumerMu.Lock()
	r.consumers[queueName] = consumer
	r.consumerMu.Unlock()

	// Start consuming messages in a goroutine to avoid blocking
	go func() {
		// consumer.Run automatically handles reconnection and only returns on unrecoverable errors
		if err := consumer.Run(handler); err != nil {
			// Only log the error but don't remove the consumer from the map
			// This preserves the ability to properly close the consumer when needed
			r.logger.Error("Consumer error",
				zap.String("queue", queueName),
				zap.Error(err))
		}
	}()

	return consumer, nil
}

// StopConsumer stops a specific consumer safely.
func (r *RabbitMQ) StopConsumer(queueName string) {
	r.consumerMu.Lock()
	defer r.consumerMu.Unlock()

	if consumer, exists := r.consumers[queueName]; exists {
		consumer.Close()
		delete(r.consumers, queueName)
	}
}

// StopAllConsumers safely stops all active consumers.
func (r *RabbitMQ) StopAllConsumers() {
	r.consumerMu.Lock()
	defer r.consumerMu.Unlock()

	for _, consumer := range r.consumers {
		consumer.Close()
	}
	// Clear consumers map
	r.consumers = make(map[string]*rabbitmq.Consumer)
}

// Close properly closes all resources in the correct order as specified in the documentation.
// First close consumers, then publishers, then connection.
func (r *RabbitMQ) Close() {
	// Step 1: Stop all consumers
	r.StopAllConsumers()

	// Step 2: Close publisher
	if r.publisher != nil {
		r.publisher.Close()
	}

	// Step 3: Close connection only after all consumers and publishers are closed
	if r.conn != nil {
		_ = r.conn.Close()
	}
}

// GetPublisher returns the RabbitMQ publisher
func (r *RabbitMQ) GetPublisher() *rabbitmq.Publisher {
	return r.publisher
}

// GetConnection returns the RabbitMQ connection
func (r *RabbitMQ) GetConnection() *rabbitmq.Conn {
	return r.conn
}

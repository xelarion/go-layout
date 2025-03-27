// Package mq provides message queue functionality.
package mq

import (
	"context"
	"fmt"
	"sync"

	"github.com/wagslane/go-rabbitmq"
	"go.uber.org/zap"

	"github.com/xelarion/go-layout/pkg/config"
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
	logger = logger.Named("rabbitmq")

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
		logger:    logger,
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

	err := r.publisher.Publish(
		data,
		routingKeys,
		options...,
	)
	if err != nil {
		r.logger.Error("Failed to publish message to RabbitMQ",
			zap.String("exchange", exchange),
			zap.Strings("routing_keys", routingKeys),
			zap.Error(err))
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
		r.logger.Error("Failed to create consumer",
			zap.String("queue", queueName),
			zap.Error(err))
		return nil, fmt.Errorf("failed to create consumer: %w", err)
	}

	// Store the consumer for lifecycle management before running
	r.consumerMu.Lock()
	r.consumers[queueName] = consumer
	r.consumerMu.Unlock()

	// Start consuming messages in a goroutine to avoid blocking
	go func() {
		// consumer.Run会自动处理重连，只有当发生不可恢复的错误时才会返回
		err := consumer.Run(handler)
		if err != nil {
			// 只记录错误，但不从map中删除consumer
			// 这样保留了close时能正确关闭consumer的能力
			r.logger.Error("Consumer encountered unrecoverable error",
				zap.String("queue", queueName),
				zap.Error(err))
		} else {
			r.logger.Info("Consumer stopped normally",
				zap.String("queue", queueName))
		}
	}()

	r.logger.Info("Consumer registered successfully",
		zap.String("queue", queueName))
	return consumer, nil
}

// StopConsumer stops a specific consumer safely.
func (r *RabbitMQ) StopConsumer(queueName string) {
	r.consumerMu.Lock()
	defer r.consumerMu.Unlock()

	if consumer, exists := r.consumers[queueName]; exists {
		consumer.Close()
		delete(r.consumers, queueName)
		r.logger.Info("Consumer stopped", zap.String("queue", queueName))
	}
}

// StopAllConsumers safely stops all active consumers.
func (r *RabbitMQ) StopAllConsumers() {
	r.consumerMu.Lock()
	defer r.consumerMu.Unlock()

	for queueName, consumer := range r.consumers {
		consumer.Close()
		r.logger.Info("Consumer closed", zap.String("queue", queueName))
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
		r.logger.Info("Publisher closed")
	}

	// Step 3: Close connection only after all consumers and publishers are closed
	if r.conn != nil {
		_ = r.conn.Close()
		r.logger.Info("RabbitMQ connection closed")
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

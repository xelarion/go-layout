// Package tasks provides poller task implementations.
package tasks

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/xelarion/go-layout/internal/task"
	"github.com/xelarion/go-layout/internal/task/poller"
	"github.com/xelarion/go-layout/pkg/errs"
)

// AmqpMessagePublishHandler is an amqp-message-publish poller task handler.
type AmqpMessagePublishHandler struct {
	deps   *task.Dependencies
	logger *zap.Logger
}

// NewAmqpMessagePublishHandler creates a new amqp-message-publish handler.
func NewAmqpMessagePublishHandler(deps *task.Dependencies, logger *zap.Logger) *AmqpMessagePublishHandler {
	return &AmqpMessagePublishHandler{
		deps:   deps,
		logger: logger.Named("amqp-message-publish-handler"),
	}
}

// Register registers this handler with the poller.
func (t *AmqpMessagePublishHandler) Register(p *poller.Poller) error {
	if err := p.Register("amqp-message-publish-task", time.Second*5, t.Execute); err != nil {
		return fmt.Errorf("failed to register amqp-message-publish task handler: %w", err)
	}
	return nil
}

// Execute runs the amqp-message-publish task handler.
func (t *AmqpMessagePublishHandler) Execute(ctx context.Context) error {
	messages, _, err := t.deps.AmqpMessageRepo.List(ctx, map[string]any{}, 500, -1, "id ASC")
	if err != nil {
		return err
	}

	for _, message := range messages {
		options := message.Options.Data()
		optFuncs := options.ToOptionFuncs()
		err := t.deps.MQRepo.Publish(ctx, message.Exchange, message.RoutingKeys, message.Data, optFuncs...)
		if err != nil {
			return errs.WrapInternal(err, "failed to publish message")
		}

		// 发送成功后删除数据库记录
		err = t.deps.AmqpMessageRepo.Delete(ctx, message.ID)
		if err != nil {
			return err
		}
	}

	return nil
}

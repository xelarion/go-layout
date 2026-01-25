package repository

import (
	"context"

	"github.com/xelarion/go-layout/internal/infra/mq"
	"github.com/xelarion/go-layout/internal/usecase"
	"github.com/xelarion/go-layout/pkg/errs"
)

var _ usecase.MQRepository = (*MQRepository)(nil)

// MQRepository RabbitMQ 消息队列实现
type MQRepository struct {
	rabbitmq *mq.RabbitMQ
}

// NewMQRepository 创建消息队列仓储实例
func NewMQRepository(rabbitmq *mq.RabbitMQ) usecase.MQRepository {
	return &MQRepository{
		rabbitmq: rabbitmq,
	}
}

// Publish 发布消息到 RabbitMQ
func (r *MQRepository) Publish(ctx context.Context, exchange string, routingKeys []string, data []byte, options ...func(*mq.PublishOptions)) error {
	// 应用函数选项
	opts := &mq.PublishOptions{}
	for _, option := range options {
		option(opts)
	}

	// 转换为 RabbitMQ 选项并发布
	mqOpts := opts.ToRabbitMQOptions()
	if err := r.rabbitmq.Publish(ctx, data, exchange, routingKeys, mqOpts...); err != nil {
		return errs.WrapInternal(err, "failed to publish message to rabbitmq")
	}

	return nil
}

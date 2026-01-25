package usecase

import (
	"context"

	"github.com/xelarion/go-layout/internal/infra/mq"
)

// MQRepository 消息队列数据访问接口
// 抽象层,用于解耦具体的 MQ 实现(RabbitMQ/Kafka/etc)
type MQRepository interface {
	// Publish 发布消息到消息队列
	// 使用函数选项模式配置发布选项
	Publish(ctx context.Context, exchange string, routingKeys []string, data []byte, options ...func(*mq.PublishOptions)) error
}

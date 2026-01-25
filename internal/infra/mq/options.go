package mq

import "github.com/wagslane/go-rabbitmq"

// PublishOptions 消息发布选项
type PublishOptions struct {
	// Expiration 消息过期时间(毫秒)
	// 参见 https://www.rabbitmq.com/ttl.html#per-message-ttl-in-publishers
	Expiration string
}

// WithExpiration 设置消息过期时间
func WithExpiration(expiration string) func(*PublishOptions) {
	return func(opts *PublishOptions) {
		opts.Expiration = expiration
	}
}

// ToRabbitMQOptions 转换为 RabbitMQ 的发布选项
func (opts *PublishOptions) ToRabbitMQOptions() []func(*rabbitmq.PublishOptions) {
	var mqOpts []func(*rabbitmq.PublishOptions)

	if opts.Expiration != "" {
		mqOpts = append(mqOpts, rabbitmq.WithPublishOptionsExpiration(opts.Expiration))
	}

	return mqOpts
}

// ToOptionFuncs 将结构体转换为选项函数切片
// 用于从数据库读取的 PublishOptions 转换为函数选项
func (opts *PublishOptions) ToOptionFuncs() []func(*PublishOptions) {
	var optFuncs []func(*PublishOptions)

	if opts.Expiration != "" {
		optFuncs = append(optFuncs, WithExpiration(opts.Expiration))
	}

	return optFuncs
}

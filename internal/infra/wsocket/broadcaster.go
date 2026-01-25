package wsocket

import (
	"context"
	"encoding/json"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	// 默认 Redis 频道
	defaultChannel = "wsocket:broadcast"
)

// BroadcastRequest 广播请求（用于 Redis 传输）
type BroadcastRequest struct {
	Message   *Message `json:"message"`
	TargetAll bool     `json:"target_all"`           // 广播给所有
	SourceIDs []string `json:"source_ids,omitempty"` // 目标源ID列表
}

// Broadcaster 广播服务接口（支持分布式）
type Broadcaster interface {
	// Broadcast 广播给所有在线连接
	Broadcast(ctx context.Context, msg *Message) error
	// SendTo 发送给指定源
	SendTo(ctx context.Context, sourceID string, msg *Message) error
	// SendToMany 发送给多个源
	SendToMany(ctx context.Context, sourceIDs []string, msg *Message) error
}

// RedisBroadcaster 基于 Redis Pub/Sub 的广播实现
type RedisBroadcaster struct {
	redis   *redis.Client
	manager *Manager
	channel string
	logger  *zap.Logger
}

// BroadcasterOption 配置选项
type BroadcasterOption func(*RedisBroadcaster)

// WithSubscriber 设置订阅者（web-api 使用，task 服务不需要）
func WithSubscriber(manager *Manager) BroadcasterOption {
	return func(b *RedisBroadcaster) {
		b.manager = manager
	}
}

// NewRedisBroadcaster 创建 Redis 广播器
// 如果设置了 WithSubscriber，会自动启动订阅 goroutine
func NewRedisBroadcaster(redisClient *redis.Client, logger *zap.Logger, opts ...BroadcasterOption) *RedisBroadcaster {
	b := &RedisBroadcaster{
		redis:   redisClient,
		channel: defaultChannel,
		logger:  logger.Named("broadcaster"),
	}
	for _, opt := range opts {
		opt(b)
	}

	// 如果有 manager，自动启动订阅
	if b.manager != nil {
		go b.Subscribe(context.Background())
	}

	return b
}

// Broadcast 广播给所有在线连接
func (b *RedisBroadcaster) Broadcast(ctx context.Context, msg *Message) error {
	req := &BroadcastRequest{
		Message:   msg,
		TargetAll: true,
	}
	return b.publish(ctx, req)
}

// SendTo 发送给指定源
func (b *RedisBroadcaster) SendTo(ctx context.Context, sourceID string, msg *Message) error {
	return b.SendToMany(ctx, []string{sourceID}, msg)
}

// SendToMany 发送给多个源
func (b *RedisBroadcaster) SendToMany(ctx context.Context, sourceIDs []string, msg *Message) error {
	req := &BroadcastRequest{
		Message:   msg,
		TargetAll: false,
		SourceIDs: sourceIDs,
	}
	return b.publish(ctx, req)
}

func (b *RedisBroadcaster) publish(ctx context.Context, req *BroadcastRequest) error {
	data, err := json.Marshal(req)
	if err != nil {
		return err
	}
	return b.redis.Publish(ctx, b.channel, data).Err()
}

// Subscribe 订阅 Redis 消息（web-api 调用，需在 goroutine 中运行）
func (b *RedisBroadcaster) Subscribe(ctx context.Context) {
	if b.manager == nil {
		b.logger.Warn("manager is nil, skip subscribe")
		return
	}

	pubsub := b.redis.Subscribe(ctx, b.channel)
	defer pubsub.Close()

	b.logger.Info("subscribed to broadcast channel", zap.String("channel", b.channel))

	ch := pubsub.Channel()
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-ch:
			if !ok || msg == nil {
				// Channel closed or nil message, exit gracefully
				return
			}
			b.handleMessage(msg.Payload)
		}
	}
}

func (b *RedisBroadcaster) handleMessage(payload string) {
	var req BroadcastRequest
	if err := json.Unmarshal([]byte(payload), &req); err != nil {
		b.logger.Error("failed to unmarshal broadcast request", zap.Error(err))
		return
	}

	if req.Message == nil {
		return
	}

	if req.TargetAll {
		b.manager.Broadcast(req.Message)
	} else {
		b.manager.SendToMany(req.SourceIDs, req.Message)
	}
}

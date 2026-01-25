// Package tasks provides scheduler task implementations.
package tasks

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/xelarion/go-layout/internal/enum"
	"github.com/xelarion/go-layout/internal/infra/wsocket"
	"github.com/xelarion/go-layout/internal/task"
	"github.com/xelarion/go-layout/internal/task/scheduler"
	"github.com/xelarion/go-layout/internal/usecase"
)

// BroadcastTestHandler 广播测试任务（用于测试 WebSocket 推送）
type BroadcastTestHandler struct {
	deps   *task.Dependencies
	logger *zap.Logger
}

// NewBroadcastTestHandler 创建广播测试任务
func NewBroadcastTestHandler(deps *task.Dependencies, logger *zap.Logger) *BroadcastTestHandler {
	return &BroadcastTestHandler{
		deps:   deps,
		logger: logger.Named("broadcast-test"),
	}
}

// Register 注册任务到调度器
func (t *BroadcastTestHandler) Register(s *scheduler.Scheduler) error {
	// 每 1 分钟执行一次（测试用，生产环境可删除或调整）
	if err := s.Register("broadcast-test", "@every 1m", t.Execute); err != nil {
		return fmt.Errorf("failed to register broadcast test task: %w", err)
	}
	return nil
}

// Execute 执行广播测试
func (t *BroadcastTestHandler) Execute(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if _, err := t.deps.NotificationUseCase.Create(ctx, usecase.CreateNotificationParams{
		CreatedBy: 0,
		Title:     fmt.Sprintf("来自 task 服务的测试推送 - %s", time.Now().Format("2006-01-02 15:04:05")),
		Content:   fmt.Sprintf("来自 task 服务的测试推送 - %s", time.Now().Format("2006-01-02 15:04:05")),
		Type:      enum.NotificationTypeProductionReport,
		Category:  enum.NotificationCategoryMessage,
		Scope:     enum.NotificationScopeUser,
		UserIDs:   []uint{1, 2},
	}); err != nil {
		return err
	}

	// 发送 todo stats 通知，这样的格式 [{ type: 'work_order', name: '工单任务', count: 5 }]
	message := wsocket.NewMessage("todo:update", []interface{}{
		map[string]interface{}{
			"type":  enum.TodoStatTypeSlug[enum.TodoStatTypeWorkOrder],
			"name":  enum.TodoStatTypeText[enum.TodoStatTypeWorkOrder],
			"count": time.Now().Minute()},
	})
	_ = t.deps.Broadcaster.SendToMany(ctx, []string{"1", "2"}, message)

	return nil
}

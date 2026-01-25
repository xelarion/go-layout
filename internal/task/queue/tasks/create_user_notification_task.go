// Package tasks provides queue task implementations.
package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"go.uber.org/zap"

	"github.com/xelarion/go-layout/internal/enum"
	"github.com/xelarion/go-layout/internal/infra/wsocket"
	"github.com/xelarion/go-layout/internal/model"
	"github.com/xelarion/go-layout/internal/model/gen"
	"github.com/xelarion/go-layout/internal/task"
	"github.com/xelarion/go-layout/internal/task/queue"
	"github.com/xelarion/go-layout/internal/util"
	"github.com/xelarion/go-layout/pkg/errs"
)

// CreateUserNotificationPayload 创建用户通知任务的 payload
type CreateUserNotificationPayload struct {
	NotificationID uint `json:"notification_id"`
}

// CreateUserNotificationHandler 处理创建用户通知的任务
type CreateUserNotificationHandler struct {
	deps   *task.Dependencies
	logger *zap.Logger
}

// NewCreateUserNotificationHandler creates a new handler.
func NewCreateUserNotificationHandler(deps *task.Dependencies, logger *zap.Logger) *CreateUserNotificationHandler {
	return &CreateUserNotificationHandler{
		deps:   deps,
		logger: logger.Named("create-user-notification-handler"),
	}
}

// Register registers this handler with the queue.
func (h *CreateUserNotificationHandler) Register(q *queue.Manager) error {
	if err := q.RegisterConsumer(
		"create-user-notification-handler",
		"create-user-notification-queue",
		h.Execute,
		queue.WithConsumerOptionsConcurrency(1),
	); err != nil {
		return fmt.Errorf("failed to register create user notification handler: %w", err)
	}
	return nil
}

// Execute runs the task.
func (h *CreateUserNotificationHandler) Execute(ctx context.Context, msg queue.Message) (queue.Action, error) {
	var payload CreateUserNotificationPayload
	if err := json.Unmarshal(msg.Body, &payload); err != nil {
		return queue.NackDiscard, errs.WrapInternal(err, "failed to unmarshal payload")
	}

	// 获取通知详情（包含 scope 和 payload）
	notification, err := h.deps.NotificationRepo.FindByID(ctx, payload.NotificationID)
	if err != nil {
		if errs.IsReason(err, errs.ReasonNotFound) {
			return queue.NackDiscard, nil
		}
		h.logger.Error("failed to get notification", zap.Uint("notification_id", payload.NotificationID), zap.Error(err))
		return queue.NackRequeue, nil
	}

	var userIDs []uint

	switch notification.Scope {
	case enum.NotificationScopeAllUsers:
		// 分批获取所有用户ID并创建
		if err := h.createForAllUsers(ctx, notification.ID); err != nil {
			return queue.NackRequeue, err
		}
	case enum.NotificationScopeUser:
		// 从 notification.Payload 获取 user_ids
		userIDs = notification.Payload.Data().UserIDs
		if len(userIDs) == 0 {
			h.logger.Warn("no user_ids in notification payload for user scope")
			return queue.Ack, nil
		}
		if err := h.createUserNotifications(ctx, notification.ID, userIDs); err != nil {
			return queue.NackRequeue, err
		}
	default:
		h.logger.Error("invalid notification scope", zap.Int8("scope", int8(notification.Scope)))
		return queue.NackDiscard, nil
	}

	// 检测到通知已被删除，则删除对应的 user_notifications
	if isDeleted, err := h.deleteUserNotificationsIfNotificationDeleted(ctx, notification.ID); err != nil {
		return queue.NackRequeue, err
	} else if isDeleted {
		return queue.Ack, nil
	}

	// 推送 WebSocket 消息
	h.sendWebSocket(ctx, notification, userIDs)

	return queue.Ack, nil
}

// createForAllUsers 分批获取所有用户并创建通知
func (h *CreateUserNotificationHandler) createForAllUsers(ctx context.Context, notificationID uint) error {
	offset := 0
	fetchUserBatchSize := 1000
	for {
		// 检测到通知已被删除，则删除对应的 user_notifications
		if isDeleted, err := h.deleteUserNotificationsIfNotificationDeleted(ctx, notificationID); err != nil {
			return err
		} else if isDeleted {
			return nil
		}

		users, _, err := h.deps.UserRepo.List(ctx, nil, fetchUserBatchSize, offset, "id ASC")
		if err != nil {
			return err
		}
		if len(users) == 0 {
			break
		}

		userIDs := make([]uint, 0, len(users))
		for _, user := range users {
			userIDs = append(userIDs, user.ID)
		}

		if err := h.createUserNotifications(ctx, notificationID, userIDs); err != nil {
			return err
		}

		if len(users) < fetchUserBatchSize {
			break
		}
		offset += fetchUserBatchSize
	}
	return nil
}

// createUserNotifications 批量创建 user_notifications
func (h *CreateUserNotificationHandler) createUserNotifications(ctx context.Context, notificationID uint, userIDs []uint) error {
	userNotifications := make([]*model.UserNotification, 0, len(userIDs))
	for _, userID := range userIDs {
		// 如果已经创建过，则跳过
		hasUserNotification, err := h.deps.UserNotificationRepo.IsExists(ctx, map[string]any{"notification_id": notificationID, "user_id": userID}, nil)
		if err != nil {
			return err
		}

		if !hasUserNotification {
			userNotifications = append(userNotifications, &model.UserNotification{
				UserNotification: gen.UserNotification{
					NotificationID: notificationID,
					UserID:         userID,
				},
			})
		}
	}
	return h.deps.UserNotificationRepo.CreateInBatches(ctx, userNotifications, 500)
}

// 如果检测到通知已被删除，则删除对应的 user_notifications
func (h *CreateUserNotificationHandler) deleteUserNotificationsIfNotificationDeleted(ctx context.Context, notificationID uint) (isDeleted bool, err error) {
	if _, err := h.deps.NotificationRepo.FindByID(ctx, notificationID); err != nil {
		if errs.IsReason(err, errs.ReasonNotFound) {
			return true, h.deps.UserNotificationRepo.DeleteAllByNotificationID(ctx, notificationID)
		}
		return false, err
	}

	return false, nil
}

// sendWebSocket 推送 WebSocket 消息，匹配前端 notification:new 事件
func (h *CreateUserNotificationHandler) sendWebSocket(ctx context.Context, notification *model.Notification, userIDs []uint) {
	wsMsg := wsocket.NewMessage("notification:new", map[string]any{
		"id":            notification.ID,
		"title":         notification.Title,
		"category_slug": enum.NotificationCategorySlug[notification.Category],
		"is_read":       false,
		"created_at":    util.FormatDateTime(notification.CreatedAt),
	})

	var err error
	if len(userIDs) == 0 {
		// 全员通知：广播给所有在线用户
		err = h.deps.Broadcaster.Broadcast(ctx, wsMsg)
	} else {
		// 指定用户：发送给特定用户
		sourceIDs := make([]string, 0, len(userIDs))
		for _, uid := range userIDs {
			sourceIDs = append(sourceIDs, strconv.FormatUint(uint64(uid), 10))
		}
		err = h.deps.Broadcaster.SendToMany(ctx, sourceIDs, wsMsg)
	}

	if err != nil {
		h.logger.Warn("failed to send notification via websocket", zap.Error(err))
	}
}

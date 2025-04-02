// Package tasks provides queue task implementations.
package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/xelarion/go-layout/internal/task/queue"
)

// NotificationPayload represents a notification sending task.
type NotificationPayload struct {
	UserID  string `json:"user_id"`
	Title   string `json:"title"`
	Message string `json:"message"`
}

// NotificationHandler handles notification processing.
type NotificationHandler struct {
	manager *queue.Manager
	logger  *zap.Logger
}

// NewNotificationHandler creates a new notification task handler.
func NewNotificationHandler(manager *queue.Manager, logger *zap.Logger) *NotificationHandler {
	return &NotificationHandler{
		manager: manager,
		logger:  logger.Named("notification-handler"),
	}
}

// Register registers this task with the queue manager.
func (t *NotificationHandler) Register() error {
	// Create consumer for notification tasks
	err := t.manager.RegisterConsumer(
		"notification-handler",
		"notification-handler",
		"notification",
		queue.ConvertHandlerFunc(t.Execute),
	)
	if err != nil {
		return fmt.Errorf("failed to register notification task consumer: %w", err)
	}
	return nil
}

// Publish publishes a notification task to the queue.
func (t *NotificationHandler) Publish(ctx context.Context, userID, title, message string) error {
	payload := NotificationPayload{
		UserID:  userID,
		Title:   title,
		Message: message,
	}

	return t.manager.PublishTask(ctx, "notification", payload,
		WithHighPriority(), WithPersistentDelivery())
}

// Execute handles notification sending tasks.
func (t *NotificationHandler) Execute(ctx context.Context, msg []byte) error {
	var payload NotificationPayload
	if err := json.Unmarshal(msg, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal notification payload: %w", err)
	}

	t.logger.Info("Processing notification task",
		zap.String("user_id", payload.UserID),
		zap.String("title", payload.Title),
		zap.String("message", payload.Message))

	// Simulate notification sending
	time.Sleep(300 * time.Millisecond)

	t.logger.Info("Notification sent successfully")
	return nil
}

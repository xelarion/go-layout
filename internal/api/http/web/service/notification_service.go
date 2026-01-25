package service

import (
	"context"

	"github.com/xelarion/go-layout/internal/api/http/web/types"
	appctx "github.com/xelarion/go-layout/internal/context"
	"github.com/xelarion/go-layout/internal/enum"
	"github.com/xelarion/go-layout/internal/usecase"
	"github.com/xelarion/go-layout/internal/util"
	"github.com/xelarion/go-layout/pkg/errs"
)

// NotificationService handles notification-related services.
type NotificationService struct {
	*BaseService
	notificationUseCase *usecase.NotificationUseCase
}

// NewNotificationService creates a new NotificationService.
func NewNotificationService(
	baseService *BaseService,
	notificationUseCase *usecase.NotificationUseCase,
) *NotificationService {
	return &NotificationService{
		BaseService:         baseService,
		notificationUseCase: notificationUseCase,
	}
}

// ListNotifications lists notifications for current user.
func (s *NotificationService) ListNotifications(ctx context.Context, req *types.ListNotificationsReq) (*types.ListNotificationsResp, error) {
	current := appctx.GetCurrent(ctx)
	if current == nil {
		return nil, errs.NewBusiness("unauthorized").WithReason(errs.ReasonUnauthorized)
	}

	filters := map[string]any{
		"user_id": current.UserID,
	}

	if req.Key != "" {
		filters["key"] = req.Key
	}

	if req.Type != "" {
		typ, err := enum.GetNotificationType(req.Type)
		if err != nil {
			return nil, err
		}
		filters["type"] = typ
	}

	if req.Category != "" {
		category, err := enum.GetNotificationCategory(req.Category)
		if err != nil {
			return nil, err
		}
		filters["category"] = category
	}

	if req.IsRead != nil {
		filters["is_read"] = *req.IsRead
	}

	notifications, count, err := s.notificationUseCase.List(ctx, filters, req.GetLimit(), req.GetOffset(), req.GetSortClause(), current.UserID)
	if err != nil {
		return nil, err
	}

	respResults := make([]*types.ListNotificationsRespResult, 0, len(notifications))
	for _, notification := range notifications {
		result := &types.ListNotificationsRespResult{
			ID:           notification.ID,
			Title:        notification.Title,
			Type:         enum.NotificationTypeText[notification.Type],
			TypeSlug:     enum.NotificationTypeSlug[notification.Type],
			Category:     enum.NotificationCategoryText[notification.Category],
			CategorySlug: enum.NotificationCategorySlug[notification.Category],
			IsRead:       notification.IsRead,
			CreatedAt:    util.FormatDateTime(notification.CreatedAt),
		}
		respResults = append(respResults, result)
	}

	return &types.ListNotificationsResp{
		Results:    respResults,
		Pagination: types.NewPageResp(count, req.GetPage(), req.GetPageSize()),
	}, nil
}

// GetRecentNotifications get recent notifications for current user.
func (s *NotificationService) GetRecentNotifications(ctx context.Context, req *types.GetRecentNotificationsReq) (*types.GetRecentNotificationsResp, error) {
	current := appctx.GetCurrent(ctx)
	if current == nil {
		return nil, errs.NewBusiness("unauthorized").WithReason(errs.ReasonUnauthorized)
	}

	filters := map[string]any{
		"user_id": current.UserID,
	}

	if req.IsRead != nil {
		filters["is_read"] = *req.IsRead
	}

	categoryResults := make([]*types.GetRecentNotificationsRespCategoryResult, 0, len(enum.NotificationCategorySlug))
	for category, categorySlug := range enum.NotificationCategorySlug {
		filters["category"] = category
		notifications, count, err := s.notificationUseCase.List(ctx, filters, 10, 0, "id DESC", current.UserID)
		if err != nil {
			return nil, err
		}

		notificationResults := make([]*types.GetRecentNotificationsRespCategoryResultResult, 0, len(notifications))
		for _, notification := range notifications {
			result := &types.GetRecentNotificationsRespCategoryResultResult{
				ID:           notification.ID,
				Title:        notification.Title,
				CategorySlug: enum.NotificationCategorySlug[notification.Category],
				IsRead:       notification.IsRead,
				CreatedAt:    util.FormatDateTime(notification.CreatedAt),
			}
			notificationResults = append(notificationResults, result)
		}

		categoryResults = append(categoryResults, &types.GetRecentNotificationsRespCategoryResult{
			CategorySlug: categorySlug,
			Count:        count,
			Results:      notificationResults,
		})
	}

	return &types.GetRecentNotificationsResp{
		CategoryResults: categoryResults,
	}, nil
}

// GetNotification retrieves a notification by ID.
func (s *NotificationService) GetNotification(ctx context.Context, req *types.GetNotificationReq) (*types.GetNotificationResp, error) {
	current := appctx.GetCurrent(ctx)
	if current == nil {
		return nil, errs.NewBusiness("unauthorized").WithReason(errs.ReasonUnauthorized)
	}

	notification, err := s.notificationUseCase.GetAndRead(ctx, req.ID, current.UserID)
	if err != nil {
		return nil, err
	}

	return &types.GetNotificationResp{
		ID:        notification.ID,
		Title:     notification.Title,
		Content:   notification.Content,
		Type:      enum.NotificationTypeSlug[notification.Type],
		IsRead:    notification.IsRead,
		CreatedAt: util.FormatDateTime(notification.CreatedAt),
	}, nil
}

// DeleteNotification handles notification deletion.
func (s *NotificationService) DeleteNotification(ctx context.Context, req *types.DeleteNotificationReq) (*types.DeleteNotificationResp, error) {
	current := appctx.GetCurrent(ctx)
	if current == nil {
		return nil, errs.NewBusiness("unauthorized").WithReason(errs.ReasonUnauthorized)
	}

	if err := s.notificationUseCase.Delete(ctx, req.ID, current.UserID); err != nil {
		return nil, err
	}

	return &types.DeleteNotificationResp{}, nil
}

// BatchDeleteNotifications handles batch notification deletion.
func (s *NotificationService) BatchDeleteNotifications(ctx context.Context, req *types.BatchDeleteNotificationsReq) (*types.BatchDeleteNotificationsResp, error) {
	current := appctx.GetCurrent(ctx)
	if current == nil {
		return nil, errs.NewBusiness("unauthorized").WithReason(errs.ReasonUnauthorized)
	}

	results := s.notificationUseCase.BatchDelete(ctx, req.IDs, current.UserID)

	successCount := 0
	failedItems := make([]string, 0)
	for _, result := range results {
		if result.Success {
			successCount++
		} else {
			errMsg := result.Identifier + ": " + errs.GetMessage(result.Error)
			failedItems = append(failedItems, errMsg)
		}
	}

	return &types.BatchDeleteNotificationsResp{
		SuccessCount: successCount,
		FailedCount:  len(failedItems),
		FailedItems:  failedItems,
	}, nil
}

// BatchReadNotifications handles batch notification read.
func (s *NotificationService) BatchReadNotifications(ctx context.Context, req *types.BatchReadNotificationsReq) (*types.BatchReadNotificationsResp, error) {
	current := appctx.GetCurrent(ctx)
	if current == nil {
		return nil, errs.NewBusiness("unauthorized").WithReason(errs.ReasonUnauthorized)
	}

	results := s.notificationUseCase.BatchRead(ctx, req.IDs, current.UserID)

	successCount := 0
	failedItems := make([]string, 0)
	for _, result := range results {
		if result.Success {
			successCount++
		} else {
			errMsg := result.Identifier + ": " + errs.GetMessage(result.Error)
			failedItems = append(failedItems, errMsg)
		}
	}

	return &types.BatchReadNotificationsResp{
		SuccessCount: successCount,
		FailedCount:  len(failedItems),
		FailedItems:  failedItems,
	}, nil
}

// ReadAllNotifications handles read all notifications.
func (s *NotificationService) ReadAllNotifications(ctx context.Context, req *types.ReadAllNotificationsReq) (*types.ReadAllNotificationsResp, error) {
	current := appctx.GetCurrent(ctx)
	if current == nil {
		return nil, errs.NewBusiness("unauthorized").WithReason(errs.ReasonUnauthorized)
	}

	var category *enum.NotificationCategory
	if req.Category != "" {
		if val, err := enum.GetNotificationCategory(req.Category); err != nil {
			return nil, err
		} else {
			category = &val
		}
	}

	if err := s.notificationUseCase.ReadAll(ctx, current.UserID, category); err != nil {
		return nil, err
	}

	return &types.ReadAllNotificationsResp{}, nil
}

// GetTodoStats gets todo statistics
func (s *NotificationService) GetTodoStats(ctx context.Context, req *types.GetTodoStatsReq) (*types.GetTodoStatsResp, error) {
	current := appctx.GetCurrent(ctx)
	if current == nil {
		return nil, errs.NewBusiness("unauthorized").WithReason(errs.ReasonUnauthorized)
	}

	stats, err := s.notificationUseCase.GetTodoStats(ctx, current.UserID)
	if err != nil {
		return nil, err
	}

	resp := &types.GetTodoStatsResp{}
	resp.Results = make([]types.GetTodoStatsRespResult, len(stats))
	for i, stat := range stats {
		resp.Results[i] = types.GetTodoStatsRespResult{
			Type:  enum.TodoStatTypeSlug[stat.Type],
			Name:  stat.Name,
			Count: stat.Count,
		}
	}

	return resp, nil
}

// Package usecase contains business logic.
package usecase

import (
	"context"
	"fmt"

	"github.com/xelarion/go-layout/internal/enum"
	"github.com/xelarion/go-layout/internal/model"
	"github.com/xelarion/go-layout/internal/model/gen"
	"github.com/xelarion/go-layout/internal/util"
	"github.com/xelarion/go-layout/pkg/errs"
)

// Announcement contains announcement data
type Announcement struct {
	*model.Announcement
	Publisher string
}

// CreateAnnouncementParams contains all parameters needed to create a announcement
type CreateAnnouncementParams struct {
	CreateRecordCommonParams
	Title   string
	Content string
	Status  enum.AnnouncementStatus
}

// UpdateAnnouncementParams contains all parameters needed to update a announcement
type UpdateAnnouncementParams struct {
	UpdateRecordCommonParams
	ID      uint
	Title   string
	Content string
	Status  enum.AnnouncementStatus

	// Fields to track which values are explicitly set
	TitleSet   bool
	ContentSet bool
	StatusSet  bool
}

// BatchDeleteAnnouncementsResult represents the result of a single announcement deletion.
type BatchDeleteAnnouncementsResult = BatchItemCommonResult

type AnnouncementPublisher struct {
	ID       uint
	FullName string
}

// AnnouncementRepository defines methods for announcement data access.
type AnnouncementRepository interface {
	Create(ctx context.Context, announcement *model.Announcement) error
	List(ctx context.Context, filters map[string]any, limit, offset int, sortClause string) ([]*model.Announcement, int, error)
	IsExists(ctx context.Context, filters map[string]any, notFilters map[string]any) (bool, error)
	Count(ctx context.Context, filters map[string]any, notFilters map[string]any) (int64, error)
	FindByID(ctx context.Context, id uint) (*model.Announcement, error)
	FindByIDForUpdate(ctx context.Context, id uint) (*model.Announcement, error)
	Update(ctx context.Context, announcement *model.Announcement, params map[string]any) error
	Delete(ctx context.Context, id uint) error
	ListPublishers(ctx context.Context, filters map[string]any, limit, offset int, sortClause string) ([]*AnnouncementPublisher, int, error)
}

// AnnouncementUseCase implements business logic for announcement operations.
type AnnouncementUseCase struct {
	tx                  Transaction
	announcementRepo    AnnouncementRepository
	userRepo            UserRepository
	notificationUseCase *NotificationUseCase
	storageUseCase      *StorageUseCase
}

// NewAnnouncementUseCase creates a new instance of AnnouncementUseCase.
func NewAnnouncementUseCase(
	tx Transaction,
	repo AnnouncementRepository,
	userRepo UserRepository,
	notificationUseCase *NotificationUseCase,
	storageUseCase *StorageUseCase,
) *AnnouncementUseCase {
	return &AnnouncementUseCase{
		tx:                  tx,
		announcementRepo:    repo,
		userRepo:            userRepo,
		notificationUseCase: notificationUseCase,
		storageUseCase:      storageUseCase,
	}
}

// Create creates a new announcement.
func (uc *AnnouncementUseCase) Create(ctx context.Context, params CreateAnnouncementParams) (uint, error) {
	announcement := &model.Announcement{
		Announcement: gen.Announcement{
			CreatedBy:     params.CreatedBy,
			UpdatedBy:     params.CreatedBy,
			CreatedRoleID: params.CreatedRoleID,
			Title:         params.Title,
			Content:       params.Content,
		},
		Status: params.Status,
	}

	if params.Status == enum.AnnouncementStatusPublished {
		announcement.PublisherID = &params.CreatedBy
		announcement.PublishedAt = util.DBTimeNowPointer()
	}

	err := uc.tx.Transaction(ctx, func(ctx context.Context) error {
		if err := uc.announcementRepo.Create(ctx, announcement); err != nil {
			return err
		}

		// 为富文本内容中的 blob 创建 attachment
		if err := uc.storageUseCase.AttachContentBlobs(ctx, announcement, params.Content); err != nil {
			return err
		}

		// 如果直接发布，创建通知
		if params.Status == enum.AnnouncementStatusPublished {
			if _, err := uc.notificationUseCase.Create(ctx, CreateNotificationParams{
				CreatedBy:      params.CreatedBy,
				Title:          announcement.Title,
				Content:        announcement.Content,
				Type:           enum.NotificationTypeAnnouncement,
				Category:       enum.NotificationCategoryNotice,
				Scope:          enum.NotificationScopeAllUsers,
				AnnouncementID: announcement.ID,
			}); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	return announcement.ID, nil
}

// List returns a list of announcements with pagination and filtering.
func (uc *AnnouncementUseCase) List(ctx context.Context, filters map[string]any, limit, offset int, sortClause string) ([]*Announcement, int, error) {
	announcements, count, err := uc.announcementRepo.List(ctx, filters, limit, offset, sortClause)
	if err != nil {
		return nil, 0, err
	}

	records := make([]*Announcement, 0, len(announcements))
	for _, announcement := range announcements {
		publisher := ""
		if announcement.PublisherID != nil {
			if user, err := uc.userRepo.FindByID(ctx, *announcement.PublisherID); err != nil {
				return nil, 0, err
			} else {
				publisher = user.FullName
			}
		}

		record := &Announcement{
			Announcement: announcement,
			Publisher:    publisher,
		}
		records = append(records, record)
	}

	return records, count, nil
}

// GetByID retrieves announcement by ID.
func (uc *AnnouncementUseCase) GetByID(ctx context.Context, id uint) (*Announcement, error) {
	announcement, err := uc.announcementRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &Announcement{
		Announcement: announcement,
	}, nil
}

// Update updates an existing announcement.
func (uc *AnnouncementUseCase) Update(ctx context.Context, params UpdateAnnouncementParams) error {
	updates := map[string]any{}

	if params.TitleSet {
		updates["title"] = params.Title
	}

	if params.ContentSet {
		updates["content"] = params.Content
	}

	isPublishing := params.StatusSet && params.Status == enum.AnnouncementStatusPublished
	if params.StatusSet {
		updates["status"] = params.Status

		if isPublishing {
			updates["publisher_id"] = params.UpdatedBy
			updates["published_at"] = util.DBTimeNow()
		}
	}

	if len(updates) > 0 {
		updates["updated_by"] = params.UpdatedBy
		err := uc.tx.Transaction(ctx, func(ctx context.Context) error {
			// 使用 FOR UPDATE 锁定记录，防止状态更新错误
			announcement, err := uc.announcementRepo.FindByIDForUpdate(ctx, params.ID)
			if err != nil {
				return err
			}

			if announcement.IsPublished() {
				return errs.NewBusiness("通知公告已经发布，无法更新")
			}

			if err := uc.announcementRepo.Update(ctx, announcement, updates); err != nil {
				return err
			}

			// 如果内容更新，重新处理富文本附件
			if params.ContentSet {
				if err := uc.storageUseCase.AttachContentBlobs(ctx, announcement, params.Content); err != nil {
					return err
				}
			}

			// 如果更新为发布状态，创建通知
			if isPublishing {
				if _, err := uc.notificationUseCase.Create(ctx, CreateNotificationParams{
					CreatedBy:      params.UpdatedBy,
					Title:          announcement.Title,
					Content:        announcement.Content,
					Type:           enum.NotificationTypeAnnouncement,
					Category:       enum.NotificationCategoryNotice,
					Scope:          enum.NotificationScopeAllUsers,
					AnnouncementID: announcement.ID,
				}); err != nil {
					return err
				}
			}

			return nil
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// Delete removes a announcement.
func (uc *AnnouncementUseCase) Delete(ctx context.Context, id uint) error {
	announcement, err := uc.announcementRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	return uc.tx.Transaction(ctx, func(ctx context.Context) error {
		if err := uc.storageUseCase.Detach(ctx, announcement); err != nil {
			return err
		}

		if err := uc.announcementRepo.Delete(ctx, id); err != nil {
			return err
		}

		return uc.notificationUseCase.DeleteByAnnouncementID(ctx, id)
	})
}

// BatchDelete removes multiple announcements and returns detailed results.
func (uc *AnnouncementUseCase) BatchDelete(ctx context.Context, ids []uint) []BatchDeleteAnnouncementsResult {
	results := make([]BatchDeleteAnnouncementsResult, 0, len(ids))

	for _, id := range ids {
		result := BatchDeleteAnnouncementsResult{
			ID:      id,
			Success: false,
		}

		announcement, err := uc.announcementRepo.FindByID(ctx, id)
		if err != nil {
			result.Identifier = fmt.Sprintf("ID %d 未知记录", id)
			result.Error = err
			results = append(results, result)
			continue
		}

		result.Identifier = announcement.Title

		err = uc.Delete(ctx, id)
		if err != nil {
			result.Error = err
		} else {
			result.Success = true
		}

		results = append(results, result)
	}

	return results
}

// GetAnnouncementPublishersOptions 获取所有公告发布者作为选项数据
func (uc *AnnouncementUseCase) GetAnnouncementPublishersOptions(ctx context.Context, filters map[string]any, sortClause string) ([]*AnnouncementPublisher, error) {
	publishers, _, err := uc.announcementRepo.ListPublishers(ctx, filters, -1, -1, sortClause)
	if err != nil {
		return nil, err
	}

	records := make([]*AnnouncementPublisher, 0, len(publishers))
	for _, publisher := range publishers {
		record := &AnnouncementPublisher{
			ID:       publisher.ID,
			FullName: publisher.FullName,
		}
		records = append(records, record)
	}

	return records, nil
}

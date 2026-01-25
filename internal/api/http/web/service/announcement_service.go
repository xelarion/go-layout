package service

import (
	"context"
	"fmt"

	"github.com/xelarion/go-layout/internal/api/http/web/types"
	appctx "github.com/xelarion/go-layout/internal/context"
	"github.com/xelarion/go-layout/internal/enum"
	"github.com/xelarion/go-layout/internal/model"
	"github.com/xelarion/go-layout/internal/usecase"
	"github.com/xelarion/go-layout/internal/util"
	"github.com/xelarion/go-layout/pkg/errs"
)

// AnnouncementService handles announcement-related services.
type AnnouncementService struct {
	*BaseService
	announcementUseCase *usecase.AnnouncementUseCase
	exportUseCase       *usecase.ExportUseCase
}

// NewAnnouncementService creates a new AnnouncementService.
func NewAnnouncementService(
	baseService *BaseService,
	announcementUseCase *usecase.AnnouncementUseCase,
	exportUseCase *usecase.ExportUseCase,
) *AnnouncementService {
	return &AnnouncementService{
		BaseService:         baseService,
		announcementUseCase: announcementUseCase,
		exportUseCase:       exportUseCase,
	}
}

// CreateAnnouncement registers a new announcement.
func (s *AnnouncementService) CreateAnnouncement(ctx context.Context, req *types.CreateAnnouncementReq) (*types.CreateAnnouncementResp, error) {
	current := appctx.GetCurrent(ctx)
	if current == nil {
		return nil, errs.NewBusiness("unauthorized").WithReason(errs.ReasonUnauthorized)
	}

	status, err := enum.GetAnnouncementStatus(req.Status)
	if err != nil {
		return nil, err
	}

	params := usecase.CreateAnnouncementParams{
		CreateRecordCommonParams: usecase.CreateRecordCommonParams{
			CreatedBy:     current.UserID,
			CreatedRoleID: current.RoleID,
		},
		Title:   req.Title,
		Content: req.Content,
		Status:  status,
	}

	id, err := s.announcementUseCase.Create(ctx, params)
	if err != nil {
		return nil, err
	}

	return &types.CreateAnnouncementResp{
		ID: id,
	}, nil
}

// ListAnnouncements lists announcements as a tree without pagination.
func (s *AnnouncementService) ListAnnouncements(ctx context.Context, req *types.ListAnnouncementsReq) (*types.ListAnnouncementsResp, error) {
	filters := map[string]any{}
	if req.Key != "" {
		filters["key"] = req.Key
	}

	if req.PublisherID != 0 {
		filters["publisher_id"] = req.PublisherID
	}

	if req.PublishedAtStart != "" {
		filters["published_at_start"] = util.BeginningOfMinute(util.ParseStringToBaseTime(req.PublishedAtStart, "2006-01-02 15:04")).UTC()
	}

	if req.PublishedAtEnd != "" {
		filters["published_at_end"] = util.EndOfMinute(util.ParseStringToBaseTime(req.PublishedAtEnd, "2006-01-02 15:04")).UTC()
	}

	if req.Status != "" {
		status, err := enum.GetAnnouncementStatus(req.Status)
		if err != nil {
			return nil, err
		}
		filters["status"] = status
	}

	if err := s.ApplyDataScope(ctx, filters, &model.Announcement{}); err != nil {
		return nil, err
	}

	announcements, count, err := s.announcementUseCase.List(ctx, filters, req.GetLimit(), req.GetOffset(), req.GetSortClause())
	if err != nil {
		return nil, err
	}

	respResults := make([]*types.ListAnnouncementsRespResult, 0, len(announcements))
	for _, announcement := range announcements {
		result := &types.ListAnnouncementsRespResult{
			ID:          announcement.ID,
			Title:       announcement.Title,
			Status:      enum.AnnouncementStatusText[announcement.Status],
			StatusSlug:  enum.AnnouncementStatusSlug[announcement.Status],
			Publisher:   announcement.Publisher,
			PublishedAt: util.FormatDateTimePointer(announcement.PublishedAt),
		}
		respResults = append(respResults, result)
	}

	return &types.ListAnnouncementsResp{
		Results:    respResults,
		Pagination: types.NewPageResp(count, req.GetPage(), req.GetPageSize()),
	}, nil
}

// GetAnnouncement retrieves a announcement by ID.
func (s *AnnouncementService) GetAnnouncement(ctx context.Context, req *types.GetAnnouncementReq) (*types.GetAnnouncementResp, error) {
	announcement, err := s.announcementUseCase.GetByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	if err := s.CheckRecordAccess(ctx, announcement.Announcement); err != nil {
		return nil, err
	}

	return &types.GetAnnouncementResp{
		ID:          announcement.ID,
		Title:       announcement.Title,
		Content:     announcement.Content,
		PublishedAt: util.FormatDateTimePointer(announcement.PublishedAt),
	}, nil
}

// GetAnnouncementFormData provides data needed for announcement forms (update).
func (s *AnnouncementService) GetAnnouncementFormData(ctx context.Context, req *types.GetAnnouncementFormDataReq) (*types.GetAnnouncementFormDataResp, error) {
	announcement, err := s.announcementUseCase.GetByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	if announcement.IsPublished() {
		return nil, errs.NewBusiness("通知公告已经发布，无法编辑")
	}

	if err := s.CheckRecordAccess(ctx, announcement.Announcement); err != nil {
		return nil, err
	}

	return &types.GetAnnouncementFormDataResp{
		ID:      announcement.ID,
		Title:   announcement.Title,
		Content: announcement.Content,
	}, nil
}

// UpdateAnnouncement updates a announcement.
func (s *AnnouncementService) UpdateAnnouncement(ctx context.Context, req *types.UpdateAnnouncementReq) (*types.UpdateAnnouncementResp, error) {
	current := appctx.GetCurrent(ctx)
	if current == nil {
		return nil, errs.NewBusiness("unauthorized").WithReason(errs.ReasonUnauthorized)
	}

	announcement, err := s.announcementUseCase.GetByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	if err := s.CheckRecordAccess(ctx, announcement.Announcement); err != nil {
		return nil, err
	}

	params := usecase.UpdateAnnouncementParams{
		UpdateRecordCommonParams: usecase.UpdateRecordCommonParams{
			UpdatedBy: current.UserID,
		},
		ID: req.ID,
	}

	params.Title = req.Title
	params.TitleSet = true

	params.Content = req.Content
	params.ContentSet = true

	status, err := enum.GetAnnouncementStatus(req.Status)
	if err != nil {
		return nil, err
	}
	params.Status = status
	params.StatusSet = true

	if err := s.announcementUseCase.Update(ctx, params); err != nil {
		return nil, err
	}

	return &types.UpdateAnnouncementResp{}, nil
}

// DeleteAnnouncement handles announcement deletion.
func (s *AnnouncementService) DeleteAnnouncement(ctx context.Context, req *types.DeleteAnnouncementReq) (*types.DeleteAnnouncementResp, error) {
	announcement, err := s.announcementUseCase.GetByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	if err := s.CheckRecordAccess(ctx, announcement.Announcement); err != nil {
		return nil, err
	}

	if err := s.announcementUseCase.Delete(ctx, req.ID); err != nil {
		return nil, err
	}

	return &types.DeleteAnnouncementResp{}, nil
}

// BatchDeleteAnnouncements handles batch announcement deletion.
func (s *AnnouncementService) BatchDeleteAnnouncements(ctx context.Context, req *types.BatchDeleteAnnouncementsReq) (*types.BatchDeleteAnnouncementsResp, error) {
	accessibleIDs := make([]uint, 0, len(req.IDs))
	failedItems := make([]string, 0)

	for _, id := range req.IDs {
		announcement, err := s.announcementUseCase.GetByID(ctx, id)
		if err != nil {
			failedItems = append(failedItems, fmt.Sprintf("ID %d: 记录不存在或已被删除", id))
			continue
		}

		if err := s.CheckRecordAccess(ctx, announcement.Announcement); err != nil {
			failedItems = append(failedItems, announcement.Title+": 无权限删除")
			continue
		}

		accessibleIDs = append(accessibleIDs, id)
	}

	results := s.announcementUseCase.BatchDelete(ctx, accessibleIDs)

	successCount := 0
	for _, result := range results {
		if result.Success {
			successCount++
		} else {
			errMsg := result.Identifier + ": " + errs.GetMessage(result.Error)
			failedItems = append(failedItems, errMsg)
		}
	}

	return &types.BatchDeleteAnnouncementsResp{
		SuccessCount: successCount,
		FailedCount:  len(failedItems),
		FailedItems:  failedItems,
	}, nil
}

func (s *AnnouncementService) GetAnnouncementPublishersOptions(ctx context.Context, req *types.GetAnnouncementPublishersOptionsReq) (*types.GetAnnouncementPublishersOptionsResp, error) {
	filters := map[string]any{}

	publishers, err := s.announcementUseCase.GetAnnouncementPublishersOptions(ctx, filters, req.GetSortClause())
	if err != nil {
		return nil, err
	}

	results := make([]types.GetAnnouncementPublishersOptionsRespResult, 0, len(publishers))
	for _, publisher := range publishers {
		results = append(results, types.GetAnnouncementPublishersOptionsRespResult{
			GetOptionsRespResult: types.GetOptionsRespResult{
				Label: publisher.FullName,
				Value: publisher.ID,
			},
		})
	}

	return &types.GetAnnouncementPublishersOptionsResp{
		Results: results,
	}, nil
}

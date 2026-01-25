package service

import (
	"context"

	"go.uber.org/zap"

	"github.com/xelarion/go-layout/internal/api/http/web/types"
	appctx "github.com/xelarion/go-layout/internal/context"
	"github.com/xelarion/go-layout/internal/enum"
	"github.com/xelarion/go-layout/internal/infra/config"
	"github.com/xelarion/go-layout/internal/usecase"
	"github.com/xelarion/go-layout/internal/util"
	"github.com/xelarion/go-layout/pkg/errs"
)

// ExportService 导出服务
type ExportService struct {
	*BaseService
	exportUseCase *usecase.ExportUseCase
	config        *config.Export
	logger        *zap.Logger
}

// NewExportService 创建导出服务实例
func NewExportService(
	baseService *BaseService,
	exportUseCase *usecase.ExportUseCase,
	cfg *config.Export,
	logger *zap.Logger,
) *ExportService {
	return &ExportService{
		BaseService:   baseService,
		exportUseCase: exportUseCase,
		config:        cfg,
		logger:        logger,
	}
}

// ListExportTasks 查询导出任务列表
func (s *ExportService) ListExportTasks(ctx context.Context, req *types.ListExportTasksReq) (*types.ListExportTasksResp, error) {
	current := appctx.GetCurrent(ctx)
	if current == nil {
		return nil, errs.NewBusiness("unauthorized").WithReason(errs.ReasonUnauthorized)
	}

	// 构建过滤条件
	filters := map[string]any{
		"created_by": current.UserID, // 只能查看自己创建的任务
	}

	if req.TaskType != "" {
		filters["task_type"] = req.TaskType
	}

	if req.Status != "" {
		if status, err := enum.GetExportTaskStatus(req.Status); err != nil {
			return nil, err
		} else {
			filters["status"] = status
		}
	}

	// 查询任务列表
	tasks, count, err := s.exportUseCase.ListExportTasks(ctx, filters, req.GetLimit(), req.GetOffset(), req.GetSortClause())
	if err != nil {
		return nil, err
	}

	// 转换为响应类型
	results := make([]types.ExportTaskResult, 0, len(tasks))
	for _, task := range tasks {
		result := types.ExportTaskResult{
			ID:           task.ID,
			TaskType:     enum.ExportTaskTypeText[task.TaskType],
			Filename:     task.Filename,
			Status:       enum.ExportTaskStatusText[task.Status],
			StatusSlug:   enum.ExportTaskStatusSlug[task.Status],
			IsAsync:      task.IsAsync,
			TotalCount:   int(task.TotalCount),
			FileSize:     task.FileSize,
			ErrorMessage: task.ErrorMessage,
			CreatedAt:    util.FormatDateTime(task.CreatedAt),
			StartedAt:    util.FormatDateTimePointer(task.StartedAt),
			CompletedAt:  util.FormatDateTimePointer(task.CompletedAt),
		}

		results = append(results, result)
	}

	return &types.ListExportTasksResp{
		Results:    results,
		Pagination: types.NewPageResp(count, req.GetPage(), req.GetPageSize()),
	}, nil
}

// GetExportFileDownloadURL 获取导出文件下载URL
func (s *ExportService) GetExportFileDownloadURL(ctx context.Context, req *types.GetExportFileDownloadURLReq) (*types.GetExportFileDownloadURLResp, error) {
	current := appctx.GetCurrent(ctx)
	if current == nil {
		return nil, errs.NewBusiness("unauthorized").WithReason(errs.ReasonUnauthorized)
	}

	url, err := s.exportUseCase.GetExportFileDownloadURL(ctx, req.ID, current.UserID)
	if err != nil {
		return nil, err
	}

	return &types.GetExportFileDownloadURLResp{
		URL: url,
	}, nil
}

// DeleteExportTask 删除导出任务
func (s *ExportService) DeleteExportTask(ctx context.Context, req *types.DeleteExportTaskReq) (*types.DeleteExportTaskResp, error) {
	current := appctx.GetCurrent(ctx)
	if current == nil {
		return nil, errs.NewBusiness("unauthorized").WithReason(errs.ReasonUnauthorized)
	}

	if err := s.exportUseCase.DeleteExportTask(ctx, req.ID, current.UserID); err != nil {
		return nil, err
	}

	return &types.DeleteExportTaskResp{}, nil
}

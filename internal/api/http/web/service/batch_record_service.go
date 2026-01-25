package service

import (
	"context"

	"go.uber.org/zap"

	"github.com/xelarion/go-layout/internal/api/http/web/types"
	"github.com/xelarion/go-layout/internal/enum"
	"github.com/xelarion/go-layout/internal/model"
	"github.com/xelarion/go-layout/internal/usecase"
	"github.com/xelarion/go-layout/internal/util"
	"github.com/xelarion/go-layout/pkg/errs"
)

type BatchRecordService struct {
	*BaseService       // 嵌入 BaseService 提供数据权限检查方法
	batchRecordUseCase *usecase.BatchRecordUseCase
	logger             *zap.Logger
}

func NewBatchRecordService(baseService *BaseService, batchRecordUseCase *usecase.BatchRecordUseCase, logger *zap.Logger) *BatchRecordService {
	return &BatchRecordService{
		BaseService:        baseService,
		batchRecordUseCase: batchRecordUseCase,
		logger:             logger,
	}
}

func (s *BatchRecordService) ListBatchRecords(ctx context.Context, req *types.ListBatchRecordsReq) (*types.ListBatchRecordsResp, error) {
	filters := map[string]any{}

	if req.CreatedBy != 0 {
		filters["created_by"] = req.CreatedBy
	}

	if req.OperationType != "" {
		operationType, err := enum.GetBatchRecordOperationType(req.OperationType)
		if err != nil {
			return nil, err
		}
		filters["operation_type"] = operationType
	}

	if req.Status != "" {
		status, err := enum.GetBatchRecordStatus(req.Status)
		if err != nil {
			return nil, err
		}
		filters["status"] = status
	}

	if req.CreatedAtStart != "" {
		filters["created_at_start"] = util.BeginningOfMinute(util.ParseStringToBaseTime(req.CreatedAtStart, "2006-01-02 15:04")).UTC()
	}

	if req.CreatedAtEnd != "" {
		filters["created_at_end"] = util.EndOfMinute(util.ParseStringToBaseTime(req.CreatedAtEnd, "2006-01-02 15:04")).UTC()
	}

	// 应用数据权限过滤（传入空的 BatchRecord 实例用于获取字段映射）
	if err := s.ApplyDataScope(ctx, filters, &model.BatchRecord{}); err != nil {
		return nil, err
	}

	batchRecords, count, err := s.batchRecordUseCase.List(ctx, filters, req.GetLimit(), req.GetOffset(), req.GetSortClause())
	if err != nil {
		return nil, err
	}

	respResults := make([]types.ListBatchRecordsRespResult, 0, len(batchRecords))
	for _, batchRecord := range batchRecords {
		u := types.ListBatchRecordsRespResult{
			ID:                batchRecord.ID,
			CreatedAt:         util.FormatDateTime(batchRecord.CreatedAt),
			CreatedBy:         batchRecord.CreatedByFullName,
			OperationType:     enum.BatchRecordOperationTypeText[batchRecord.OperationType],
			OperationTypeSlug: enum.BatchRecordOperationTypeSlug[batchRecord.OperationType],
			Status:            enum.BatchRecordStatusText[batchRecord.Status],
			StatusSlug:        enum.BatchRecordStatusSlug[batchRecord.Status],
			FailureReason:     batchRecord.FailureReason,
		}
		respResults = append(respResults, u)
	}

	return &types.ListBatchRecordsResp{
		Results:    respResults,
		Pagination: types.NewPageResp(count, req.GetPage(), req.GetPageSize()),
	}, nil
}

func (s *BatchRecordService) GetBatchRecord(ctx context.Context, req *types.GetBatchRecordReq) (*types.GetBatchRecordResp, error) {
	batchRecord, err := s.batchRecordUseCase.GetByID(ctx, req.ID)
	if err != nil {
		return nil, errs.WrapInternal(err, "failed to get batch record")
	}

	return &types.GetBatchRecordResp{
		ID:                batchRecord.ID,
		CreatedAt:         util.FormatDateTime(batchRecord.CreatedAt),
		CreatedBy:         batchRecord.CreatedByFullName,
		OperationType:     enum.BatchRecordOperationTypeText[batchRecord.OperationType],
		OperationTypeSlug: enum.BatchRecordOperationTypeSlug[batchRecord.OperationType],
		Status:            enum.BatchRecordStatusText[batchRecord.Status],
		StatusSlug:        enum.BatchRecordStatusSlug[batchRecord.Status],
		FailureReason:     batchRecord.FailureReason,
		ExtraData:         batchRecord.ExtraData,
	}, nil
}

func (s *BatchRecordService) GetBatchRecordOperationTypeOptions(ctx context.Context, req *types.GetBatchRecordOperationTypeOptionsReq) (*types.GetBatchRecordOperationTypeOptionsResp, error) {
	results := make([]types.GetBatchRecordOperationTypeOptionsRespResult, 0, len(enum.BatchRecordOperationTypeText))
	for operationType, operationTypeText := range enum.BatchRecordOperationTypeText {
		results = append(results, types.GetBatchRecordOperationTypeOptionsRespResult{
			GetOptionsRespResult: types.GetOptionsRespResult{
				Label: operationTypeText,
				Value: enum.BatchRecordOperationTypeSlug[operationType],
			},
		})
	}

	return &types.GetBatchRecordOperationTypeOptionsResp{
		Results: results,
	}, nil
}

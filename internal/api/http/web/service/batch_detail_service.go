package service

import (
	"context"

	"github.com/xelarion/go-layout/internal/util"

	"go.uber.org/zap"

	"github.com/xelarion/go-layout/internal/api/http/web/types"
	"github.com/xelarion/go-layout/internal/enum"
	"github.com/xelarion/go-layout/internal/usecase"
)

type BatchDetailService struct {
	batchDetailUseCase *usecase.BatchDetailUseCase
	logger             *zap.Logger
}

func NewBatchDetailService(batchDetailUseCase *usecase.BatchDetailUseCase, logger *zap.Logger) *BatchDetailService {
	return &BatchDetailService{
		batchDetailUseCase: batchDetailUseCase,
		logger:             logger,
	}
}

func (s *BatchDetailService) ListBatchDetails(ctx context.Context, req *types.ListBatchDetailsReq) (*types.ListBatchDetailsResp, error) {
	filters := map[string]any{}
	if req.Status != "" {
		status, err := enum.GetBatchDetailStatus(req.Status)
		if err != nil {
			return nil, err
		}
		filters["status"] = status
	}

	if req.Key != "" {
		filters["key"] = req.Key
	}

	batchDetails, count, err := s.batchDetailUseCase.List(ctx, req.BatchRecordID, filters, req.GetLimit(), req.GetOffset(), req.GetSortClause())
	if err != nil {
		return nil, err
	}

	respResults := make([]types.ListBatchDetailsRespResult, 0, len(batchDetails))
	for _, batchDetail := range batchDetails {
		u := types.ListBatchDetailsRespResult{
			ID:            batchDetail.ID,
			CreatedAt:     util.FormatDateTime(batchDetail.CreatedAt),
			Key:           batchDetail.Key,
			Status:        enum.BatchDetailStatusText[batchDetail.Status],
			StatusSlug:    enum.BatchDetailStatusSlug[batchDetail.Status],
			FailureReason: batchDetail.FailureReason,
			ExtraData:     batchDetail.ExtraData,
		}
		respResults = append(respResults, u)
	}

	return &types.ListBatchDetailsResp{
		Results:    respResults,
		Pagination: types.NewPageResp(count, req.GetPage(), req.GetPageSize()),
	}, nil
}

// Package usecase contains business logic.
package usecase

import (
	"context"

	"gorm.io/datatypes"

	"github.com/xelarion/go-layout/internal/enum"
	"github.com/xelarion/go-layout/internal/model"
	"github.com/xelarion/go-layout/internal/model/gen"
)

// BatchDetail contains batch_detail data
type BatchDetail struct {
	*model.BatchDetail
	ExtraData map[string]any
}

// CreateBatchDetailParams defines parameters for creating a batch_detail.
type CreateBatchDetailParams struct {
	BatchRecordID uint
	Key           string
	Data          datatypes.JSON
	Status        enum.BatchDetailStatus
	FailureReason string
}

// BatchDetailRepository defines methods for batch_detail data access.
type BatchDetailRepository interface {
	Create(ctx context.Context, batchDetail *model.BatchDetail) error
	List(ctx context.Context, filters map[string]any, limit, offset int, sortClause string) ([]*model.BatchDetail, int, error)
	IsExists(ctx context.Context, filters map[string]any, notFilters map[string]any) (bool, error)
	Count(ctx context.Context, filters map[string]any, notFilters map[string]any) (int64, error)
	FindByID(ctx context.Context, id uint) (*model.BatchDetail, error)
	Update(ctx context.Context, batchDetail *model.BatchDetail, params map[string]any) error
	Delete(ctx context.Context, id uint) error
}

// BatchDetailUseCase implements business logic for batch_detail operations.
type BatchDetailUseCase struct {
	tx              Transaction
	batchDetailRepo BatchDetailRepository
	batchRecordRepo BatchRecordRepository
}

// NewBatchDetailUseCase creates a new instance of BatchDetailUseCase.
func NewBatchDetailUseCase(tx Transaction, repo BatchDetailRepository, batchRecordRepo BatchRecordRepository) *BatchDetailUseCase {
	return &BatchDetailUseCase{
		tx:              tx,
		batchDetailRepo: repo,
		batchRecordRepo: batchRecordRepo,
	}
}

// List returns a list of batch_details with pagination and filtering.
func (uc *BatchDetailUseCase) List(ctx context.Context, batchRecordID uint, filters map[string]any, limit, offset int, sortClause string) ([]*BatchDetail, int, error) {
	batchRecord, err := uc.batchRecordRepo.FindByID(ctx, batchRecordID)
	if err != nil {
		return nil, 0, err
	}

	filters["batch_record_id"] = batchRecord.ID
	batchDetails, count, err := uc.batchDetailRepo.List(ctx, filters, limit, offset, sortClause)
	if err != nil {
		return nil, 0, err
	}

	records := make([]*BatchDetail, 0, len(batchDetails))
	for _, batchDetail := range batchDetails {
		extraData := make(map[string]any)
		switch batchRecord.OperationType {
		default: // None data for example
			extraData, err = uc.getExtraDataNone(ctx, batchDetail)
			if err != nil {
				return nil, 0, err
			}
		}

		records = append(records, &BatchDetail{
			BatchDetail: batchDetail,
			ExtraData:   extraData,
		})
	}

	return records, count, nil
}

// GetByID retrieves a batch_detail by ID.
func (uc *BatchDetailUseCase) GetByID(ctx context.Context, id uint) (*BatchDetail, error) {
	batchDetail, err := uc.batchDetailRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &BatchDetail{
		BatchDetail: batchDetail,
	}, nil
}

// Create creates a new batch_detail.
func (uc *BatchDetailUseCase) Create(ctx context.Context, params CreateBatchDetailParams) (uint, error) {
	batchDetail := &model.BatchDetail{
		BatchDetail: gen.BatchDetail{
			BatchRecordID: params.BatchRecordID,
			Key:           params.Key,
			Data:          params.Data,
			FailureReason: params.FailureReason,
		},
		Status: params.Status,
	}
	err := uc.tx.Transaction(ctx, func(ctx context.Context) error {
		if err := uc.batchDetailRepo.Create(ctx, batchDetail); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return 0, err
	}

	return batchDetail.ID, nil
}

// getExtraDataNone returns the extra data for a batch_record with None data type.
func (uc *BatchDetailUseCase) getExtraDataNone(ctx context.Context, batchDetail *model.BatchDetail) (map[string]any, error) {
	data := &model.BatchDetailDataNone{}
	if err := json.Unmarshal(batchDetail.Data, data); err != nil {
		return nil, err
	}

	return map[string]any{
		// "example_field": data.ExampleField,
	}, nil
}

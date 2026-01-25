// Package usecase contains business logic.
package usecase

import (
	"context"

	"github.com/xelarion/go-layout/internal/enum"
	"github.com/xelarion/go-layout/internal/model"
	"github.com/xelarion/go-layout/internal/model/gen"
	"github.com/xelarion/go-layout/pkg/errs"
)

// BatchRecord contains batch_record data
type BatchRecord struct {
	*model.BatchRecord
	CreatedByFullName string
	ExtraData         map[string]any
}

// BatchRecordRepository defines methods for batch_record data access.
type BatchRecordRepository interface {
	Create(ctx context.Context, batchRecord *model.BatchRecord) error
	List(ctx context.Context, filters map[string]any, limit, offset int, sortClause string) ([]*model.BatchRecord, int, error)
	IsExists(ctx context.Context, filters map[string]any, notFilters map[string]any) (bool, error)
	Count(ctx context.Context, filters map[string]any, notFilters map[string]any) (int64, error)
	FindByID(ctx context.Context, id uint) (*model.BatchRecord, error)
	FindByIDForUpdate(ctx context.Context, id uint) (*model.BatchRecord, error)
	Update(ctx context.Context, batchRecord *model.BatchRecord, params map[string]any) error
	Delete(ctx context.Context, id uint) error
}

// BatchRecordUseCase implements business logic for batch_record operations.
type BatchRecordUseCase struct {
	tx              Transaction
	batchRecordRepo BatchRecordRepository
	batchDetailRepo BatchDetailRepository
	userRepo        UserRepository
}

// NewBatchRecordUseCase creates a new instance of BatchRecordUseCase.
func NewBatchRecordUseCase(tx Transaction, repo BatchRecordRepository, batchDetailRepo BatchDetailRepository, userRepo UserRepository) *BatchRecordUseCase {
	return &BatchRecordUseCase{
		tx:              tx,
		batchRecordRepo: repo,
		batchDetailRepo: batchDetailRepo,
		userRepo:        userRepo,
	}
}

// List returns a list of batch_records with pagination and filtering.
func (uc *BatchRecordUseCase) List(ctx context.Context, filters map[string]any, limit, offset int, sortClause string) ([]*BatchRecord, int, error) {
	batchRecords, count, err := uc.batchRecordRepo.List(ctx, filters, limit, offset, sortClause)
	if err != nil {
		return nil, 0, err
	}

	records := make([]*BatchRecord, 0, len(batchRecords))
	for _, batchRecord := range batchRecords {
		user, err := uc.userRepo.FindByID(ctx, batchRecord.CreatedBy)
		if err != nil {
			return nil, 0, err
		}

		records = append(records, &BatchRecord{
			BatchRecord:       batchRecord,
			CreatedByFullName: user.FullName,
			ExtraData:         nil, // 列表不需要
		})
	}

	return records, count, nil
}

// GetByID retrieves a batch_record by ID.
func (uc *BatchRecordUseCase) GetByID(ctx context.Context, id uint) (*BatchRecord, error) {
	batchRecord, err := uc.batchRecordRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	user, err := uc.userRepo.FindByID(ctx, batchRecord.CreatedBy)
	if err != nil {
		return nil, err
	}

	extraData := make(map[string]any)
	switch batchRecord.OperationType {
	default: // None data for example
		extraData, err = uc.getExtraDataNone(ctx, batchRecord)
		if err != nil {
			return nil, err
		}
	}

	return &BatchRecord{
		BatchRecord:       batchRecord,
		CreatedByFullName: user.FullName,
		ExtraData:         extraData,
	}, nil
}

// CreatePendingBatchRecord creates a new pending batch record.
func (uc *BatchRecordUseCase) CreatePendingBatchRecord(
	ctx context.Context,
	createdBy uint,
	createdRoleID uint,
	operationType enum.BatchRecordOperationType,
	data []byte,
) (*model.BatchRecord, error) {
	batchRecord := &model.BatchRecord{
		BatchRecord: gen.BatchRecord{
			CreatedBy:     createdBy,
			CreatedRoleID: createdRoleID,
			Data:          data,
			FailureReason: "",
		},
		Status:        enum.BatchRecordStatusPending,
		OperationType: operationType,
	}

	if err := uc.batchRecordRepo.Create(ctx, batchRecord); err != nil {
		return nil, err
	}

	return batchRecord, nil
}

// Start 启动批次记录,将状态从 Pending 更新为 Processing
// 使用事务和 FOR UPDATE 锁防止并发问题
func (uc *BatchRecordUseCase) Start(ctx context.Context, id uint) error {
	return uc.tx.Transaction(ctx, func(txCtx context.Context) error {
		// 使用 FOR UPDATE 锁定记录
		batchRecord, err := uc.batchRecordRepo.FindByIDForUpdate(txCtx, id)
		if err != nil {
			return err
		}

		// 检查状态是否为 Pending
		if batchRecord.Status != enum.BatchRecordStatusPending {
			return errs.NewBusiness("批次记录状态不是待处理,无法启动").
				WithReason(errs.ReasonInvalidState).
				WithMeta("id", id).
				WithMeta("current_status", enum.BatchRecordStatusText[batchRecord.Status])
		}

		// 更新状态为 Processing
		return uc.batchRecordRepo.Update(txCtx, batchRecord, map[string]any{
			"status": enum.BatchRecordStatusProcessing,
		})
	})
}

// Finish 完成批次记录,将状态从 Processing 更新为 Success 或 Failed
// 使用事务和 FOR UPDATE 锁防止并发问题
func (uc *BatchRecordUseCase) Finish(ctx context.Context, id uint, status enum.BatchRecordStatus, failureReason string) error {
	// 验证目标状态
	if status != enum.BatchRecordStatusSuccess && status != enum.BatchRecordStatusFailed {
		return errs.NewBusiness("状态不合法").
			WithReason(errs.ReasonInvalidState).
			WithMeta("id", id).
			WithMeta("status", enum.BatchRecordStatusText[status])
	}

	return uc.tx.Transaction(ctx, func(txCtx context.Context) error {
		// 使用 FOR UPDATE 锁定记录
		batchRecord, err := uc.batchRecordRepo.FindByIDForUpdate(txCtx, id)
		if err != nil {
			return err
		}

		// 检查状态是否为 Processing
		if batchRecord.Status != enum.BatchRecordStatusProcessing {
			return errs.NewBusiness("批次记录状态不是处理中,无法完成").
				WithReason(errs.ReasonInvalidState).
				WithMeta("id", id).
				WithMeta("current_status", enum.BatchRecordStatusText[batchRecord.Status])
		}

		// 更新状态和失败原因
		updateParams := map[string]any{
			"status": status,
		}
		if failureReason != "" {
			updateParams["failure_reason"] = failureReason
		}

		return uc.batchRecordRepo.Update(txCtx, batchRecord, updateParams)
	})
}

// UpdateResult 根据批量详情记录成功失败数量更新批次记录结果
func (uc *BatchRecordUseCase) UpdateResult(ctx context.Context, id uint) error {
	failureCount, err := uc.batchDetailRepo.Count(ctx, map[string]any{"batch_record_id": id, "status": enum.BatchDetailStatusFailed}, nil)
	if err != nil {
		return err
	}

	status := enum.BatchRecordStatusSuccess
	if failureCount > 0 {
		status = enum.BatchRecordStatusFailed
	}

	return uc.Finish(ctx, id, status, "")
}

// getExtraDataNone returns the extra data for a batch_record with None data type.
func (uc *BatchRecordUseCase) getExtraDataNone(ctx context.Context, batchRecord *model.BatchRecord) (map[string]any, error) {
	data := &model.BatchRecordDataNone{}
	if err := json.Unmarshal(batchRecord.Data, data); err != nil {
		return nil, err
	}

	return map[string]any{
		// "example_field": data.ExampleField,
	}, nil
}

// Package usecase contains business logic.
package usecase

import (
	"context"
	"fmt"

	"github.com/xelarion/go-layout/internal/model"
)

// ChangeLog contains change_log data
type ChangeLog struct {
	*model.ChangeLog
	OperatorDisplay string
	Details         []*model.ChangeLogDetail // 变更细节
}

// ChangeLogRepository defines methods for change_log data access.
type ChangeLogRepository interface {
	Create(ctx context.Context, changeLog *model.ChangeLog) error
	List(ctx context.Context, filters map[string]any, limit, offset int, sortClause string) ([]*model.ChangeLog, int, error)
	IsExists(ctx context.Context, filters map[string]any, notFilters map[string]any) (bool, error)
	Count(ctx context.Context, filters map[string]any, notFilters map[string]any) (int64, error)
	FindByID(ctx context.Context, id uint) (*model.ChangeLog, error)
	FindByResourceTypeAndResourceID(ctx context.Context, resourceType string, resourceID uint) (*model.ChangeLog, error)
	Update(ctx context.Context, changeLog *model.ChangeLog, params map[string]any) error
	Delete(ctx context.Context, id uint) error
}

// ChangeLogUseCase implements business logic for change_log operations.
type ChangeLogUseCase struct {
	tx                  Transaction
	changeLogRepo       ChangeLogRepository
	changeLogDetailRepo ChangeLogDetailRepository
	userRepo            UserRepository
}

// NewChangeLogUseCase creates a new instance of ChangeLogUseCase.
func NewChangeLogUseCase(
	tx Transaction,
	repo ChangeLogRepository,
	changeLogDetailRepo ChangeLogDetailRepository,
	userRepo UserRepository,
) *ChangeLogUseCase {
	return &ChangeLogUseCase{
		tx:                  tx,
		changeLogRepo:       repo,
		changeLogDetailRepo: changeLogDetailRepo,
		userRepo:            userRepo,
	}
}

// List returns a list of change_logs with pagination and filtering.
func (uc *ChangeLogUseCase) List(ctx context.Context, filters map[string]any, limit, offset int, sortClause string) ([]*ChangeLog, int, error) {
	changeLogs, count, err := uc.changeLogRepo.List(ctx, filters, limit, offset, sortClause)
	if err != nil {
		return nil, 0, err
	}

	records := make([]*ChangeLog, 0, len(changeLogs))
	for _, changeLog := range changeLogs {
		operatorDisplay := fmt.Sprintf("%d", changeLog.OperatorID)
		if changeLog.Operator != nil && changeLog.Operator.FullName != "" {
			operatorDisplay = changeLog.Operator.FullName
		}

		// 查询变更细节
		details, _, err := uc.changeLogDetailRepo.List(ctx, map[string]any{"change_log_id": changeLog.ID}, -1, 0, "change_log_details.id ASC")
		if err != nil {
			return nil, 0, err
		}

		record := &ChangeLog{
			ChangeLog:       changeLog,
			OperatorDisplay: operatorDisplay,
			Details:         details,
		}
		records = append(records, record)
	}

	return records, count, nil
}

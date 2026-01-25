// Package usecase contains business logic.
package usecase

import (
	"context"

	"github.com/xelarion/go-layout/internal/model"
	"github.com/xelarion/go-layout/internal/model/gen"
	"github.com/xelarion/go-layout/pkg/errs"
)

// Position contains position data
type Position struct {
	*model.Position
	OrganizationName string
	UserCount        int64
}

// CreatePositionParams contains all parameters needed to create a position
type CreatePositionParams struct {
	CreateRecordCommonParams
	Code           string
	Name           string
	OrganizationID uint
	Remark         string
}

// UpdatePositionParams contains all parameters needed to update a position
type UpdatePositionParams struct {
	UpdateRecordCommonParams
	ID             uint
	Code           string
	Name           string
	OrganizationID uint
	Remark         string

	// Fields to track which values are explicitly set
	CodeSet           bool
	NameSet           bool
	OrganizationIDSet bool
	RemarkSet         bool
}

// BatchDeletePositionsResult represents the result of a single position deletion.
type BatchDeletePositionsResult = BatchItemCommonResult

// PositionRepository defines methods for position data access.
type PositionRepository interface {
	Create(ctx context.Context, position *model.Position) error
	List(ctx context.Context, filters map[string]any, limit, offset int, sortClause string) ([]*model.Position, int, error)
	IsExists(ctx context.Context, filters map[string]any, notFilters map[string]any) (bool, error)
	Count(ctx context.Context, filters map[string]any, notFilters map[string]any) (int64, error)
	FindByID(ctx context.Context, id uint) (*model.Position, error)
	Update(ctx context.Context, position *model.Position, params map[string]any) error
	Delete(ctx context.Context, id uint) error
}

// PositionUseCase implements business logic for position operations.
type PositionUseCase struct {
	tx               Transaction
	positionRepo     PositionRepository
	userRepo         UserRepository
	organizationRepo OrganizationRepository
}

// NewPositionUseCase creates a new instance of PositionUseCase.
func NewPositionUseCase(tx Transaction, repo PositionRepository, userRepo UserRepository, organizationRepo OrganizationRepository) *PositionUseCase {
	return &PositionUseCase{
		tx:               tx,
		positionRepo:     repo,
		userRepo:         userRepo,
		organizationRepo: organizationRepo,
	}
}

// Create creates a new position.
func (uc *PositionUseCase) Create(ctx context.Context, params CreatePositionParams) (uint, error) {
	// Check if position already exists
	if exists, err := uc.positionRepo.IsExists(ctx, map[string]any{"code": params.Code}, nil); err != nil {
		return 0, err
	} else if exists {
		return 0, errs.NewBusiness("岗位编号已存在").
			WithReason(errs.ReasonDuplicate)
	}

	// Check if position already exists
	if exists, err := uc.positionRepo.IsExists(ctx, map[string]any{"name": params.Name}, nil); err != nil {
		return 0, err
	} else if exists {
		return 0, errs.NewBusiness("岗位名称已存在").
			WithReason(errs.ReasonDuplicate)
	}

	position := &model.Position{
		Position: gen.Position{
			CreatedBy:      params.CreatedBy,
			UpdatedBy:      params.CreatedBy,
			CreatedRoleID:  params.CreatedRoleID, // 设置创建时的角色
			Name:           params.Name,
			Code:           params.Code,
			OrganizationID: params.OrganizationID,
			Remark:         params.Remark,
		},
	}
	err := uc.tx.Transaction(ctx, func(ctx context.Context) error {
		if err := uc.positionRepo.Create(ctx, position); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return 0, err
	}

	return position.ID, nil
}

// List returns a list of positions with pagination and filtering.
func (uc *PositionUseCase) List(ctx context.Context, filters map[string]any, limit, offset int, sortClause string) ([]*Position, int, error) {
	positions, count, err := uc.positionRepo.List(ctx, filters, limit, offset, sortClause)
	if err != nil {
		return nil, 0, err
	}

	records := make([]*Position, 0, len(positions))
	for _, position := range positions {
		userCount, err := uc.userRepo.Count(ctx, map[string]any{"position_id": position.ID}, nil)
		if err != nil {
			return nil, 0, err
		}

		organization, err := uc.organizationRepo.FindByID(ctx, position.OrganizationID)
		if err != nil {
			return nil, 0, err
		}

		records = append(records, &Position{
			Position:         position,
			OrganizationName: organization.Name,
			UserCount:        userCount,
		})
	}

	return records, count, nil
}

// GetByID retrieves a position by ID.
func (uc *PositionUseCase) GetByID(ctx context.Context, id uint) (*Position, error) {
	position, err := uc.positionRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &Position{
		Position: position,
	}, nil
}

// Update updates an existing position.
func (uc *PositionUseCase) Update(ctx context.Context, params UpdatePositionParams) error {
	// Get the existing position
	position, err := uc.positionRepo.FindByID(ctx, params.ID)
	if err != nil {
		return err
	}

	updates := map[string]any{}

	// Update fields that are explicitly set
	if params.CodeSet {
		// Check if position already exists
		exists, err := uc.positionRepo.IsExists(ctx, map[string]any{"code": params.Code}, map[string]any{"id": params.ID})
		if err != nil {
			return err
		}

		if exists {
			return errs.NewBusiness("岗位编号已存在").
				WithReason(errs.ReasonDuplicate)
		}

		updates["code"] = params.Code
	}

	if params.NameSet {
		// Check if position already exists
		exists, err := uc.positionRepo.IsExists(ctx, map[string]any{"name": params.Name}, map[string]any{"id": params.ID})
		if err != nil {
			return err
		}

		if exists {
			return errs.NewBusiness("岗位名称已存在").
				WithReason(errs.ReasonDuplicate)
		}

		updates["name"] = params.Name
	}

	if params.OrganizationIDSet {
		updates["organization_id"] = params.OrganizationID
	}

	if params.RemarkSet {
		updates["remark"] = params.Remark
	}

	if len(updates) > 0 {
		updates["updated_by"] = params.UpdatedBy
		err = uc.tx.Transaction(ctx, func(ctx context.Context) error {
			if err := uc.positionRepo.Update(ctx, position, updates); err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// Delete removes a position.
func (uc *PositionUseCase) Delete(ctx context.Context, id uint) error {
	// Check if position has users
	userExists, err := uc.userRepo.IsExists(ctx, map[string]any{"position_id": id}, nil)
	if err != nil {
		return err
	}

	if userExists {
		return errs.NewBusiness("岗位下存在用户，无法删除").
			WithReason(errs.ReasonInvalidState)
	}

	return uc.tx.Transaction(ctx, func(ctx context.Context) error {
		if err := uc.positionRepo.Delete(ctx, id); err != nil {
			return err
		}
		return nil
	})
}

// BatchDelete removes multiple positions and returns detailed results.
func (uc *PositionUseCase) BatchDelete(ctx context.Context, ids []uint) []BatchDeletePositionsResult {
	results := make([]BatchDeletePositionsResult, 0, len(ids))

	for _, id := range ids {
		result := BatchDeletePositionsResult{
			ID:      id,
			Success: false,
		}

		// Get position info first for better error messages
		position, err := uc.positionRepo.FindByID(ctx, id)
		if err != nil {
			result.Identifier = "未知岗位"
			result.Error = err
			results = append(results, result)
			continue
		}

		result.Identifier = position.Name

		// Attempt to delete
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

// GetPositionOptions 获取岗位作为选项数据
func (uc *PositionUseCase) GetPositionOptions(ctx context.Context, filters map[string]any, sortClause string) ([]*Position, error) {
	positions, _, err := uc.positionRepo.List(ctx, filters, -1, -1, sortClause)
	if err != nil {
		return nil, err
	}

	records := make([]*Position, 0, len(positions))
	for _, position := range positions {
		organization, err := uc.organizationRepo.FindByID(ctx, position.OrganizationID)
		if err != nil {
			return nil, err
		}

		records = append(records, &Position{
			Position:         position,
			OrganizationName: organization.Name,
		})
	}

	return records, nil
}

// Package usecase contains business logic.
package usecase

import (
	"context"

	"github.com/xelarion/go-layout/internal/enum"
	"github.com/xelarion/go-layout/internal/model"
	"github.com/xelarion/go-layout/internal/model/gen"
	"github.com/xelarion/go-layout/pkg/errs"
)

// Organization contains organization data
type Organization struct {
	*model.Organization
	UserCount int64
	OwnerName string
}

// CreateOrganizationParams contains all parameters needed to create a organization
type CreateOrganizationParams struct {
	CreateRecordCommonParams
	Code     string
	Name     string
	ParentID *uint
	Type     enum.OrganizationType
	OwnerID  uint
	Remark   string
}

// UpdateOrganizationParams contains all parameters needed to update a organization
type UpdateOrganizationParams struct {
	UpdateRecordCommonParams
	ID       uint
	Code     string
	Name     string
	ParentID *uint
	Type     enum.OrganizationType
	OwnerID  uint
	Remark   string

	// Fields to track which values are explicitly set
	CodeSet     bool
	NameSet     bool
	ParentIDSet bool
	TypeSet     bool
	OwnerIDSet  bool
	RemarkSet   bool
}

// OrganizationRepository defines methods for organization data access.
type OrganizationRepository interface {
	Create(ctx context.Context, organization *model.Organization) error
	List(ctx context.Context, filters map[string]any, limit, offset int, sortClause string) ([]*model.Organization, int, error)
	IsExists(ctx context.Context, filters map[string]any, notFilters map[string]any) (bool, error)
	Count(ctx context.Context, filters map[string]any, notFilters map[string]any) (int64, error)
	FindByID(ctx context.Context, id uint) (*model.Organization, error)
	Update(ctx context.Context, organization *model.Organization, params map[string]any) error
	Delete(ctx context.Context, id uint) error
}

// OrganizationUseCase implements business logic for organization operations.
type OrganizationUseCase struct {
	tx               Transaction
	organizationRepo OrganizationRepository
	positionRepo     PositionRepository
	userRepo         UserRepository
}

// NewOrganizationUseCase creates a new instance of OrganizationUseCase.
func NewOrganizationUseCase(
	tx Transaction,
	repo OrganizationRepository,
	positionRepo PositionRepository,
	userRepo UserRepository,
) *OrganizationUseCase {
	return &OrganizationUseCase{
		tx:               tx,
		organizationRepo: repo,
		positionRepo:     positionRepo,
		userRepo:         userRepo,
	}
}

// Create creates a new organization.
func (uc *OrganizationUseCase) Create(ctx context.Context, params CreateOrganizationParams) (uint, error) {
	// Check if organization already exists
	if exists, err := uc.organizationRepo.IsExists(ctx, map[string]any{"code": params.Code}, nil); err != nil {
		return 0, err
	} else if exists {
		return 0, errs.NewBusiness("组织编号已存在").
			WithReason(errs.ReasonDuplicate)
	}

	// Check if organization already exists
	if exists, err := uc.organizationRepo.IsExists(ctx, map[string]any{"name": params.Name}, nil); err != nil {
		return 0, err
	} else if exists {
		return 0, errs.NewBusiness("组织名称已存在").
			WithReason(errs.ReasonDuplicate)
	}

	organization := &model.Organization{
		Organization: gen.Organization{
			CreatedBy:     params.CreatedBy,
			UpdatedBy:     params.CreatedBy,
			CreatedRoleID: params.CreatedRoleID, // 设置创建时的角色
			Name:          params.Name,
			Code:          params.Code,
			ParentID:      params.ParentID,
			OwnerID:       params.OwnerID,
			Remark:        params.Remark,
		},
		Type: params.Type,
	}
	err := uc.tx.Transaction(ctx, func(ctx context.Context) error {
		if err := uc.organizationRepo.Create(ctx, organization); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return 0, err
	}

	return organization.ID, nil
}

// List returns a list of organizations with pagination and filtering.
func (uc *OrganizationUseCase) List(ctx context.Context, filters map[string]any, limit, offset int, sortClause string) ([]*Organization, int, error) {
	organizations, count, err := uc.organizationRepo.List(ctx, filters, limit, offset, sortClause)
	if err != nil {
		return nil, 0, err
	}

	records := make([]*Organization, 0, len(organizations))
	for _, organization := range organizations {
		userCount, err := uc.userRepo.CountByOrganizationID(ctx, organization.ID)
		if err != nil {
			return nil, 0, err
		}

		owner, err := uc.userRepo.FindByID(ctx, organization.OwnerID)
		if err != nil {
			return nil, 0, err
		}

		records = append(records, &Organization{
			Organization: organization,
			UserCount:    userCount,
			OwnerName:    owner.FullName,
		})
	}

	return records, count, nil
}

// GetByID retrieves a organization by ID.
func (uc *OrganizationUseCase) GetByID(ctx context.Context, id uint) (*Organization, error) {
	organization, err := uc.organizationRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &Organization{
		Organization: organization,
	}, nil
}

// Update updates an existing organization.
func (uc *OrganizationUseCase) Update(ctx context.Context, params UpdateOrganizationParams) error {
	// Get the existing organization
	organization, err := uc.organizationRepo.FindByID(ctx, params.ID)
	if err != nil {
		return err
	}

	updates := map[string]any{}

	// Update fields that are explicitly set
	if params.CodeSet {
		// Check if organization already exists
		exists, err := uc.organizationRepo.IsExists(ctx, map[string]any{"code": params.Code}, map[string]any{"id": params.ID})
		if err != nil {
			return err
		}

		if exists {
			return errs.NewBusiness("组织编号已存在").
				WithReason(errs.ReasonDuplicate)
		}

		updates["code"] = params.Code
	}

	if params.NameSet {
		// Check if organization already exists
		exists, err := uc.organizationRepo.IsExists(ctx, map[string]any{"name": params.Name}, map[string]any{"id": params.ID})
		if err != nil {
			return err
		}

		if exists {
			return errs.NewBusiness("组织名称已存在").
				WithReason(errs.ReasonDuplicate)
		}

		updates["name"] = params.Name
	}

	if params.ParentIDSet {
		// 检查上级组织不能选择自己或者自己的所有子级
		// 通过向上检查父级链路，确保新的上级及其所有祖先节点中不包含当前组织
		if params.ParentID != nil {
			isCircular, err := uc.wouldCreateCircularReference(ctx, params.ID, params.ParentID)
			if err != nil {
				return err
			}
			if isCircular {
				return errs.NewBusiness("上级组织不能选择自己或自己的子级组织").
					WithReason(errs.ReasonInvalidState)
			}
		}

		updates["parent_id"] = params.ParentID
	}

	if params.TypeSet {
		updates["type"] = params.Type
	}

	if params.OwnerIDSet {
		updates["owner_id"] = params.OwnerID
	}

	if params.RemarkSet {
		updates["remark"] = params.Remark
	}

	if len(updates) > 0 {
		updates["updated_by"] = params.UpdatedBy
		err = uc.tx.Transaction(ctx, func(ctx context.Context) error {
			if err := uc.organizationRepo.Update(ctx, organization, updates); err != nil {
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

// Delete removes a organization.
func (uc *OrganizationUseCase) Delete(ctx context.Context, id uint) error {
	// Check if organization has sub organizations
	subOrgExists, err := uc.organizationRepo.IsExists(ctx, map[string]any{"parent_id": id}, nil)
	if err != nil {
		return err
	}

	if subOrgExists {
		return errs.NewBusiness("组织下存在下级，无法删除").
			WithReason(errs.ReasonInvalidState)
	}

	// Check if organization has positions
	positionExists, err := uc.positionRepo.IsExists(ctx, map[string]any{"organization_id": id}, nil)
	if err != nil {
		return err
	}

	if positionExists {
		return errs.NewBusiness("组织下存在岗位，无法删除").
			WithReason(errs.ReasonInvalidState)
	}

	return uc.tx.Transaction(ctx, func(ctx context.Context) error {
		if err := uc.organizationRepo.Delete(ctx, id); err != nil {
			return err
		}
		return nil
	})
}

// wouldCreateCircularReference 检查将 newParentID 设置为 orgID 的上级是否会造成循环引用
// 通过向上遍历 newParentID 的所有祖先节点，检查是否包含 orgID
// 这种方式比向下遍历子树更高效，因为组织树的高度通常远小于子孙节点数量
func (uc *OrganizationUseCase) wouldCreateCircularReference(ctx context.Context, recordID uint, newParentID *uint) (bool, error) {
	if newParentID == nil {
		return false, nil
	}

	// 直接检查：不能选择自己作为上级
	if *newParentID == recordID {
		return true, nil
	}

	// 向上遍历 newParentID 的所有祖先节点
	checkID := newParentID
	for checkID != nil {
		// 获取当前节点的信息
		checkRecord, err := uc.organizationRepo.FindByID(ctx, *checkID)
		if err != nil {
			return false, err
		}

		// 检查父节点是否是目标组织（会形成循环）
		if checkRecord.ParentID != nil && *checkRecord.ParentID == recordID {
			return true, nil
		}

		// 继续向上查找
		checkID = checkRecord.ParentID
	}

	return false, nil
}

// GetAllOrganizations 获取所有组织
func (uc *OrganizationUseCase) GetAllOrganizations(ctx context.Context) ([]*model.Organization, error) {
	organizations, _, err := uc.organizationRepo.List(ctx, map[string]any{}, -1, -1, "")
	if err != nil {
		return nil, err
	}

	return organizations, nil
}

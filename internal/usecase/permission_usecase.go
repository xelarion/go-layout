package usecase

import (
	"context"

	"github.com/xelarion/go-layout/internal/permission"
)

// PermissionUseCase represents the permission use case
type PermissionUseCase struct {
	permBuilder *permission.Builder
}

// NewPermissionUseCase creates a new instance of PermissionUseCase
func NewPermissionUseCase() *PermissionUseCase {
	return &PermissionUseCase{
		permBuilder: permission.NewBuilder(),
	}
}

// GetPermissionTree returns the permission tree
func (uc *PermissionUseCase) GetPermissionTree(ctx context.Context) ([]*permission.Node, error) {
	// Get permission tree from the builder
	return uc.permBuilder.GetTree(ctx)
}

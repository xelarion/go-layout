// Package service provides HTTP service implementations.
package service

import "github.com/google/wire"

// ProviderSet provides all service layer dependencies.
var ProviderSet = wire.NewSet(
	// Services sorted by name
	NewAuthService,
	NewDepartmentService,
	NewPermissionService,
	NewRoleService,
	NewUserService,
)

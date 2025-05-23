// Package usecase provides business logic implementations.
package usecase

import "github.com/google/wire"

// ProviderSet provides all usecase layer dependencies.
var ProviderSet = wire.NewSet(
	// Repositories sorted by name
	NewDepartmentUseCase,
	NewPermissionUseCase,
	NewRoleUseCase,
	NewUserUseCase,
)

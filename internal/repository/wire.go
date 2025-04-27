// Package repository provides data access implementations.
package repository

import "github.com/google/wire"

// ProviderSet provides all repository layer dependencies.
var ProviderSet = wire.NewSet(
	NewData,
	NewTransaction,

	// Repositories sorted by name
	NewDepartmentRepository,
	NewRoleRepository,
	NewUserRepository,
)

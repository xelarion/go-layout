// Package repository provides data access implementations.
package repository

import "github.com/google/wire"

// ProviderSet provides all repository layer dependencies.
var ProviderSet = wire.NewSet(
	NewData,
	NewTransaction,
	NewUserRepository,
	NewDepartmentRepository,
	NewRoleRepository,
)

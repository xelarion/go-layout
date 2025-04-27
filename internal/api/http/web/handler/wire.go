// Package handler provides HTTP handlers for the web API.
package handler

import "github.com/google/wire"

// ProviderSet provides all handler dependencies.
var ProviderSet = wire.NewSet(
	// Handlers sorted by name
	NewAuthHandler,
	NewDepartmentHandler,
	NewPermissionHandler,
	NewRoleHandler,
	NewUserHandler,
)

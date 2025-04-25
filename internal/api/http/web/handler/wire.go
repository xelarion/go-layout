// Package handler provides HTTP handlers for the web API.
package handler

import "github.com/google/wire"

// ProviderSet provides all handler dependencies.
var ProviderSet = wire.NewSet(
	NewAuthHandler,
	NewUserHandler,
	NewDepartmentHandler,
	NewRoleHandler,
	NewPermissionHandler,
)

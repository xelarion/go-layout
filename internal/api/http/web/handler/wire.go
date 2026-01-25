// Package handler provides HTTP handlers for the web API.
package handler

import "github.com/google/wire"

// ProviderSet provides all handler dependencies.
var ProviderSet = wire.NewSet(
	// Handlers sorted by name
	NewAnnouncementHandler,
	NewAuthHandler,
	NewBatchDetailHandler,
	NewBatchRecordHandler,
	NewChangeLogHandler,
	NewCodeHandler,
	NewCodeRuleHandler,
	NewExportHandler,
	NewImportHandler,
	NewNotificationHandler,
	NewOrganizationHandler,
	NewPermissionHandler,
	NewPositionHandler,
	NewRoleHandler,
	NewStorageHandler,
	NewUserHandler,
	NewWSHandler,
)

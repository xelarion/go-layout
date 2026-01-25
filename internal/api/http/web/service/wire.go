// Package service provides HTTP service implementations.
package service

import "github.com/google/wire"

// ProviderSet provides all service layer dependencies.
var ProviderSet = wire.NewSet(
	// Base Service (provides data scope functionality)
	NewBaseService,

	// Services sorted by name
	NewAnnouncementService,
	NewAuthService,
	NewBatchDetailService,
	NewBatchRecordService,
	NewChangeLogService,
	NewCodeRuleService,
	NewCodeService,
	NewExportService,
	NewImportService,
	NewNotificationService,
	NewOrganizationService,
	NewPermissionService,
	NewPositionService,
	NewRoleService,
	NewStorageService,
	NewUserService,
)

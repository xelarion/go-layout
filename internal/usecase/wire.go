// Package usecase provides business logic implementations.
package usecase

import "github.com/google/wire"

// ProviderSet provides all usecase layer dependencies.
var ProviderSet = wire.NewSet(
	// UseCases sorted by name
	NewAmqpMessageUseCase,
	NewAnnouncementUseCase,
	NewBatchDetailUseCase,
	NewBatchRecordUseCase,
	NewChangeLogDetailUseCase,
	NewChangeLogUseCase,
	NewCodeGeneratorUseCase,
	NewCodeRuleUseCase,
	NewExportUseCase,
	NewImportUseCase,
	NewNotificationUseCase,
	NewOrganizationUseCase,
	NewPermissionUseCase,
	NewPositionUseCase,
	NewRoleUseCase,
	NewStorageAttachmentUseCase,
	NewStorageBlobUseCase,
	NewStorageUseCase,
	NewUserUseCase,
)

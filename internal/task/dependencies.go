// Package task provides common functionality for all task types.
package task

import (
	"go.uber.org/zap"

	"github.com/xelarion/go-layout/internal/infra/wsocket"
	"github.com/xelarion/go-layout/internal/repository"
	"github.com/xelarion/go-layout/internal/usecase"
)

// Dependencies contains all dependencies needed by tasks.
type Dependencies struct {
	// Core infrastructure (unexported)
	data   *repository.Data
	logger *zap.Logger
	TX     usecase.Transaction

	// Broadcaster for WebSocket push
	Broadcaster wsocket.Broadcaster

	// @insert:dependencies_repository_define 请对以下内容进行排序
	AmqpMessageRepo      usecase.AmqpMessageRepository
	ExportTaskRepo       usecase.ExportTaskRepository
	MQRepo               usecase.MQRepository
	NotificationRepo     usecase.NotificationRepository
	StorageBlobRepo      usecase.StorageBlobRepository
	UserNotificationRepo usecase.UserNotificationRepository
	UserRepo             usecase.UserRepository

	// @insert:dependencies_usecase_define 请对以下内容进行排序
	BatchDetailUseCase  *usecase.BatchDetailUseCase
	BatchRecordUseCase  *usecase.BatchRecordUseCase
	ExportUseCase       *usecase.ExportUseCase
	NotificationUseCase *usecase.NotificationUseCase
	StorageUseCase      *usecase.StorageUseCase
	UserUseCase         *usecase.UserUseCase
}

// NewDependencies creates a new Dependencies instance with all required dependencies.
func NewDependencies(
	data *repository.Data,
	logger *zap.Logger,
	tx usecase.Transaction,
	broadcaster wsocket.Broadcaster,

	// @insert:new_dependencies_repository_define 请对以下内容进行排序
	amqpMessageRepo usecase.AmqpMessageRepository,
	exportTaskRepo usecase.ExportTaskRepository,
	mqRepo usecase.MQRepository,
	notificationRepo usecase.NotificationRepository,
	storageBlobRepo usecase.StorageBlobRepository,
	userNotificationRepo usecase.UserNotificationRepository,
	userRepo usecase.UserRepository,

	// @insert:new_dependencies_usecase_define 请对以下内容进行排序
	batchDetailUseCase *usecase.BatchDetailUseCase,
	batchRecordUseCase *usecase.BatchRecordUseCase,
	exportUseCase *usecase.ExportUseCase,
	notificationUseCase *usecase.NotificationUseCase,
	storageUseCase *usecase.StorageUseCase,
	userUseCase *usecase.UserUseCase,
) *Dependencies {
	return &Dependencies{
		data:        data,
		logger:      logger,
		TX:          tx,
		Broadcaster: broadcaster,

		// @insert:new_dependencies_repository_return 请对以下内容进行排序
		AmqpMessageRepo:      amqpMessageRepo,
		ExportTaskRepo:       exportTaskRepo,
		MQRepo:               mqRepo,
		NotificationRepo:     notificationRepo,
		StorageBlobRepo:      storageBlobRepo,
		UserNotificationRepo: userNotificationRepo,
		UserRepo:             userRepo,

		// @insert:new_dependencies_usecase_return 请对以下内容进行排序
		BatchDetailUseCase:  batchDetailUseCase,
		BatchRecordUseCase:  batchRecordUseCase,
		ExportUseCase:       exportUseCase,
		NotificationUseCase: notificationUseCase,
		StorageUseCase:      storageUseCase,
		UserUseCase:         userUseCase,
	}
}

// Data returns the data access layer
func (d *Dependencies) Data() *repository.Data {
	return d.data
}

// Logger returns the logger
func (d *Dependencies) Logger() *zap.Logger {
	return d.logger
}

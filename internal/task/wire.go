// Package task provides common functionality for all task types.
package task

import (
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"

	"github.com/xelarion/go-layout/internal/repository"
	"github.com/xelarion/go-layout/internal/usecase"
)

// ProviderSet provides all task dependencies.
var ProviderSet = wire.NewSet(
	repository.NewData,
	repository.NewTransaction,
	ProvideRedis,

	// @insert:repository_provider 请对以下内容进行排序
	repository.NewAmqpMessageRepository,
	repository.NewBatchDetailRepository,
	repository.NewBatchRecordRepository,
	repository.NewCacheRepository,
	repository.NewExportTaskRepository,
	repository.NewMQRepository,
	repository.NewNotificationRepository,
	repository.NewOrganizationRepository,
	repository.NewPositionRepository,
	repository.NewRoleRepository,
	repository.NewStorageAttachmentRepository,
	repository.NewStorageBlobRepository,
	repository.NewUserNotificationRepository,
	repository.NewUserRepository,

	// @insert:usecase_provider 请对以下内容进行排序
	usecase.NewAmqpMessageUseCase,
	usecase.NewBatchDetailUseCase,
	usecase.NewBatchRecordUseCase,
	usecase.NewExportUseCase,
	usecase.NewNotificationUseCase,
	usecase.NewStorageUseCase,
	usecase.NewUserUseCase,

	// Dependencies constructor
	NewDependencies,
)

// ProvideRedis provides Redis client from Data.
func ProvideRedis(data *repository.Data) *redis.Client {
	return data.Redis()
}

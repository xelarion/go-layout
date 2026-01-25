package usecase

import (
	"context"

	"gorm.io/datatypes"

	"github.com/xelarion/go-layout/internal/infra/mq"
	"github.com/xelarion/go-layout/internal/model"
	"github.com/xelarion/go-layout/internal/model/gen"
	"github.com/xelarion/go-layout/pkg/errs"
)

// AmqpMessageRepository defines methods for amqp_message data access.
type AmqpMessageRepository interface {
	Create(ctx context.Context, amqpMessage *model.AmqpMessage) error
	List(ctx context.Context, filters map[string]any, limit, offset int, sortClause string) ([]*model.AmqpMessage, int, error)
	IsExists(ctx context.Context, filters map[string]any, notFilters map[string]any) (bool, error)
	Count(ctx context.Context, filters map[string]any, notFilters map[string]any) (int64, error)
	FindByID(ctx context.Context, id uint) (*model.AmqpMessage, error)
	Update(ctx context.Context, amqpMessage *model.AmqpMessage, params map[string]any) error
	Delete(ctx context.Context, id uint) error
}

// AmqpMessageUseCase implements business logic for amqp_message operations.
type AmqpMessageUseCase struct {
	tx              Transaction
	amqpMessageRepo AmqpMessageRepository
	mqRepository    MQRepository
	exchange        string
}

// NewAmqpMessageUseCase creates a new instance of AmqpMessageUseCase.
func NewAmqpMessageUseCase(
	tx Transaction,
	amqpMessageRepo AmqpMessageRepository,
	mqRepository MQRepository,
	exchange string,
) *AmqpMessageUseCase {
	return &AmqpMessageUseCase{
		tx:              tx,
		amqpMessageRepo: amqpMessageRepo,
		mqRepository:    mqRepository,
		exchange:        exchange,
	}
}

// Publish 发布消息到消息队列
// 如果在事务中,则存储到数据库等待异步发送;否则直接发送到 MQ
func (uc *AmqpMessageUseCase) Publish(ctx context.Context, data any, routingKeys []string, options ...func(*mq.PublishOptions)) error {
	// 序列化数据
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return errs.WrapInternal(err, "failed to marshal message data")
	}

	// 检测是否在事务中
	if uc.tx.IsInTransaction(ctx) {
		// 在事务中,存储到数据库,稍后由异步任务发送
		// 应用选项并保存为结构体
		publishOpts := &mq.PublishOptions{}
		for _, option := range options {
			option(publishOpts)
		}

		message := &model.AmqpMessage{
			AmqpMessage: gen.AmqpMessage{
				Exchange:    uc.exchange,
				RoutingKeys: routingKeys,
				Data:        dataBytes,
			},
			Options: datatypes.NewJSONType(*publishOpts),
		}

		return uc.amqpMessageRepo.Create(ctx, message)
	}

	// 不在事务中,直接发送到 MQ（使用函数选项）
	return uc.mqRepository.Publish(ctx, uc.exchange, routingKeys, dataBytes, options...)
}

func (uc *AmqpMessageUseCase) PublishWebNotifier(ctx context.Context, data map[string]interface{}) error {
	return uc.Publish(ctx, data, []string{"notifier-queue"})
}

func (uc *AmqpMessageUseCase) PublishDataExportTask(ctx context.Context, exportTaskID uint) error {
	return uc.Publish(ctx, map[string]any{"export_task_id": exportTaskID}, []string{"export-queue"})
}

func (uc *AmqpMessageUseCase) PublishImportCustomers(ctx context.Context, batchRecordID uint) error {
	return uc.Publish(ctx, map[string]any{"batch_record_id": batchRecordID}, []string{"import-customer-queue"})
}
func (uc *AmqpMessageUseCase) PublishImportSuppliers(ctx context.Context, batchRecordID uint) error {
	return uc.Publish(ctx, map[string]any{"batch_record_id": batchRecordID}, []string{"import-supplier-queue"})
}

func (uc *AmqpMessageUseCase) PublishImportEquipments(ctx context.Context, batchRecordID uint) error {
	return uc.Publish(ctx, map[string]any{"batch_record_id": batchRecordID}, []string{"import-equipment-queue"})
}

func (uc *AmqpMessageUseCase) PublishImportEquipmentTypes(ctx context.Context, batchRecordID uint) error {
	return uc.Publish(ctx, map[string]any{"batch_record_id": batchRecordID}, []string{"import-equipment-type-queue"})
}

// PublishCreateUserNotification 发布创建用户通知任务
func (uc *AmqpMessageUseCase) PublishCreateUserNotification(ctx context.Context, notificationID uint) error {
	return uc.Publish(ctx, map[string]any{"notification_id": notificationID}, []string{"create-user-notification-queue"})
}

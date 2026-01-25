package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/xelarion/go-layout/internal/model"
	"github.com/xelarion/go-layout/internal/usecase"
	"github.com/xelarion/go-layout/pkg/errs"
)

var _ usecase.AmqpMessageRepository = (*AmqpMessageRepository)(nil)

// AmqpMessageRepository is an implementation of the queue repository.
type AmqpMessageRepository struct {
	data *Data
}

// NewAmqpMessageRepository creates a new instance of queue repository.
func NewAmqpMessageRepository(data *Data) usecase.AmqpMessageRepository {
	return &AmqpMessageRepository{
		data: data,
	}
}

// Create adds a new amqp_message to the database.
func (r *AmqpMessageRepository) Create(ctx context.Context, amqpMessage *model.AmqpMessage) error {
	if err := r.data.DB(ctx).Create(amqpMessage).Error; err != nil {
		return errs.WrapInternal(err, "failed to create amqp_message")
	}
	return nil
}

// List retrieves amqp_messages with pagination and filtering.
func (r *AmqpMessageRepository) List(ctx context.Context, filters map[string]any, limit, offset int, sortClause string) ([]*model.AmqpMessage, int, error) {
	query := r.data.DB(ctx).Model(&model.AmqpMessage{})

	for field, value := range filters {
		if value == nil {
			continue
		}

		switch field {
		default:
			query = query.Where("amqp_messages."+field+" = ?", value)
		}
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errs.WrapInternal(err, "failed to count amqp_messages")
	}

	if sortClause != "" {
		query = query.Order(sortClause)
	} else {
		query = query.Order("amqp_messages.id desc")
	}

	var amqpMessages []*model.AmqpMessage
	if err := query.Limit(limit).Offset(offset).Find(&amqpMessages).Error; err != nil {
		return nil, 0, errs.WrapInternal(err, "failed to list amqp_messages")
	}

	return amqpMessages, int(total), nil
}

// IsExists checks if a amqp_message exists in the database.
func (r *AmqpMessageRepository) IsExists(ctx context.Context, filters map[string]any, notFilters map[string]any) (bool, error) {
	return IsExists(ctx, r.data.DB(ctx), &model.AmqpMessage{}, filters, notFilters)
}

// Count counts the number of amqp_messages.
func (r *AmqpMessageRepository) Count(ctx context.Context, filters map[string]any, notFilters map[string]any) (int64, error) {
	return Count(ctx, r.data.DB(ctx), &model.AmqpMessage{}, filters, notFilters)
}

// FindByID retrieves a amqp_message by ID.
func (r *AmqpMessageRepository) FindByID(ctx context.Context, id uint) (*model.AmqpMessage, error) {
	var amqpMessage model.AmqpMessage
	err := r.data.DB(ctx).First(&amqpMessage, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewBusiness("amqp_message not found").
				WithReason(errs.ReasonNotFound).
				WithMeta("id", id)
		}
		return nil, errs.WrapInternal(err, "failed to find amqp_message by ID")
	}
	return &amqpMessage, nil
}

// Update updates a amqp_message.
func (r *AmqpMessageRepository) Update(ctx context.Context, amqpMessage *model.AmqpMessage, params map[string]any) error {
	result := r.data.DB(ctx).Model(amqpMessage).Updates(params)
	if result.Error != nil {
		return errs.WrapInternal(result.Error, "failed to update amqp_message")
	}

	if result.RowsAffected == 0 {
		return errs.NewBusiness("amqp_message not found").
			WithReason(errs.ReasonNotFound).
			WithMeta("id", amqpMessage.ID)
	}

	return nil
}

// Delete removes a amqp_message by ID.
func (r *AmqpMessageRepository) Delete(ctx context.Context, id uint) error {
	result := r.data.DB(ctx).Delete(&model.AmqpMessage{}, id)
	if result.Error != nil {
		return errs.WrapInternal(result.Error, "failed to delete amqp_message")
	}

	if result.RowsAffected == 0 {
		return errs.NewBusiness("amqp_message not found").
			WithReason(errs.ReasonNotFound).
			WithMeta("id", id)
	}

	return nil
}

package model

import (
	"gorm.io/datatypes"

	"github.com/xelarion/go-layout/internal/infra/mq"
	"github.com/xelarion/go-layout/internal/model/gen"
)

// AmqpMessage mapped from table <amqp_messages>
type AmqpMessage struct {
	gen.AmqpMessage

	Options datatypes.JSONType[mq.PublishOptions] `gorm:"column:options;type:jsonb;not null;default:{};comment:选项" json:"options"` // 选项
}

func (m *AmqpMessage) GetResourceType() string {
	return "amqp_message"
}

func (m *AmqpMessage) GetResourceID() uint {
	return m.ID
}

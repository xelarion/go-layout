package model

import (
	"gorm.io/datatypes"

	"github.com/xelarion/go-layout/internal/enum"
	"github.com/xelarion/go-layout/internal/model/gen"
)

type ChangeLogDetail struct {
	gen.ChangeLogDetail

	FieldType    enum.ChangeLogDetailFieldType                   `gorm:"column:field_type;type:smallint;not null;comment:字段类型" json:"field_type"`       // 字段类型
	RelationType enum.ChangeLogDetailRelationType                `gorm:"column:relation_type;type:smallint;not null;comment:关联类型" json:"relation_type"` // 关联类型
	RelationMeta datatypes.JSONType[ChangeLogDetailRelationMeta] `gorm:"column:relation_meta;type:jsonb;comment:关联元数据" json:"relation_meta,omitempty"`  // 关联元数据(仅一对多关联使用)
}

// ChangeLogDetailRelationMeta 一对多关联的元数据
type ChangeLogDetailRelationMeta struct {
	Schema  []ChangeLogDetailSchemaField `json:"schema,omitempty"`   // 列定义
	GroupBy string                       `json:"group_by,omitempty"` // 分组字段名
}

// ChangeLogDetailSchemaField 关联字段的列定义
type ChangeLogDetailSchemaField struct {
	Name  string `json:"name"`  // 字段名
	Label string `json:"label"` // 字段显示名称
}

// GetResourceType 实现 Resource 接口
func (e *ChangeLogDetail) GetResourceType() string {
	return "change_log_detail"
}

// GetResourceID 实现 Resource 接口
func (e *ChangeLogDetail) GetResourceID() uint {
	return e.ID
}

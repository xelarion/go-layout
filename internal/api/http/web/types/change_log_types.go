package types

// Common ChangeLog related types

// ListChangeLogsReq represents change_log list query parameters.
type ListChangeLogsReq struct {
	SortReq
	ResourceType string `form:"resource_type" binding:"required,max=255"` // 资源类型
	ResourceID   uint   `form:"resource_id" binding:"required"`           // 资源ID
}

// ListChangeLogsResp represents change_log list response.
type ListChangeLogsResp struct {
	Results []*ListChangeLogsRespResult `json:"results"`
}

// ListChangeLogsRespResult represents a single change_log in the list.
type ListChangeLogsRespResult struct {
	ID              uint                              `json:"id"`               // ID
	CreatedAt       string                            `json:"created_at"`       // 创建时间
	OperatorDisplay string                            `json:"operator_display"` // 操作人ID（关联显示）
	Action          string                            `json:"action"`           // 操作类型（显示文本）
	ActionSlug      string                            `json:"action_slug"`      // 操作类型（slug）
	Details         []*ListChangeLogsRespResultDetail `json:"details"`          // 变更细节
}

// ListChangeLogsRespResultDetail represents a change log detail.
type ListChangeLogsRespResultDetail struct {
	ID              uint                             `json:"id"`                          // ID
	FieldName       string                           `json:"field_name"`                  // 字段名(结构体字段名)
	FieldLabel      string                           `json:"field_label"`                 // 字段显示名称
	OldValue        string                           `json:"old_value"`                   // 旧值
	NewValue        string                           `json:"new_value"`                   // 新值
	OldRelations    []map[string]string              `json:"old_relations,omitempty"`     // 旧关联
	NewRelations    []map[string]string              `json:"new_relations,omitempty"`     // 新关联
	RelationSchema  []ChangeLogDetailSchemaFieldResp `json:"relation_schema,omitempty"`   // 关联字段的列定义
	RelationGroupBy string                           `json:"relation_group_by,omitempty"` // 关联字段的分组字段名
}

// ChangeLogDetailSchemaFieldResp 关联字段的列定义
type ChangeLogDetailSchemaFieldResp struct {
	Name  string `json:"name"`  // 字段名
	Label string `json:"label"` // 字段显示名称
}

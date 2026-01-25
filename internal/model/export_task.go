package model

import (
	"github.com/xelarion/go-layout/internal/enum"
	"github.com/xelarion/go-layout/internal/model/gen"
)

// ExportTask 导出任务模型
type ExportTask struct {
	gen.ExportTask

	TaskType enum.ExportTaskType   `gorm:"column:task_type;type:character varying(50);not null;comment:任务类型: user, role, position等" json:"task_type"`           // 任务类型: user, role, position等
	Status   enum.ExportTaskStatus `gorm:"column:status;type:smallint;not null;index:idx_export_tasks_created_by_status,priority:2;comment:任务状态" json:"status"` // 任务状态
}

// GetResourceType 实现 Resource 接口
func (m *ExportTask) GetResourceType() string {
	return "export_task"
}

// GetResourceID 实现 Resource 接口
func (m *ExportTask) GetResourceID() uint {
	return m.ID
}

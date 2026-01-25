package tasks

import (
	"context"

	"go.uber.org/zap"

	"github.com/xelarion/go-layout/internal/task"
	"github.com/xelarion/go-layout/internal/task/scheduler"
)

// ExportCleanerHandler 导出文件清理任务
type ExportCleanerHandler struct {
	deps   *task.Dependencies
	logger *zap.Logger
}

// NewExportCleanerHandler 创建导出清理任务实例
func NewExportCleanerHandler(deps *task.Dependencies, logger *zap.Logger) *ExportCleanerHandler {
	return &ExportCleanerHandler{
		deps:   deps,
		logger: logger.Named("export-cleaner-handler"),
	}
}

// Register 注册定时任务
func (t *ExportCleanerHandler) Register(s *scheduler.Scheduler) error {
	// 每天凌晨 2 点执行清理任务
	return s.Register("export-cleaner", "0 0 2 * * *", t.Execute)
}

// Execute 执行清理任务
func (t *ExportCleanerHandler) Execute(ctx context.Context) error {
	err := t.deps.ExportUseCase.CleanExpiredTasks(ctx)
	if err != nil {
		return err
	}
	return nil
}

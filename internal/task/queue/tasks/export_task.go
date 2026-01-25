package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"go.uber.org/zap"

	"github.com/xelarion/go-layout/internal/enum"
	"github.com/xelarion/go-layout/internal/model"
	"github.com/xelarion/go-layout/internal/task"
	"github.com/xelarion/go-layout/internal/task/queue"
	"github.com/xelarion/go-layout/pkg/errs"
)

// ExportTaskPayload 导出任务载荷
type ExportTaskPayload struct {
	ExportTaskID uint `json:"export_task_id"` // 任务ID
}

// ExportTaskHandler 导出任务处理器
type ExportTaskHandler struct {
	deps   *task.Dependencies
	logger *zap.Logger
}

// NewExportTaskHandler 创建导出任务处理器
func NewExportTaskHandler(deps *task.Dependencies, logger *zap.Logger) *ExportTaskHandler {
	return &ExportTaskHandler{
		deps:   deps,
		logger: logger.Named("export-task-handler"),
	}
}

// Register 注册任务消费者
func (h *ExportTaskHandler) Register(q *queue.Manager) error {
	// 注册用户导出任务
	if err := q.RegisterConsumer(
		"export-handler",
		"export-queue",
		h.Execute,
		queue.WithConsumerOptionsConcurrency(8),
		//queue.WithConsumerOptionsDurable(true),
		//queue.WithConsumerOptionsRoutingKey("example.task"),
	); err != nil {
		return fmt.Errorf("failed to register export user task handler: %w", err)
	}

	return nil
}

// Execute 执行导出任务
func (h *ExportTaskHandler) Execute(ctx context.Context, msg queue.Message) (queue.Action, error) {
	var payload ExportTaskPayload
	if err := json.Unmarshal(msg.Body, &payload); err != nil {
		return queue.NackDiscard, errs.WrapInternal(err, "failed to unmarshal payload")
	}

	exportTask, err := h.deps.ExportTaskRepo.FindByID(ctx, payload.ExportTaskID)
	if err != nil {
		if errs.IsReason(err, errs.ReasonNotFound) {
			// 导出任务可能被删除
			return queue.NackDiscard, nil
		}
		return queue.NackRequeue, err
	}

	// 开始
	if err := h.deps.ExportUseCase.Start(ctx, payload.ExportTaskID); err != nil {
		h.logger.Error("failed to start export task", zap.Error(err))
	}

	// 根据任务类型执行不同的导出逻辑
	switch exportTask.TaskType {
	case enum.ExportTaskTypeUser:
		err = h.executeByExportFunc(ctx, exportTask, h.deps.UserUseCase.ExportUsers)
	default:
		err = fmt.Errorf("unsupported export exportTask type: %s", exportTask.TaskType)
	}

	if err != nil {
		// 导出失败
		errMsg := err.Error()
		if updateErr := h.deps.ExportUseCase.Fail(ctx, payload.ExportTaskID, errMsg); updateErr != nil {
			h.logger.Error("failed to mark export task as failed", zap.Error(updateErr))
		}
		return queue.Ack, err
	}

	return queue.Ack, nil
}

func (h *ExportTaskHandler) executeByExportFunc(
	ctx context.Context,
	task *model.ExportTask,
	exportFunc func(ctx context.Context, filters map[string]any, filePath string) error,
) error {
	isAttachmentExist := false
	_, err := h.deps.ExportUseCase.GetExportFileAttachment(ctx, task)
	if err != nil {
		if !errs.IsReason(err, errs.ReasonNotFound) {
			return err
		}
	} else {
		isAttachmentExist = true
	}

	// 如果文件不存在，执行导出，防止程序重启后重复导出
	if !isAttachmentExist {
		var filters map[string]any
		if err := json.Unmarshal(task.Params, &filters); err != nil {
			return errs.WrapInternal(err, "failed to unmarshal export params")
		}

		// 1. 生成临时本地文件
		tmpFile, err := os.CreateTemp("", "export-*.xlsx")
		if err != nil {
			return errs.WrapInternal(err, "创建临时文件失败")
		}
		tmpPath := tmpFile.Name()
		tmpFile.Close()
		defer os.Remove(tmpPath) // 确保清理临时文件

		// 2. 调用导出函数写入数据
		if err := exportFunc(ctx, filters, tmpPath); err != nil {
			return err
		}

		// 3. 获取文件信息
		fileInfo, err := os.Stat(tmpPath)
		if err != nil {
			return errs.WrapInternal(err, "获取临时文件信息失败")
		}

		// 4. 打开文件准备上传
		file, err := os.Open(tmpPath)
		if err != nil {
			return errs.WrapInternal(err, "打开临时文件失败")
		}
		defer file.Close()

		// 5. 上传文件并关联到任务（直接使用 task.Filename）
		if err := h.deps.ExportUseCase.AttachExportFile(
			ctx,
			task,
			file,
			task.Filename,
			fileInfo.Size(),
		); err != nil {
			return err
		}
	}

	// 6. 标记任务完成
	return h.deps.ExportUseCase.Complete(ctx, task.ID)
}

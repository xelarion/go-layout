package usecase

import (
	"context"
	"io"
	"os"
	"time"

	"go.uber.org/zap"

	"github.com/xelarion/go-layout/internal/enum"
	"github.com/xelarion/go-layout/internal/infra/config"
	"github.com/xelarion/go-layout/internal/infra/mq"
	"github.com/xelarion/go-layout/internal/model"
	"github.com/xelarion/go-layout/internal/model/gen"
	"github.com/xelarion/go-layout/pkg/errs"
)

// ExportTask 导出任务
type ExportTask struct {
	*model.ExportTask
	FileSize int64
}

// ExportResult 导出结果
type ExportResult struct {
	TaskID      uint   // 任务ID
	TotalCount  int    // 总记录数
	IsAsync     bool   // 是否异步导出
	DownloadURL string // 下载URL（仅同步导出时有值）
}

// ExportTaskRepository 导出任务数据访问接口
type ExportTaskRepository interface {
	Create(ctx context.Context, exportTask *model.ExportTask) error
	List(ctx context.Context, filters map[string]any, limit, offset int, sortClause string) ([]*model.ExportTask, int, error)
	IsExists(ctx context.Context, filters map[string]any, notFilters map[string]any) (bool, error)
	Count(ctx context.Context, filters map[string]any, notFilters map[string]any) (int64, error)
	FindByID(ctx context.Context, id uint) (*model.ExportTask, error)
	FindByIDForUpdate(ctx context.Context, id uint) (*model.ExportTask, error)
	Update(ctx context.Context, exportTask *model.ExportTask, params map[string]any) error
	Delete(ctx context.Context, id uint) error
}

// ExportUseCase 导出业务逻辑
type ExportUseCase struct {
	tx                 Transaction
	config             *config.Export
	rmq                *mq.RabbitMQ
	exportTaskRepo     ExportTaskRepository
	storageBlobRepo    StorageBlobRepository
	storageUC          *StorageUseCase
	amqpMessageUseCase *AmqpMessageUseCase
	logger             *zap.Logger
}

// NewExportUseCase 创建导出用例实例
func NewExportUseCase(
	tx Transaction,
	config *config.Export,
	rmq *mq.RabbitMQ,
	exportTaskRepo ExportTaskRepository,
	storageBlobRepo StorageBlobRepository,
	storageUC *StorageUseCase,
	amqpMessageUseCase *AmqpMessageUseCase,
	logger *zap.Logger,
) *ExportUseCase {
	return &ExportUseCase{
		tx:                 tx,
		config:             config,
		rmq:                rmq,
		exportTaskRepo:     exportTaskRepo,
		storageBlobRepo:    storageBlobRepo,
		storageUC:          storageUC,
		amqpMessageUseCase: amqpMessageUseCase,
		logger:             logger.Named("export-usecase"),
	}
}

// ShouldExportAsync 判断是否应该异步导出
func (uc *ExportUseCase) ShouldExportAsync(totalCount int) bool {
	return totalCount > uc.config.SyncThreshold
}

// Execute 执行导出（自动判断同步/异步）
func (uc *ExportUseCase) Execute(
	ctx context.Context,
	taskType enum.ExportTaskType,
	totalCount int,
	params map[string]any,
	userID uint,
	filename string,
	exportFunc func(ctx context.Context, filters map[string]any, filePath string) error,
) (*ExportResult, error) {
	// 判断是否需要异步导出
	if uc.ShouldExportAsync(totalCount) {
		// 异步导出：创建任务并发送到队列
		task, err := uc.CreateExportTask(ctx, taskType, totalCount, params, userID, filename)
		if err != nil {
			return nil, err
		}

		return &ExportResult{
			IsAsync:    true,
			TaskID:     task.ID,
			TotalCount: totalCount,
		}, nil
	}

	// 同步导出：立即执行
	return uc.ExecuteSyncExport(ctx, taskType, totalCount, params, userID, filename, exportFunc)
}

// CreateExportTask 创建导出任务
func (uc *ExportUseCase) CreateExportTask(
	ctx context.Context,
	taskType enum.ExportTaskType,
	totalCount int,
	params map[string]any,
	userID uint,
	filename string,
) (*model.ExportTask, error) {
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return nil, errs.WrapInternal(err, "failed to marshal export params")
	}

	task := &model.ExportTask{
		ExportTask: gen.ExportTask{
			CreatedBy:    userID,
			Filename:     filename,
			IsAsync:      true,
			TotalCount:   int32(totalCount),
			Params:       paramsJSON,
			ErrorMessage: "",
			StartedAt:    nil,
			CompletedAt:  nil,
		},
		TaskType: taskType,
		Status:   enum.ExportTaskStatusPending,
	}

	err = uc.tx.Transaction(ctx, func(ctx context.Context) error {
		if err := uc.exportTaskRepo.Create(ctx, task); err != nil {
			return err
		}

		if err := uc.amqpMessageUseCase.PublishDataExportTask(ctx, task.ID); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return task, nil
}

// ExecuteSyncExport 执行同步导出（封装公共逻辑）
func (uc *ExportUseCase) ExecuteSyncExport(
	ctx context.Context,
	taskType enum.ExportTaskType,
	totalCount int,
	params map[string]any,
	userID uint,
	filename string,
	exportFunc func(ctx context.Context, filters map[string]any, filePath string) error,
) (*ExportResult, error) {
	// 1. 创建临时文件
	tmpFile, err := os.CreateTemp("", "export-*.xlsx")
	if err != nil {
		return nil, errs.WrapInternal(err, "创建临时文件失败")
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath) // 确保清理临时文件

	// 2. 导出数据到临时文件
	if err := exportFunc(ctx, params, tmpPath); err != nil {
		return nil, err
	}

	// 3. 创建已完成的任务记录
	task, err := uc.CreateCompletedExportTask(
		ctx,
		taskType,
		totalCount,
		params,
		userID,
		filename,
	)
	if err != nil {
		return nil, err
	}

	// 4. 获取文件信息并上传
	fileInfo, err := os.Stat(tmpPath)
	if err != nil {
		return nil, errs.WrapInternal(err, "获取文件信息失败")
	}

	file, err := os.Open(tmpPath)
	if err != nil {
		return nil, errs.WrapInternal(err, "打开临时文件失败")
	}
	defer file.Close()

	// 5. 上传文件并关联到任务
	if err := uc.AttachExportFile(ctx, task, file, filename, fileInfo.Size()); err != nil {
		return nil, err
	}

	// 6. 生成下载URL（同步导出立即返回）
	attachment, err := uc.GetExportFileAttachment(ctx, task)
	if err != nil {
		return nil, err
	}

	downloadURL, err := uc.storageUC.GenerateDownloadURL(ctx, attachment.StorageBlob.Key)
	if err != nil {
		return nil, err
	}

	return &ExportResult{
		TaskID:      task.ID,
		TotalCount:  totalCount,
		IsAsync:     false,
		DownloadURL: downloadURL,
	}, nil
}

// CreateCompletedExportTask 创建已完成的导出任务（用于同步导出）
func (uc *ExportUseCase) CreateCompletedExportTask(
	ctx context.Context,
	taskType enum.ExportTaskType,
	totalCount int,
	params map[string]any,
	userID uint,
	filename string,
) (*model.ExportTask, error) {
	now := time.Now().UTC()

	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return nil, errs.WrapInternal(err, "failed to marshal export params")
	}

	task := &model.ExportTask{
		ExportTask: gen.ExportTask{
			Filename:    filename,
			IsAsync:     false,
			TotalCount:  int32(totalCount),
			Params:      paramsJSON,
			CreatedBy:   userID,
			StartedAt:   &now,
			CompletedAt: &now,
		},
		TaskType: taskType,
		Status:   enum.ExportTaskStatusCompleted,
	}

	if err := uc.exportTaskRepo.Create(ctx, task); err != nil {
		return nil, err
	}

	return task, nil
}

// GetExportTask 获取导出任务
func (uc *ExportUseCase) GetExportTask(ctx context.Context, id uint) (*ExportTask, error) {
	task, err := uc.exportTaskRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	fileSize := int64(0)
	attachment, err := uc.GetExportFileAttachment(ctx, task)
	if err != nil {
		if !errs.IsReason(err, errs.ReasonNotFound) {
			return nil, err
		}
	} else {
		fileSize = attachment.StorageBlob.ByteSize
	}

	return &ExportTask{
		ExportTask: task,
		FileSize:   fileSize,
	}, nil
}

// ListExportTasks 查询导出任务列表
func (uc *ExportUseCase) ListExportTasks(ctx context.Context, filters map[string]any, limit, offset int, sortClause string) ([]*ExportTask, int, error) {
	tasks, count, err := uc.exportTaskRepo.List(ctx, filters, limit, offset, sortClause)
	if err != nil {
		return nil, 0, err
	}

	records := make([]*ExportTask, 0, len(tasks))
	for _, task := range tasks {
		fileSize := int64(0)
		attachment, err := uc.GetExportFileAttachment(ctx, task)
		if err != nil {
			if !errs.IsReason(err, errs.ReasonNotFound) {
				return nil, 0, err
			}
		} else {
			fileSize = attachment.StorageBlob.ByteSize
		}

		records = append(records, &ExportTask{
			ExportTask: task,
			FileSize:   fileSize,
		})
	}

	return records, count, nil
}

// AttachExportFile 上传导出文件并关联到任务
func (uc *ExportUseCase) AttachExportFile(
	ctx context.Context,
	task *model.ExportTask,
	file io.Reader,
	filename string,
	size int64,
) error {
	// 1. 上传文件到对象存储
	uploadResult, err := uc.storageUC.Upload(ctx, UploadParams{
		File:     file,
		Filename: filename,
		Size:     size,
		MimeType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	})
	if err != nil {
		return err
	}

	blob, err := uc.storageBlobRepo.FindByKey(ctx, uploadResult.Key)
	if err != nil {
		return err
	}

	return uc.storageUC.AttachOne(ctx, task, enum.StorageAttachmentNameTypeFile, blob.Key)
}

// GetExportFileAttachment 获取导出任务的文件附件
func (uc *ExportUseCase) GetExportFileAttachment(ctx context.Context, task *model.ExportTask) (*StorageAttachment, error) {
	return uc.storageUC.FindByResourceAndName(ctx, task, enum.StorageAttachmentNameTypeFile)
}

// GetExportFileDownloadURL 获取导出文件下载URL
func (uc *ExportUseCase) GetExportFileDownloadURL(ctx context.Context, id uint, userID uint) (string, error) {
	task, err := uc.exportTaskRepo.FindByID(ctx, id)
	if err != nil {
		return "", err
	}

	// 权限校验
	if task.CreatedBy != userID {
		return "", errs.NewBusiness("无权下载此文件")
	}

	// 状态校验
	if task.Status != enum.ExportTaskStatusCompleted {
		return "", errs.NewBusiness("文件尚未生成")
	}

	// 获取文件附件
	attachment, err := uc.GetExportFileAttachment(ctx, task)
	if err != nil {
		return "", errs.NewBusiness("文件不存在").WithReason(errs.ReasonNotFound)
	}

	// 生成下载URL
	url, err := uc.storageUC.GenerateDownloadURL(ctx, attachment.StorageBlob.Key)
	if err != nil {
		return "", err
	}

	return url, nil
}

// DeleteExportTask 删除导出任务
func (uc *ExportUseCase) DeleteExportTask(ctx context.Context, id uint, userID uint) error {
	task, err := uc.exportTaskRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	// 权限校验
	if task.CreatedBy != userID {
		return errs.NewBusiness("无权删除此任务")
	}

	if task.Status == enum.ExportTaskStatusProcessing {
		return errs.NewBusiness("任务正在处理中，无法删除")
	}

	return uc.tx.Transaction(ctx, func(txCtx context.Context) error {
		// 1. 分离文件附件（blob 由定时任务清理）
		if err := uc.storageUC.Detach(txCtx, task); err != nil {
			uc.logger.Warn("failed to detach export file",
				zap.Uint("task_id", task.ID),
				zap.Error(err))
		}

		// 2. 删除任务记录
		return uc.exportTaskRepo.Delete(txCtx, id)
	})
}

// CleanExpiredTasks 清理过期任务
// 分批循环查询并删除,确保所有过期任务都被处理
func (uc *ExportUseCase) CleanExpiredTasks(ctx context.Context) error {
	expireTime := time.Now().AddDate(0, 0, -uc.config.ExpireDays)

	filters := map[string]any{
		"status_in":        []enum.ExportTaskStatus{enum.ExportTaskStatusCompleted, enum.ExportTaskStatusFailed},
		"completed_at_end": expireTime,
	}

	const batchSize = 100

	// 循环查询并删除,直到没有更多过期任务
	for {
		// 每次查询一批
		tasks, _, err := uc.exportTaskRepo.List(ctx, filters, batchSize, 0, "")
		if err != nil {
			return errs.WrapInternal(err, "failed to list expired tasks")
		}

		// 没有更多任务,退出循环
		if len(tasks) == 0 {
			break
		}

		// 处理这一批任务
		for _, task := range tasks {
			if err := uc.cleanSingleTask(ctx, task); err != nil {
				// 记录错误但继续处理其他任务
				uc.logger.Error("failed to clean expired task",
					zap.Uint("id", task.ID),
					zap.Error(err))
				continue
			}
		}

		// 如果这批数据少于 batchSize,说明已经是最后一批
		if len(tasks) < batchSize {
			break
		}
	}

	return nil
}

// cleanSingleTask 清理单个过期任务
func (uc *ExportUseCase) cleanSingleTask(ctx context.Context, task *model.ExportTask) error {
	return uc.tx.Transaction(ctx, func(txCtx context.Context) error {
		// 1. 分离文件附件（blob 由定时任务清理）
		if err := uc.storageUC.Detach(txCtx, task); err != nil {
			uc.logger.Warn("failed to detach expired file",
				zap.Uint("task_id", task.ID),
				zap.Error(err))
		}

		// 2. 删除任务记录
		return uc.exportTaskRepo.Delete(txCtx, task.ID)
	})
}

// Start 启动导出任务,将状态从 Pending 更新为 Processing
// 使用事务和 FOR UPDATE 锁防止并发问题
func (uc *ExportUseCase) Start(ctx context.Context, id uint) error {
	return uc.tx.Transaction(ctx, func(txCtx context.Context) error {
		// 使用 FOR UPDATE 锁定记录
		task, err := uc.exportTaskRepo.FindByIDForUpdate(txCtx, id)
		if err != nil {
			return err
		}

		// 检查状态是否为 Pending
		if task.Status != enum.ExportTaskStatusPending {
			return errs.NewBusiness("导出任务状态不是待处理,无法启动").
				WithReason(errs.ReasonInvalidState).
				WithMeta("id", id).
				WithMeta("current_status", task.Status)
		}

		// 更新状态为 Processing
		now := time.Now().UTC()
		return uc.exportTaskRepo.Update(txCtx, task, map[string]any{
			"status":     enum.ExportTaskStatusProcessing,
			"started_at": now,
		})
	})
}

// Complete 完成导出任务,将状态从 Processing 更新为 Completed
// 使用事务和 FOR UPDATE 锁防止并发问题
func (uc *ExportUseCase) Complete(ctx context.Context, id uint) error {
	return uc.tx.Transaction(ctx, func(txCtx context.Context) error {
		// 使用 FOR UPDATE 锁定记录
		task, err := uc.exportTaskRepo.FindByIDForUpdate(txCtx, id)
		if err != nil {
			return err
		}

		// 检查状态是否为 Processing
		if task.Status != enum.ExportTaskStatusProcessing {
			return errs.NewBusiness("导出任务状态不是处理中,无法完成").
				WithReason(errs.ReasonInvalidState).
				WithMeta("id", id).
				WithMeta("current_status", task.Status)
		}

		// 更新为已完成
		now := time.Now().UTC()
		return uc.exportTaskRepo.Update(txCtx, task, map[string]any{
			"status":       enum.ExportTaskStatusCompleted,
			"completed_at": now,
		})
	})
}

// Fail 标记导出任务失败,将状态从 Processing 更新为 Failed
// 使用事务和 FOR UPDATE 锁防止并发问题
func (uc *ExportUseCase) Fail(ctx context.Context, id uint, errorMsg string) error {
	return uc.tx.Transaction(ctx, func(txCtx context.Context) error {
		// 使用 FOR UPDATE 锁定记录
		task, err := uc.exportTaskRepo.FindByIDForUpdate(txCtx, id)
		if err != nil {
			return err
		}

		// 检查状态是否为 Processing
		if task.Status != enum.ExportTaskStatusProcessing {
			return errs.NewBusiness("导出任务状态不是处理中,无法标记失败").
				WithReason(errs.ReasonInvalidState).
				WithMeta("id", id).
				WithMeta("current_status", task.Status)
		}

		// 更新为失败
		return uc.exportTaskRepo.Update(txCtx, task, map[string]any{
			"status":        enum.ExportTaskStatusFailed,
			"error_message": errorMsg,
		})
	})
}

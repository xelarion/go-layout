// Package tasks provides queue task implementations.
package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"

	"github.com/xuri/excelize/v2"
	"go.uber.org/zap"

	"github.com/xelarion/go-layout/internal/enum"
	"github.com/xelarion/go-layout/internal/task"
	"github.com/xelarion/go-layout/internal/task/queue"
	"github.com/xelarion/go-layout/internal/usecase"
	"github.com/xelarion/go-layout/pkg/errs"
)

var importCustomerExcelHeaders = []string{
	"客户编号",
	"客户名称",
	"客户类型",
	"联系人",
	"联系电话",
	"状态",
	"描述",
}

// ImportCustomerPayload is the payload for the import-customer task.
type ImportCustomerPayload struct {
	BatchRecordID uint `json:"batch_record_id"` // The batch record id to import
}

// ImportCustomerHandler is an import-customer queue task handler.
type ImportCustomerHandler struct {
	deps   *task.Dependencies
	logger *zap.Logger
}

// NewImportCustomerHandler creates a new import-customer handler.
func NewImportCustomerHandler(deps *task.Dependencies, logger *zap.Logger) *ImportCustomerHandler {
	return &ImportCustomerHandler{
		deps:   deps,
		logger: logger.Named("import-customer-handler"),
	}
}

// Register registers this handler with the queue.
func (t *ImportCustomerHandler) Register(q *queue.Manager) error {
	if err := q.RegisterConsumer(
		"import-customer-handler",
		"import-customer-queue",
		t.Execute,
		queue.WithConsumerOptionsConcurrency(1),
	); err != nil {
		return fmt.Errorf("failed to register import-customer task handler: %w", err)
	}
	return nil
}

// Execute runs the import-customer task handler.
func (t *ImportCustomerHandler) Execute(ctx context.Context, msg queue.Message) (queue.Action, error) {
	var payload ImportCustomerPayload
	if err := json.Unmarshal(msg.Body, &payload); err != nil {
		return queue.NackDiscard, errs.WrapInternal(err, "failed to unmarshal payload")
	}

	batchRecord, err := t.deps.BatchRecordUseCase.GetByID(ctx, payload.BatchRecordID)
	if err != nil {
		if errs.IsReason(err, errs.ReasonNotFound) {
			return queue.NackDiscard, err
		}
		return queue.NackRequeue, err
	}

	if batchRecord.IsStatusPending() {
		if err := t.deps.BatchRecordUseCase.Start(ctx, batchRecord.ID); err != nil {
			return queue.NackRequeue, err
		}
	} else if batchRecord.IsStatusProcessing() {
		// TODO 删除历史数据
	} else if batchRecord.IsStatusFinished() {
		return queue.NackDiscard, nil
	}

	attachment, err := t.deps.StorageUseCase.FindByResourceAndName(ctx, batchRecord, enum.StorageAttachmentNameTypeFile)
	if err != nil {
		return queue.NackDiscard, err
	}

	downloadURL, err := t.deps.StorageUseCase.GenerateDownloadURL(ctx, attachment.StorageBlob.Key)
	if err != nil {
		return queue.NackDiscard, err
	}

	if err := t.HandleExcel(ctx, downloadURL, batchRecord); err != nil {
		if errs.IsInternal(err) {
			return queue.NackRequeue, err
		}

		if errs.IsBusiness(err) {
			failureReason := err.Error()
			if err := t.deps.BatchRecordUseCase.Finish(ctx, payload.BatchRecordID, enum.BatchRecordStatusFailed, failureReason); err != nil {
				t.logger.Error("failed to finish batch record", zap.Error(err))
				return queue.NackDiscard, err
			}
		}

		return queue.NackDiscard, nil
	}

	if err := t.deps.BatchRecordUseCase.UpdateResult(ctx, payload.BatchRecordID); err != nil {
		t.logger.Error("failed to update batch record result", zap.Error(err))
		return queue.NackDiscard, err
	}

	return queue.Ack, nil
}

// HandleExcel downloads and reads the Excel file from the provided URL
func (t *ImportCustomerHandler) HandleExcel(ctx context.Context, downloadURL string, batchRecord *usecase.BatchRecord) error {
	// Download the Excel file
	xlsx, err := t.downloadExcelFile(ctx, downloadURL)
	if err != nil {
		return err
	}
	defer xlsx.Close()

	// Get all the sheet names
	sheets := xlsx.GetSheetList()
	if len(sheets) == 0 {
		return errs.NewBusiness("Excel文件中没有找到任何工作表")
	}

	// Process the first sheet using streaming API
	sheetName := sheets[0]
	_, err = t.processSheet(ctx, xlsx, sheetName, batchRecord)
	if err != nil {
		return err
	}

	return nil
}

// downloadExcelFile downloads an Excel file from the given URL and opens it
func (t *ImportCustomerHandler) downloadExcelFile(ctx context.Context, downloadURL string) (*excelize.File, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	if err != nil {
		return nil, errs.WrapInternal(err, "创建请求失败")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errs.WrapInternal(err, "下载文件失败")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errs.NewInternal(fmt.Sprintf("下载文件失败, 状态码: %s", resp.Status))
	}

	// Read the Excel file with streaming mode
	xlsx, err := excelize.OpenReader(resp.Body)
	if err != nil {
		return nil, errs.WrapInternal(err, "打开Excel文件失败")
	}

	return xlsx, nil
}

// processSheet processes a single sheet from the Excel file using streaming API
func (t *ImportCustomerHandler) processSheet(ctx context.Context, xlsx *excelize.File, sheetName string, batchRecord *usecase.BatchRecord) (int, error) {
	rows, err := xlsx.Rows(sheetName)
	if err != nil {
		return 0, errs.WrapInternal(err, fmt.Sprintf("获取工作表%s的流式行失败", sheetName))
	}
	defer rows.Close()

	// Process the header row first
	headers, err := t.processHeaderRow(rows)
	if err != nil {
		return 0, err
	}

	// Process data rows
	rowCount, err := t.processDataRows(ctx, rows, headers, batchRecord)
	if err != nil {
		return 0, err
	}

	// Check for any errors that occurred during row iteration
	if err := rows.Error(); err != nil {
		return 0, errs.WrapInternal(err, fmt.Sprintf("处理工作表%s的行时发生错误", sheetName))
	}

	return rowCount, nil
}

// processHeaderRow processes the header row of the Excel sheet
func (t *ImportCustomerHandler) processHeaderRow(rows *excelize.Rows) ([]string, error) {
	if !rows.Next() {
		return nil, errs.NewBusiness("Excel文件中没有找到头部行")
	}

	headers, err := rows.Columns()
	if err != nil {
		return nil, errs.WrapInternal(err, "读取头部行失败")
	}

	if len(headers) == 0 {
		return nil, errs.NewBusiness("头部行为空")
	}

	// is headers equal importCustomerExcelHeaders
	if !reflect.DeepEqual(headers, importCustomerExcelHeaders) {
		return nil, errs.NewBusiness("头部行不匹配")
	}

	return headers, nil
}

// processDataRows processes all data rows in batches
func (t *ImportCustomerHandler) processDataRows(ctx context.Context, rows *excelize.Rows, headers []string, batchRecord *usecase.BatchRecord) (int, error) {
	const batchSize = 1000
	rowCount := 0
	rowNumber := 1 // 头部占一行
	batch := make([][]string, 0, batchSize)

	for rows.Next() {
		// Check for context cancellation periodically
		select {
		case <-ctx.Done():
			return rowCount, ctx.Err()
		default:
		}

		rowNumber += 1

		// Get the row values
		row, err := rows.Columns()
		if err != nil {
			return rowCount, errs.WrapInternal(err, fmt.Sprintf("读取行%d失败", rowCount+1))
		}

		// Skip empty rows
		if isEmptyRow(row) {
			continue
		}

		batch = append(batch, row)
		rowCount++

		// Process the batch when it reaches batchSize
		if len(batch) >= batchSize {
			if err := t.processBatch(ctx, rowNumber-len(batch)+1, batch, headers, batchRecord); err != nil {
				return rowCount, errs.WrapBusiness(err, fmt.Sprintf("处理批次失败, 行%d", rowCount))
			}
			// Clear the batch but keep the underlying capacity
			batch = batch[:0]
		}
	}

	// Process any remaining rows in the final batch
	if len(batch) > 0 {
		if err := t.processBatch(ctx, rowNumber-len(batch)+1, batch, headers, batchRecord); err != nil {
			return rowCount, err
		}
	} else {
		return rowCount, errs.NewBusiness("Excel文件中没有找到任何数据行")
	}

	return rowCount, nil
}

// isEmptyRow checks if a row is empty (all cells are empty)
func isEmptyRow(row []string) bool {
	for _, cell := range row {
		if cell != "" {
			return false
		}
	}
	return true
}

// processBatch handles processing a batch of Excel rows
func (t *ImportCustomerHandler) processBatch(ctx context.Context, rowNumberStart int, rows [][]string, headers []string, batchRecord *usecase.BatchRecord) error {
	for i, row := range rows {
		if err := t.importRow(ctx, rowNumberStart+i, row, batchRecord); err != nil {
			return err
		}
	}

	return nil
}

// importRow imports a single row
func (t *ImportCustomerHandler) importRow(ctx context.Context, rowNumber int, rowData []string, batchRecord *usecase.BatchRecord) error {
	//row := excelx.NewRow(rowData)
	//
	//// 字段获取与验证
	//code := row.Col(0).Validate("客户编号", excelx.Required(), excelx.MaxLen(100)).String()
	//name := row.Col(1).Validate("客户名称", excelx.Required(), excelx.MaxLen(100)).String()
	//customerType := excelx.Val(row.Col(2).Validate("客户类型", excelx.Required(), excelx.In(map[string]any{
	//	"企业": enum.CustomerTypeEnterprise,
	//	"个人": enum.CustomerTypePersonal,
	//})), enum.CustomerType(0))
	//contactName := row.Col(3).Validate("联系人", excelx.Required(), excelx.MaxLen(100)).String()
	//contactPhone := row.Col(4).Validate("联系电话", excelx.Required(), excelx.Phone()).String()
	//enabled := row.Col(5).Validate("状态", excelx.Required(), excelx.In(map[string]any{"启用": true, "禁用": false})).Bool()
	//description := row.Col(6).Validate("备注", excelx.MaxLen(1000)).String()
	//
	//isFailed := row.HasError()
	//failureReason := strings.Join(row.Errors(), "；")
	//
	//params := usecase.CreateCustomerParams{
	//	CreateRecordCommonParams: usecase.CreateRecordCommonParams{
	//		CreatedBy:     batchRecord.CreatedBy,
	//		CreatedRoleID: batchRecord.CreatedRoleID,
	//		AssignedTo:    0,
	//	},
	//	Code:         code,
	//	Name:         name,
	//	CustomerType: customerType,
	//	ContactName:  contactName,
	//	ContactPhone: contactPhone,
	//	Enabled:      enabled,
	//	Description:  description,
	//}
	//
	//err := t.deps.TX.Transaction(ctx, func(ctx context.Context) error {
	//	if isFailed {
	//		// 添加批次详情
	//		if _, err := t.deps.BatchDetailUseCase.Create(ctx, usecase.CreateBatchDetailParams{
	//			BatchRecordID: batchRecord.ID,
	//			Key:           strconv.Itoa(rowNumber),
	//			Data:          datatypes.JSON(`{}`),
	//			Status:        enum.BatchDetailStatusFailed,
	//			FailureReason: failureReason,
	//		}); err != nil {
	//			return err
	//		}
	//
	//		return nil
	//	}
	//
	//	// 保存
	//	if _, err := t.deps.CustomerUseCase.Create(ctx, params); err != nil {
	//		return err
	//	}
	//
	//	// 添加批次详情
	//	if _, err := t.deps.BatchDetailUseCase.Create(ctx, usecase.CreateBatchDetailParams{
	//		BatchRecordID: batchRecord.ID,
	//		Key:           strconv.Itoa(rowNumber),
	//		Data:          datatypes.JSON(`{}`),
	//		Status:        enum.BatchDetailStatusSuccess,
	//	}); err != nil {
	//		return err
	//	}
	//
	//	return nil
	//})
	//
	//if err != nil {
	//	// 添加批次详情
	//	if _, err := t.deps.BatchDetailUseCase.Create(ctx, usecase.CreateBatchDetailParams{
	//		BatchRecordID: batchRecord.ID,
	//		Key:           strconv.Itoa(rowNumber),
	//		Data:          datatypes.JSON(`{}`),
	//		Status:        enum.BatchDetailStatusFailed,
	//		FailureReason: err.Error(),
	//	}); err != nil {
	//		return err
	//	}
	//}

	return nil
}

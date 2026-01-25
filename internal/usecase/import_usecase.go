// Package usecase contains business logic.
package usecase

import (
	"context"

	excelexporter "github.com/xelarion/excel-exporter"
	"github.com/xuri/excelize/v2"
	"gorm.io/datatypes"

	"github.com/xelarion/go-layout/internal/enum"
	"github.com/xelarion/go-layout/pkg/errs"
)

type ImportCommonParams struct {
	CreateRecordCommonParams
	BlobKey string
}

// ImportUseCase implements business logic for import operations.
type ImportUseCase struct {
	tx                 Transaction
	amqpMessageUseCase *AmqpMessageUseCase
	batchRecordUseCase *BatchRecordUseCase
	storageUseCase     *StorageUseCase
}

// NewImportUseCase creates a new instance of ImportUseCase.
func NewImportUseCase(tx Transaction, amqpMessageUseCase *AmqpMessageUseCase, batchRecordUseCase *BatchRecordUseCase, storageUseCase *StorageUseCase) *ImportUseCase {
	return &ImportUseCase{
		tx:                 tx,
		amqpMessageUseCase: amqpMessageUseCase,
		batchRecordUseCase: batchRecordUseCase,
		storageUseCase:     storageUseCase,
	}
}

func (uc *ImportUseCase) ImportCustomers(ctx context.Context, params ImportCommonParams) error {
	err := uc.tx.Transaction(ctx, func(ctx context.Context) error {
		batchRecord, err := uc.batchRecordUseCase.CreatePendingBatchRecord(
			ctx,
			params.CreatedBy,
			params.CreatedRoleID,
			enum.BatchRecordOperationTypeImportCustomer,
			datatypes.JSON(`{}`),
		)
		if err != nil {
			return err
		}

		// attach attachment to batch record
		if err := uc.storageUseCase.AttachOne(
			ctx,
			batchRecord,
			enum.StorageAttachmentNameTypeFile,
			params.BlobKey,
		); err != nil {
			return err
		}

		return uc.amqpMessageUseCase.PublishImportCustomers(ctx, batchRecord.ID)
	})
	if err != nil {
		return err
	}

	return nil
}

func (uc *ImportUseCase) ImportEquipments(ctx context.Context, params ImportCommonParams) error {
	err := uc.tx.Transaction(ctx, func(ctx context.Context) error {
		batchRecord, err := uc.batchRecordUseCase.CreatePendingBatchRecord(
			ctx,
			params.CreatedBy,
			params.CreatedRoleID,
			enum.BatchRecordOperationTypeImportEquipment,
			datatypes.JSON(`{}`),
		)
		if err != nil {
			return err
		}

		if err := uc.storageUseCase.AttachOne(
			ctx,
			batchRecord,
			enum.StorageAttachmentNameTypeFile,
			params.BlobKey,
		); err != nil {
			return err
		}

		return uc.amqpMessageUseCase.PublishImportEquipments(ctx, batchRecord.ID)
	})
	if err != nil {
		return err
	}

	return nil
}

func (uc *ImportUseCase) DownloadImportCustomersTemplate(ctx context.Context, fileName string) error {
	queryDataToChannelFunc := func(exporter *excelexporter.Exporter) func(dataCh chan excelexporter.Row) error {
		return func(dataCh chan excelexporter.Row) error {
			titleStyle, _ := exporter.File.NewStyle(
				&excelize.Style{
					Font:      &excelize.Font{Color: "000000", Size: 12},
					Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
				},
			)

			// Set column width
			_ = exporter.StreamWriter.SetColWidth(1, 7, 20)

			// Title
			dataCh <- excelexporter.Row{
				Cells: []excelize.Cell{
					{Value: "客户编号", StyleID: titleStyle},
					{Value: "客户名称", StyleID: titleStyle},
					{Value: "客户类型", StyleID: titleStyle},
					{Value: "联系人", StyleID: titleStyle},
					{Value: "联系电话", StyleID: titleStyle},
					{Value: "状态", StyleID: titleStyle},
					{Value: "描述", StyleID: titleStyle},
				},
			}

			return nil
		}
	}

	exporter := excelexporter.New(fileName, true)
	sheets := []excelexporter.SheetData{
		{
			Name:    "Sheet1",
			RowFunc: excelexporter.UseRowChan(queryDataToChannelFunc(exporter)),
		},
	}

	if err := exporter.Export(sheets); err != nil {
		return errs.NewInternal("生成Excel文件失败")
	}

	return nil
}

func (uc *ImportUseCase) ImportEquipmentTypes(ctx context.Context, params ImportCommonParams) error {
	err := uc.tx.Transaction(ctx, func(ctx context.Context) error {
		batchRecord, err := uc.batchRecordUseCase.CreatePendingBatchRecord(
			ctx,
			params.CreatedBy,
			params.CreatedRoleID,
			enum.BatchRecordOperationTypeImportEquipmentType,
			datatypes.JSON(`{}`),
		)
		if err != nil {
			return err
		}

		// attach attachment to batch record
		if err := uc.storageUseCase.AttachOne(
			ctx,
			batchRecord,
			enum.StorageAttachmentNameTypeFile,
			params.BlobKey,
		); err != nil {
			return err
		}

		return uc.amqpMessageUseCase.PublishImportEquipmentTypes(ctx, batchRecord.ID)
	})
	if err != nil {
		return err
	}

	return nil
}

func (uc *ImportUseCase) DownloadImportEquipmentTypesTemplate(ctx context.Context, fileName string) error {
	queryDataToChannelFunc := func(exporter *excelexporter.Exporter) func(dataCh chan excelexporter.Row) error {
		return func(dataCh chan excelexporter.Row) error {
			titleStyle, _ := exporter.File.NewStyle(
				&excelize.Style{
					Font:      &excelize.Font{Color: "000000", Size: 12},
					Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
				},
			)

			// Set column width
			_ = exporter.StreamWriter.SetColWidth(1, 5, 20)

			// Title
			dataCh <- excelexporter.Row{
				Cells: []excelize.Cell{
					{Value: "设备类型编号", StyleID: titleStyle},
					{Value: "设备类型名称", StyleID: titleStyle},
					{Value: "上游类型名称", StyleID: titleStyle},
					{Value: "状态", StyleID: titleStyle},
					{Value: "备注", StyleID: titleStyle},
				},
			}

			return nil
		}
	}

	exporter := excelexporter.New(fileName, true)
	sheets := []excelexporter.SheetData{
		{
			Name:    "Sheet1",
			RowFunc: excelexporter.UseRowChan(queryDataToChannelFunc(exporter)),
		},
	}

	if err := exporter.Export(sheets); err != nil {
		return errs.NewInternal("生成Excel文件失败")
	}

	return nil
}

func (uc *ImportUseCase) DownloadImportEquipmentsTemplate(ctx context.Context, fileName string) error {
	queryDataToChannelFunc := func(exporter *excelexporter.Exporter) func(dataCh chan excelexporter.Row) error {
		return func(dataCh chan excelexporter.Row) error {
			titleStyle, _ := exporter.File.NewStyle(
				&excelize.Style{
					Font:      &excelize.Font{Color: "000000", Size: 12},
					Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
				},
			)

			_ = exporter.StreamWriter.SetColWidth(1, 8, 20)

			dataCh <- excelexporter.Row{
				Cells: []excelize.Cell{
					{Value: "设备编号", StyleID: titleStyle},
					{Value: "设备名称", StyleID: titleStyle},
					{Value: "规格型号", StyleID: titleStyle},
					{Value: "品牌", StyleID: titleStyle},
					{Value: "所属车间", StyleID: titleStyle},
					{Value: "设备类型", StyleID: titleStyle},
					{Value: "备注", StyleID: titleStyle},
				},
			}

			return nil
		}
	}

	exporter := excelexporter.New(fileName, true)
	sheets := []excelexporter.SheetData{
		{
			Name:    "Sheet1",
			RowFunc: excelexporter.UseRowChan(queryDataToChannelFunc(exporter)),
		},
	}

	if err := exporter.Export(sheets); err != nil {
		return errs.NewInternal("生成Excel文件失败")
	}

	return nil
}

func (uc *ImportUseCase) ImportSuppliers(ctx context.Context, params ImportCommonParams) error {
	err := uc.tx.Transaction(ctx, func(ctx context.Context) error {
		batchRecord, err := uc.batchRecordUseCase.CreatePendingBatchRecord(
			ctx,
			params.CreatedBy,
			params.CreatedRoleID,
			enum.BatchRecordOperationTypeImportSupplier,
			datatypes.JSON(`{}`),
		)
		if err != nil {
			return err
		}

		// attach attachment to batch record
		if err := uc.storageUseCase.AttachOne(
			ctx,
			batchRecord,
			enum.StorageAttachmentNameTypeFile,
			params.BlobKey,
		); err != nil {
			return err
		}

		return uc.amqpMessageUseCase.PublishImportSuppliers(ctx, batchRecord.ID)
	})
	if err != nil {
		return err
	}

	return nil
}

func (uc *ImportUseCase) DownloadImportSuppliersTemplate(ctx context.Context, fileName string) error {
	queryDataToChannelFunc := func(exporter *excelexporter.Exporter) func(dataCh chan excelexporter.Row) error {
		return func(dataCh chan excelexporter.Row) error {
			titleStyle, _ := exporter.File.NewStyle(
				&excelize.Style{
					Font:      &excelize.Font{Color: "000000", Size: 12},
					Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
				},
			)

			// Set column width
			_ = exporter.StreamWriter.SetColWidth(1, 7, 20)

			// Title
			dataCh <- excelexporter.Row{
				Cells: []excelize.Cell{
					{Value: "供应商编号", StyleID: titleStyle},
					{Value: "供应商名称", StyleID: titleStyle},
					{Value: "供应商等级", StyleID: titleStyle},
					{Value: "联系人", StyleID: titleStyle},
					{Value: "联系电话", StyleID: titleStyle},
					{Value: "状态", StyleID: titleStyle},
					{Value: "备注", StyleID: titleStyle},
				},
			}

			return nil
		}
	}

	exporter := excelexporter.New(fileName, true)
	sheets := []excelexporter.SheetData{
		{
			Name:    "Sheet1",
			RowFunc: excelexporter.UseRowChan(queryDataToChannelFunc(exporter)),
		},
	}

	if err := exporter.Export(sheets); err != nil {
		return errs.NewInternal("生成Excel文件失败")
	}

	return nil
}

package enum

import "github.com/xelarion/go-layout/pkg/errs"

type BatchRecordStatus int8

const (
	BatchRecordStatusPending    BatchRecordStatus = iota // 待处理
	BatchRecordStatusProcessing                          // 处理中
	BatchRecordStatusSuccess                             // 成功
	BatchRecordStatusFailed                              // 失败
)

var BatchRecordStatusSlug = map[BatchRecordStatus]string{
	BatchRecordStatusPending:    "pending",
	BatchRecordStatusProcessing: "processing",
	BatchRecordStatusSuccess:    "success",
	BatchRecordStatusFailed:     "failed",
}

var BatchRecordStatusText = map[BatchRecordStatus]string{
	BatchRecordStatusPending:    "待处理",
	BatchRecordStatusProcessing: "处理中",
	BatchRecordStatusSuccess:    "成功",
	BatchRecordStatusFailed:     "失败",
}

func GetBatchRecordStatus(slug string) (BatchRecordStatus, error) {
	for k, v := range BatchRecordStatusSlug {
		if v == slug {
			return k, nil
		}
	}
	return 0, errs.NewBusiness("invalid batch record status slug")
}

// BatchRecordOperationType 批量操作类型
type BatchRecordOperationType int8

const (
	BatchRecordOperationTypeImportCustomer      BatchRecordOperationType = iota + 1 // 导入客户
	BatchRecordOperationTypeImportEquipment                                         // 导入设备
	BatchRecordOperationTypeImportEquipmentType                                     // 导入设备类型
	BatchRecordOperationTypeImportSupplier
)

var BatchRecordOperationTypeSlug = map[BatchRecordOperationType]string{
	BatchRecordOperationTypeImportCustomer:      "import_customer",
	BatchRecordOperationTypeImportSupplier:      "import_supplier",
	BatchRecordOperationTypeImportEquipment:     "import_equipment",
	BatchRecordOperationTypeImportEquipmentType: "import_equipment_type",
}

var BatchRecordOperationTypeText = map[BatchRecordOperationType]string{
	BatchRecordOperationTypeImportCustomer:      "导入客户",
	BatchRecordOperationTypeImportEquipment:     "导入设备",
	BatchRecordOperationTypeImportEquipmentType: "导入设备类型",
	BatchRecordOperationTypeImportSupplier:      "导入供应商",
}

func GetBatchRecordOperationType(slug string) (BatchRecordOperationType, error) {
	for k, v := range BatchRecordOperationTypeSlug {
		if v == slug {
			return k, nil
		}
	}
	return 0, errs.NewBusiness("invalid batch record operation type slug")
}

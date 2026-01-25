package enum

import "github.com/xelarion/go-layout/pkg/errs"

// ExportTaskStatus 导出任务状态
type ExportTaskStatus int8

const (
	ExportTaskStatusPending    ExportTaskStatus = 1 // 待处理
	ExportTaskStatusProcessing ExportTaskStatus = 2 // 处理中
	ExportTaskStatusCompleted  ExportTaskStatus = 3 // 已完成
	ExportTaskStatusFailed     ExportTaskStatus = 4 // 失败
)

var ExportTaskStatusSlug = map[ExportTaskStatus]string{
	ExportTaskStatusPending:    "pending",
	ExportTaskStatusProcessing: "processing",
	ExportTaskStatusCompleted:  "completed",
	ExportTaskStatusFailed:     "failed",
}

var ExportTaskStatusText = map[ExportTaskStatus]string{
	ExportTaskStatusPending:    "待处理",
	ExportTaskStatusProcessing: "处理中",
	ExportTaskStatusCompleted:  "已完成",
	ExportTaskStatusFailed:     "失败",
}

func GetExportTaskStatus(slug string) (ExportTaskStatus, error) {
	for k, v := range ExportTaskStatusSlug {
		if v == slug {
			return k, nil
		}
	}
	return 0, errs.NewBusiness("invalid ExportTaskStatus")
}

// ExportTaskType 导出任务类型
type ExportTaskType string

const (
	ExportTaskTypeUser                ExportTaskType = "user"                  // 用户导出
	ExportTaskTypeProductionPlan      ExportTaskType = "production_plan"       // 生产计划导出
	ExportTaskTypeBomProcessComponent ExportTaskType = "bom_process_component" // BOM管理-工艺路线-工序用料导出
	ExportTaskTypeProcessRouteStep    ExportTaskType = "process_route_step"    // 工艺路线-工序组成导出
	ExportTaskTypeProcessRoute        ExportTaskType = "process_route"         // 工艺路线导出
	ExportTaskTypeProcess             ExportTaskType = "process"               // 工序导出
	ExportTaskTypeBomComponent        ExportTaskType = "bom_component"         // 物料配件导出
	ExportTaskTypeMaterial            ExportTaskType = "material"              // 物料导出
	ExportTaskTypeSupplier            ExportTaskType = "supplier"              // 供应商导出
	ExportTaskTypeCustomer            ExportTaskType = "customer"              // 客户导出
	ExportTaskTypeEquipmentType       ExportTaskType = "equipment_type"
	ExportTaskTypeEquipment           ExportTaskType = "equipment" // 设备导出
	ExportTaskTypeProduct             ExportTaskType = "product"
)

var ExportTaskTypeSlug = map[ExportTaskType]string{
	ExportTaskTypeUser:                "user",
	ExportTaskTypeProductionPlan:      "production_plan",
	ExportTaskTypeBomProcessComponent: "bom_process_component",
	ExportTaskTypeProcessRouteStep:    "process_route_step",
	ExportTaskTypeProcessRoute:        "process_route",
	ExportTaskTypeProcess:             "process",
	ExportTaskTypeBomComponent:        "bom_component",
	ExportTaskTypeMaterial:            "material",
	ExportTaskTypeSupplier:            "supplier",
	ExportTaskTypeCustomer:            "customer",
	ExportTaskTypeEquipmentType:       "equipment_type",
	ExportTaskTypeEquipment:           "equipment",
	ExportTaskTypeProduct:             "product",
}

var ExportTaskTypeText = map[ExportTaskType]string{
	ExportTaskTypeUser:                "用户导出",
	ExportTaskTypeProductionPlan:      "生产计划导出",
	ExportTaskTypeBomProcessComponent: "BOM管理-工艺路线-工序用料导出",
	ExportTaskTypeProcessRouteStep:    "工艺路线-工序组成导出",
	ExportTaskTypeProcessRoute:        "工艺路线导出",
	ExportTaskTypeProcess:             "工序导出",
	ExportTaskTypeBomComponent:        "物料配件导出",
	ExportTaskTypeMaterial:            "物料导出",
	ExportTaskTypeSupplier:            "供应商导出",
	ExportTaskTypeCustomer:            "客户导出",
	ExportTaskTypeEquipmentType:       "设备类型导出",
	ExportTaskTypeEquipment:           "设备导出",
	ExportTaskTypeProduct:             "产品导出",
}

func GetExportTaskType(slug string) (ExportTaskType, error) {
	for k, v := range ExportTaskTypeSlug {
		if v == slug {
			return k, nil
		}
	}
	return "", errs.NewBusiness("invalid ExportTaskType")
}

package usecase

//
//import (
//	"context"
//	"fmt"
//
//	"github.com/xelarion/go-layout/internal/enum"
//	"github.com/xelarion/go-layout/internal/model"
//	"github.com/xelarion/go-layout/pkg/errs"
//)
//
//// ChangeLogBomComponentDisplay 物料配件显示数据
//type ChangeLogBomComponentDisplay struct {
//	ID           uint
//	ResourceType string
//	ResourceID   uint
//	Name         string
//	Code         string
//	Type         string
//	SpecModel    string
//	Unit         string
//}
//
//// ChangeLogBomProcessComponentDisplay 工序用料显示数据
//type ChangeLogBomProcessComponentDisplay struct {
//	ID                          uint
//	ProcessRouteStepProcessName string
//	ResourceType                string
//	ResourceID                  uint
//	Name                        string
//	Code                        string
//	Type                        string
//	SpecModel                   string
//	Unit                        string
//	Quantity                    float64
//	QuantityRatio               float64
//}
//
//// ChangeLogBomTracker BOM 变更追踪器
//type ChangeLogBomTracker struct {
//	OperatorID uint
//	Action     enum.ChangeLogAction
//
//	// 主表(必须命名为 Record 和 OldRecord，供反射使用)
//	Record    *model.Bom
//	OldRecord *model.Bom
//
//	// 关联表(一对一)
//	ProcessRoute    *model.ProcessRoute
//	OldProcessRoute *model.ProcessRoute
//
//	// 关联表(一对多)
//	BomComponents    []*ChangeLogBomComponentDisplay
//	OldBomComponents []*ChangeLogBomComponentDisplay
//
//	BomProcessComponents    []*ChangeLogBomProcessComponentDisplay
//	OldBomProcessComponents []*ChangeLogBomProcessComponentDisplay
//}
//
//// NewChangeLogTracker 创建变更追踪器
//func (uc *BomUseCase) NewChangeLogTracker(operatorID uint, action enum.ChangeLogAction) *ChangeLogBomTracker {
//	return &ChangeLogBomTracker{
//		OperatorID: operatorID,
//		Action:     action,
//	}
//}
//
//// SetChangeLogOldData 设置旧数据(在更新/删除前调用)
//func (uc *BomUseCase) SetChangeLogOldData(ctx context.Context, tracker *ChangeLogBomTracker, bom *model.Bom) error {
//	tracker.OldRecord = bom.DeepCopy()
//
//	if bom.ProcessRouteID != nil {
//		processRoute, err := uc.processRouteRepo.FindByID(ctx, *bom.ProcessRouteID)
//		if err != nil && !errs.IsReason(err, errs.ReasonNotFound) {
//			return err
//		}
//		tracker.OldProcessRoute = processRoute
//	}
//
//	components, err := uc.getChangeLogBomComponentDisplay(ctx, bom.ID)
//	if err != nil {
//		return err
//	}
//	tracker.OldBomComponents = components
//
//	processComponents, err := uc.getChangeLogBomProcessComponentDisplay(ctx, bom.ID)
//	if err != nil {
//		return err
//	}
//	tracker.OldBomProcessComponents = processComponents
//
//	return nil
//}
//
//// SetChangeLogNewData 设置新数据(在创建/更新后调用)
//func (uc *BomUseCase) SetChangeLogNewData(ctx context.Context, tracker *ChangeLogBomTracker, bom *model.Bom) error {
//	tracker.Record = bom
//
//	if bom.ProcessRouteID != nil {
//		processRoute, err := uc.processRouteRepo.FindByID(ctx, *bom.ProcessRouteID)
//		if err != nil && !errs.IsReason(err, errs.ReasonNotFound) {
//			return err
//		}
//		tracker.ProcessRoute = processRoute
//	}
//
//	components, err := uc.getChangeLogBomComponentDisplay(ctx, bom.ID)
//	if err != nil {
//		return err
//	}
//	tracker.BomComponents = components
//
//	processComponents, err := uc.getChangeLogBomProcessComponentDisplay(ctx, bom.ID)
//	if err != nil {
//		return err
//	}
//	tracker.BomProcessComponents = processComponents
//
//	return nil
//}
//
//// SaveChangeLog 保存变更记录
//func (uc *BomUseCase) SaveChangeLog(ctx context.Context, tracker *ChangeLogBomTracker) error {
//	if tracker.Record == nil && tracker.OldRecord == nil {
//		return fmt.Errorf("记录不能为空")
//	}
//
//	var resourceType string
//	var resourceID uint
//	if tracker.Record != nil {
//		resourceType = tracker.Record.GetResourceType()
//		resourceID = tracker.Record.GetResourceID()
//	} else {
//		resourceType = tracker.OldRecord.GetResourceType()
//		resourceID = tracker.OldRecord.GetResourceID()
//	}
//
//	return uc.changeLogUseCase.SaveChangeLog(
//		ctx,
//		tracker,
//		resourceType,
//		resourceID,
//		tracker.OperatorID,
//		tracker.Action,
//		uc.buildChangeLogFieldConfigs(ctx),
//	)
//}
//
//// getChangeLogBomComponentDisplay 获取物料配件显示数据(用于变更记录)
//func (uc *BomUseCase) getChangeLogBomComponentDisplay(ctx context.Context, bomID uint) ([]*ChangeLogBomComponentDisplay, error) {
//	components, _, err := uc.bomComponentUseCase.List(ctx, map[string]any{"bom_id": bomID}, -1, -1, "")
//	if err != nil {
//		return nil, err
//	}
//	result := make([]*ChangeLogBomComponentDisplay, 0, len(components))
//	for _, c := range components {
//		result = append(result, &ChangeLogBomComponentDisplay{
//			ID:           c.ID,
//			ResourceType: string(c.ResourceType),
//			ResourceID:   c.ResourceID,
//			Name:         c.Name,
//			Code:         c.Code,
//			Type:         c.Type,
//			SpecModel:    c.SpecModel,
//			Unit:         c.Unit,
//		})
//	}
//	return result, nil
//}
//
//// getChangeLogBomProcessComponentDisplay 获取工序用料显示数据(用于变更记录)
//func (uc *BomUseCase) getChangeLogBomProcessComponentDisplay(ctx context.Context, bomID uint) ([]*ChangeLogBomProcessComponentDisplay, error) {
//	components, _, err := uc.bomProcessComponentUseCase.List(ctx, map[string]any{"bom_id": bomID}, -1, -1, "")
//	if err != nil {
//		return nil, err
//	}
//	result := make([]*ChangeLogBomProcessComponentDisplay, 0, len(components))
//	for _, c := range components {
//		processRouteStepProcessName := ""
//		processRouteStep, err := uc.processRouteStepRepo.FindByID(ctx, c.ProcessRouteStepID)
//		if err != nil {
//			if !errs.IsReason(err, errs.ReasonNotFound) {
//				return nil, err
//			}
//		}
//		if processRouteStep != nil {
//			processRouteStepProcessName = processRouteStep.Process.Name
//		}
//
//		result = append(result, &ChangeLogBomProcessComponentDisplay{
//			ID:                          c.ID,
//			ProcessRouteStepProcessName: processRouteStepProcessName,
//			ResourceType:                string(c.ResourceType),
//			ResourceID:                  c.ResourceID,
//			Name:                        c.Name,
//			Code:                        c.Code,
//			Type:                        c.Type,
//			SpecModel:                   c.SpecModel,
//			Unit:                        c.Unit,
//			Quantity:                    c.Quantity,
//			QuantityRatio:               c.QuantityRatio,
//		})
//	}
//	return result, nil
//}
//
//// buildChangeLogFieldConfigs 构建变更记录字段配置
//func (uc *BomUseCase) buildChangeLogFieldConfigs(ctx context.Context) []*ChangeLogFieldConfig {
//	return []*ChangeLogFieldConfig{
//		// 普通字段
//		{Name: "Version", Label: "BOM版本号"},
//		{Name: "Code", Label: "BOM编号"},
//		{Name: "Remark", Label: "备注"},
//		{
//			Name:  "IsDefault",
//			Label: "是否默认",
//			Format: func(_ interface{}, value interface{}) (string, error) {
//				if val, ok := value.(bool); ok {
//					if val {
//						return "是", nil
//					}
//					return "否", nil
//				}
//				return "", nil
//			},
//		},
//		{
//			Name:        "DailyOutput",
//			Label:       "日产量",
//			IsComposite: true,
//			CompositeFunc: func(values []string) string {
//				if len(values) < 2 || values[0] == "" {
//					return ""
//				}
//				return values[0] + values[1]
//			},
//			Children: []*ChangeLogFieldConfig{
//				{
//					Name:  "DailyOutput",
//					Label: "",
//					Format: func(_ interface{}, value interface{}) (string, error) {
//						return uc.changeLogUseCase.FormatPointerValue(nil, value), nil
//					},
//				},
//				{
//					Name:  "ProductID",
//					Label: "",
//					Format: func(_ interface{}, value interface{}) (string, error) {
//						productID, ok := value.(uint)
//						if !ok {
//							return "", nil
//						}
//						product, err := uc.productRepo.FindByID(ctx, productID)
//						if err != nil {
//							if errs.IsReason(err, errs.ReasonNotFound) {
//								return "", nil
//							}
//							return "", err
//						}
//						unit, err := uc.unitRepo.FindByID(ctx, product.UnitID)
//						if err != nil {
//							if errs.IsReason(err, errs.ReasonNotFound) {
//								return "", nil
//							}
//							return "", err
//						}
//						return unit.Name, nil
//					},
//				},
//			},
//		},
//		// 产品: 查询数据库获取产品名称
//		{
//			Name:  "ProductID",
//			Label: "产品",
//			Format: func(_ interface{}, value interface{}) (string, error) {
//				productID, ok := value.(uint)
//				if !ok {
//					return "", nil
//				}
//				product, err := uc.productRepo.FindByID(ctx, productID)
//				if err != nil {
//					if errs.IsReason(err, errs.ReasonNotFound) {
//						return "", nil
//					}
//					return "", err
//				}
//				return fmt.Sprintf("%s(%s)", product.Name, product.Code), nil
//			},
//		},
//		// 工艺路线: 一对一关联
//		{
//			Label:        "工艺路线",
//			Relation:     "ProcessRoute",
//			RelationType: enum.ChangeLogDetailRelationTypeOne,
//			Children: []*ChangeLogFieldConfig{
//				{Name: "Name", Label: "工艺路线名称"},
//				{Name: "Code", Label: "工艺路线编号"},
//			},
//		},
//		// 物料配件: 一对多关联
//		{
//			Label:        "物料配件",
//			Relation:     "BomComponents",
//			RelationType: enum.ChangeLogDetailRelationTypeMany,
//			Children: []*ChangeLogFieldConfig{
//				{Name: "ID", Label: ""},
//				{Name: "Name", Label: "物料配件名称"},
//				{Name: "Code", Label: "物料配件编号"},
//				{Name: "Type", Label: "类型"},
//				{Name: "SpecModel", Label: "规格型号"},
//				{Name: "Unit", Label: "单位"},
//			},
//		},
//		// 工序用料: 一对多关联,按工序分组显示
//		{
//			Label:           "工艺流程",
//			Relation:        "BomProcessComponents",
//			RelationType:    enum.ChangeLogDetailRelationTypeMany,
//			RelationGroupBy: "ProcessRouteStepProcessName",
//			Children: []*ChangeLogFieldConfig{
//				{Name: "ID", Label: ""},
//				{Name: "ProcessRouteStepProcessName", Label: ""},
//				{Name: "Name", Label: "物料配件名称"},
//				{Name: "Code", Label: "物料配件编号"},
//				{Name: "Type", Label: "类型"},
//				{Name: "SpecModel", Label: "规格型号"},
//				{Name: "Unit", Label: "单位"},
//				{Name: "Quantity", Label: "使用量"},
//				{Name: "QuantityRatio", Label: "用料比例"},
//			},
//		},
//	}
//}

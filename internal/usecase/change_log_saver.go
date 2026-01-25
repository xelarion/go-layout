package usecase

import (
	"context"
	"fmt"
	"reflect"

	"gorm.io/datatypes"

	"github.com/xelarion/go-layout/internal/enum"
	"github.com/xelarion/go-layout/internal/model"
	"github.com/xelarion/go-layout/internal/model/gen"
)

// ChangeLogChangeDetail 变更详情(内存中间结构)
type ChangeLogChangeDetail struct {
	FieldName    string
	FieldLabel   string
	FieldType    enum.ChangeLogDetailFieldType
	RelationType enum.ChangeLogDetailRelationType
	RelationMeta *model.ChangeLogDetailRelationMeta
	OldValue     string
	NewValue     string
}

// ChangeLogFieldConfig 字段配置
//
// 使用方式:
//   - 普通字段: Name + Label (+ Format)
//   - 组合字段: Label + IsComposite + CompositeFunc + Children
//   - 一对一关联: Label + Relation + RelationType=One + Children
//   - 一对多关联: Label + Relation + RelationType=Many + Children (+ RelationGroupBy)
type ChangeLogFieldConfig struct {
	// === 基础配置(所有类型都可能用到) ===
	Name   string                                                       // 字段名（普通字段必填,对应结构体字段名）
	Label  string                                                       // 字段显示名称（必填）
	Format func(tracker interface{}, value interface{}) (string, error) // 格式化函数（可选,nil 时使用默认格式化）

	// === 组合字段配置(IsComposite=true 时使用) ===
	IsComposite   bool                         // 是否为组合字段
	CompositeFunc func(values []string) string // 组合格式化函数

	// === 关联字段配置(Relation 非空时使用) ===
	Relation        string                           // 关联字段名（tracker 中的字段,如 "ProcessRoute", "BomComponents"）
	RelationType    enum.ChangeLogDetailRelationType // 关联类型: One=一对一, Many=一对多
	RelationGroupBy string                           // 分组字段名（仅一对多关联使用,如 "ProcessRouteStepProcessName"）

	Children []*ChangeLogFieldConfig // 子字段配置（组合字段和关联字段都使用）
}

// SaveChangeLog 保存变更记录的通用方法
func (uc *ChangeLogUseCase) SaveChangeLog(
	ctx context.Context,
	tracker interface{},
	resourceType string,
	resourceID uint,
	operatorID uint,
	action enum.ChangeLogAction,
	fieldConfigs []*ChangeLogFieldConfig,
) error {
	details := uc.collectChanges(tracker, fieldConfigs)

	if len(details) == 0 {
		return nil
	}

	changeLog := &model.ChangeLog{
		ChangeLog: gen.ChangeLog{
			ResourceType: resourceType,
			ResourceID:   resourceID,
			OperatorID:   operatorID,
		},
		Action: action,
	}

	if err := uc.changeLogRepo.Create(ctx, changeLog); err != nil {
		return err
	}

	changeLogDetails := make([]*model.ChangeLogDetail, 0, len(details))
	for _, detail := range details {
		d := &model.ChangeLogDetail{
			ChangeLogDetail: gen.ChangeLogDetail{
				ChangeLogID: changeLog.ID,
				FieldName:   detail.FieldName,
				FieldLabel:  detail.FieldLabel,
				OldValue:    detail.OldValue,
				NewValue:    detail.NewValue,
			},
			FieldType:    detail.FieldType,
			RelationType: detail.RelationType,
		}
		if detail.RelationMeta != nil {
			d.RelationMeta = datatypes.NewJSONType(*detail.RelationMeta)
		}
		changeLogDetails = append(changeLogDetails, d)
	}

	return uc.changeLogDetailRepo.CreateInBatches(ctx, changeLogDetails, 100)
}

// collectChanges 收集所有变更
func (uc *ChangeLogUseCase) collectChanges(
	tracker interface{},
	fieldConfigs []*ChangeLogFieldConfig,
) []*ChangeLogChangeDetail {
	var details []*ChangeLogChangeDetail

	trackerValue := reflect.ValueOf(tracker)
	if trackerValue.Kind() == reflect.Ptr {
		trackerValue = trackerValue.Elem()
	}

	// 分离不同类型的字段
	var mainFields []*ChangeLogFieldConfig    // 普通字段和组合字段
	var oneRelations []*ChangeLogFieldConfig  // 一对一关联
	var manyRelations []*ChangeLogFieldConfig // 一对多/多对多关联
	for _, config := range fieldConfigs {
		if config.Relation != "" {
			if config.RelationType == enum.ChangeLogDetailRelationTypeOne {
				oneRelations = append(oneRelations, config)
			} else if config.RelationType == enum.ChangeLogDetailRelationTypeMany {
				manyRelations = append(manyRelations, config)
			}
		} else {
			mainFields = append(mainFields, config)
		}
	}

	// 1. 处理主表字段变更（普通字段和组合字段）
	recordField := trackerValue.FieldByName("Record")
	oldRecordField := trackerValue.FieldByName("OldRecord")

	hasRecord := recordField.IsValid() && !recordField.IsNil()
	hasOldRecord := oldRecordField.IsValid() && !oldRecordField.IsNil()

	if hasRecord && hasOldRecord {
		// Update: 比较新旧记录
		details = append(details, uc.compareFields(
			tracker,
			recordField.Interface(),
			oldRecordField.Interface(),
			mainFields,
		)...)
	} else if hasRecord && !hasOldRecord {
		// Create: 记录所有新值
		details = append(details, uc.extractFieldValues(
			tracker,
			recordField.Interface(),
			mainFields,
			false, // isOld=false，提取新值
		)...)
	} else if !hasRecord && hasOldRecord {
		// Delete: 记录所有旧值
		details = append(details, uc.extractFieldValues(
			tracker,
			oldRecordField.Interface(),
			mainFields,
			true, // isOld=true，提取旧值
		)...)
	}

	// 2. 处理一对一关联变更
	for _, config := range oneRelations {
		relationField := trackerValue.FieldByName(config.Relation)
		oldRelationField := trackerValue.FieldByName("Old" + config.Relation)

		if !relationField.IsValid() || !oldRelationField.IsValid() {
			continue
		}

		hasRelation := !relationField.IsNil()
		hasOldRelation := !oldRelationField.IsNil()

		if hasRelation && hasOldRelation {
			// Update: 比较新旧关联
			details = append(details, uc.compareOneToOneRelation(
				tracker,
				config,
				relationField,
				oldRelationField,
			)...)
		} else if hasRelation && !hasOldRelation {
			// Create: 提取新关联值
			details = append(details, uc.extractOneToOneRelationValues(
				tracker,
				config,
				relationField,
				false,
			)...)
		} else if !hasRelation && hasOldRelation {
			// Delete: 提取旧关联值
			details = append(details, uc.extractOneToOneRelationValues(
				tracker,
				config,
				oldRelationField,
				true,
			)...)
		}
	}

	// 3. 处理一对多/多对多关联变更
	for _, config := range manyRelations {
		relationField := trackerValue.FieldByName(config.Relation)
		oldRelationField := trackerValue.FieldByName("Old" + config.Relation)

		if !relationField.IsValid() || !oldRelationField.IsValid() {
			continue
		}

		// 检查是否有变更（包括 Create/Delete 场景）
		hasChanges := !reflect.DeepEqual(relationField.Interface(), oldRelationField.Interface())
		if hasChanges {
			// 从 FieldConfig.Children 构建 schema
			schema := make([]model.ChangeLogDetailSchemaField, 0, len(config.Children))
			for _, child := range config.Children {
				if child.Label == "" {
					continue
				}
				schema = append(schema, model.ChangeLogDetailSchemaField{
					Name:  child.Name,
					Label: child.Label,
				})
			}

			details = append(details, &ChangeLogChangeDetail{
				FieldName:    config.Relation,
				FieldLabel:   config.Label,
				FieldType:    enum.ChangeLogDetailFieldTypeRelation,
				RelationType: enum.ChangeLogDetailRelationTypeMany,
				RelationMeta: &model.ChangeLogDetailRelationMeta{
					Schema:  schema,
					GroupBy: config.RelationGroupBy,
				},
				OldValue: uc.formatManyRelation(tracker, config, oldRelationField.Interface()),
				NewValue: uc.formatManyRelation(tracker, config, relationField.Interface()),
			})
		}
	}

	return details
}

// compareOneToOneRelation 比较一对一关联的子字段
func (uc *ChangeLogUseCase) compareOneToOneRelation(
	tracker interface{},
	config *ChangeLogFieldConfig,
	newField, oldField reflect.Value,
) []*ChangeLogChangeDetail {
	var details []*ChangeLogChangeDetail

	// 处理指针
	newVal := newField
	oldVal := oldField
	if newVal.Kind() == reflect.Ptr {
		if newVal.IsNil() {
			newVal = reflect.Value{}
		} else {
			newVal = newVal.Elem()
		}
	}
	if oldVal.Kind() == reflect.Ptr {
		if oldVal.IsNil() {
			oldVal = reflect.Value{}
		} else {
			oldVal = oldVal.Elem()
		}
	}

	// 比较每个子字段
	for _, child := range config.Children {
		var oldValue, newValue string

		if oldVal.IsValid() {
			if field := oldVal.FieldByName(child.Name); field.IsValid() {
				oldValue = uc.formatValue(tracker, child, field.Interface())
			}
		}
		if newVal.IsValid() {
			if field := newVal.FieldByName(child.Name); field.IsValid() {
				newValue = uc.formatValue(tracker, child, field.Interface())
			}
		}

		if oldValue != newValue {
			// FieldName 使用 Relation.ChildName 格式
			fieldName := config.Relation + "." + child.Name
			// Label 优先用 child.Label,否则用 config.Label
			label := child.Label
			if label == "" {
				label = config.Label
			}
			// 一对一关联不需要 schema，因为是单个字段
			details = append(details, &ChangeLogChangeDetail{
				FieldName:    fieldName,
				FieldLabel:   label,
				FieldType:    enum.ChangeLogDetailFieldTypeRelation,
				RelationType: enum.ChangeLogDetailRelationTypeOne,
				OldValue:     oldValue,
				NewValue:     newValue,
			})
		}
	}

	return details
}

// extractOneToOneRelationValues 提取一对一关联的字段值（用于 Create/Delete）
func (uc *ChangeLogUseCase) extractOneToOneRelationValues(
	tracker interface{},
	config *ChangeLogFieldConfig,
	field reflect.Value,
	isOld bool,
) []*ChangeLogChangeDetail {
	var details []*ChangeLogChangeDetail

	val := field
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return details
		}
		val = val.Elem()
	}

	for _, child := range config.Children {
		childField := val.FieldByName(child.Name)
		if !childField.IsValid() {
			continue
		}

		value := uc.formatValue(tracker, child, childField.Interface())
		if value == "" {
			continue
		}

		fieldName := config.Relation + "." + child.Name
		label := child.Label
		if label == "" {
			label = config.Label
		}

		detail := &ChangeLogChangeDetail{
			FieldName:    fieldName,
			FieldLabel:   label,
			FieldType:    enum.ChangeLogDetailFieldTypeRelation,
			RelationType: enum.ChangeLogDetailRelationTypeOne,
		}
		if isOld {
			detail.OldValue = value
		} else {
			detail.NewValue = value
		}
		details = append(details, detail)
	}

	return details
}

// compareFields 比较两个结构体的字段
func (uc *ChangeLogUseCase) compareFields(
	tracker interface{},
	newRecord interface{},
	oldRecord interface{},
	fieldConfigs []*ChangeLogFieldConfig,
) []*ChangeLogChangeDetail {
	var details []*ChangeLogChangeDetail

	newVal := reflect.ValueOf(newRecord)
	oldVal := reflect.ValueOf(oldRecord)

	if newVal.Kind() == reflect.Ptr {
		newVal = newVal.Elem()
	}
	if oldVal.Kind() == reflect.Ptr {
		oldVal = oldVal.Elem()
	}

	for _, config := range fieldConfigs {
		// 组合字段: 使用 Children 和 CompositeFormat
		if config.IsComposite && len(config.Children) > 0 {
			oldValue := uc.formatCompositeValue(tracker, config, oldVal)
			newValue := uc.formatCompositeValue(tracker, config, newVal)
			if oldValue != newValue {
				details = append(details, &ChangeLogChangeDetail{
					FieldName:    config.Name,
					FieldLabel:   config.Label,
					FieldType:    enum.ChangeLogDetailFieldTypeComposite,
					RelationType: enum.ChangeLogDetailRelationTypeNone,
					OldValue:     oldValue,
					NewValue:     newValue,
				})
			}
			continue
		}

		// 普通字段: 比较单个字段值
		newField := newVal.FieldByName(config.Name)
		oldField := oldVal.FieldByName(config.Name)

		if !newField.IsValid() || !oldField.IsValid() {
			continue
		}

		if !reflect.DeepEqual(newField.Interface(), oldField.Interface()) {
			details = append(details, &ChangeLogChangeDetail{
				FieldName:    config.Name,
				FieldLabel:   config.Label,
				FieldType:    enum.ChangeLogDetailFieldTypeNormal,
				RelationType: enum.ChangeLogDetailRelationTypeNone,
				OldValue:     uc.formatValue(tracker, config, oldField.Interface()),
				NewValue:     uc.formatValue(tracker, config, newField.Interface()),
			})
		}
	}

	return details
}

// extractFieldValues 提取单条记录的字段值（用于 Create/Delete）
func (uc *ChangeLogUseCase) extractFieldValues(
	tracker interface{},
	record interface{},
	fieldConfigs []*ChangeLogFieldConfig,
	isOld bool,
) []*ChangeLogChangeDetail {
	var details []*ChangeLogChangeDetail

	val := reflect.ValueOf(record)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	for _, config := range fieldConfigs {
		// 组合字段
		if config.IsComposite && len(config.Children) > 0 {
			value := uc.formatCompositeValue(tracker, config, val)
			if value == "" {
				continue
			}
			detail := &ChangeLogChangeDetail{
				FieldName:    config.Name,
				FieldLabel:   config.Label,
				FieldType:    enum.ChangeLogDetailFieldTypeComposite,
				RelationType: enum.ChangeLogDetailRelationTypeNone,
			}
			if isOld {
				detail.OldValue = value
			} else {
				detail.NewValue = value
			}
			details = append(details, detail)
			continue
		}

		// 普通字段
		field := val.FieldByName(config.Name)
		if !field.IsValid() {
			continue
		}

		value := uc.formatValue(tracker, config, field.Interface())
		if value == "" {
			continue
		}

		detail := &ChangeLogChangeDetail{
			FieldName:    config.Name,
			FieldLabel:   config.Label,
			FieldType:    enum.ChangeLogDetailFieldTypeNormal,
			RelationType: enum.ChangeLogDetailRelationTypeNone,
		}
		if isOld {
			detail.OldValue = value
		} else {
			detail.NewValue = value
		}
		details = append(details, detail)
	}

	return details
}

// formatValue 格式化值，如果 Format 为 nil 则使用 FormatSimpleValue
func (uc *ChangeLogUseCase) formatValue(tracker interface{}, config *ChangeLogFieldConfig, value interface{}) string {
	if config.Format != nil {
		result, _ := config.Format(tracker, value)
		return result
	}
	return uc.FormatSimpleValue(tracker, value)
}

// formatCompositeValue 格式化组合字段值
func (uc *ChangeLogUseCase) formatCompositeValue(tracker interface{}, config *ChangeLogFieldConfig, record reflect.Value) string {
	if record.Kind() == reflect.Ptr {
		if record.IsNil() {
			return ""
		}
		record = record.Elem()
	}

	// 收集各子字段的格式化值（按顺序）
	values := make([]string, 0, len(config.Children))
	for _, child := range config.Children {
		field := record.FieldByName(child.Name)
		if !field.IsValid() {
			values = append(values, "")
			continue
		}
		values = append(values, uc.formatValue(tracker, child, field.Interface()))
	}

	// 使用 CompositeFunc 格式化
	if config.CompositeFunc != nil {
		return config.CompositeFunc(values)
	}

	// 默认用空格连接
	var parts []string
	for _, v := range values {
		if v != "" {
			parts = append(parts, v)
		}
	}
	return changeLogJoinStrings(parts, " ")
}

// formatManyRelation 格式化一对多/多对多关联为结构化 JSON
// 返回格式: [{"field1": "value1", "field2": "value2"}, ...]
func (uc *ChangeLogUseCase) formatManyRelation(tracker interface{}, config *ChangeLogFieldConfig, records interface{}) string {
	// 如果有自定义 Format，直接使用
	if config.Format != nil {
		result, _ := config.Format(tracker, records)
		return result
	}

	val := reflect.ValueOf(records)
	if val.Kind() != reflect.Slice {
		return "[]"
	}

	if val.Len() == 0 {
		return "[]"
	}

	// 如果没有 Children，使用 JSON 格式化
	if len(config.Children) == 0 {
		return uc.FormatJSONArray(tracker, records)
	}

	// 使用 Children 格式化每条记录为 map
	var items []map[string]string
	for i := 0; i < val.Len(); i++ {
		item := val.Index(i)
		if item.Kind() == reflect.Ptr {
			if item.IsNil() {
				continue
			}
			item = item.Elem()
		}

		record := make(map[string]string)
		for _, child := range config.Children {
			field := item.FieldByName(child.Name)
			if !field.IsValid() {
				continue
			}
			value := uc.formatValue(tracker, child, field.Interface())
			key := child.Name
			record[key] = value
		}
		items = append(items, record)
	}

	jsonBytes, err := json.Marshal(items)
	if err != nil {
		return "[]"
	}
	return string(jsonBytes)
}

func changeLogJoinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}

// FormatSimpleValue 格式化简单值的通用函数
func (uc *ChangeLogUseCase) FormatSimpleValue(tracker interface{}, value interface{}) string {
	if value == nil {
		return ""
	}
	return fmt.Sprintf("%v", value)
}

// FormatPointerValue 格式化指针值的通用函数
func (uc *ChangeLogUseCase) FormatPointerValue(tracker interface{}, value interface{}) string {
	val := reflect.ValueOf(value)
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return ""
		}
		return fmt.Sprintf("%v", val.Elem().Interface())
	}
	return fmt.Sprintf("%v", value)
}

// FormatJSONArray 格式化为 JSON 数组的通用函数
func (uc *ChangeLogUseCase) FormatJSONArray(tracker interface{}, records interface{}) string {
	if records == nil {
		return "[]"
	}

	val := reflect.ValueOf(records)
	if val.Kind() == reflect.Slice && val.Len() == 0 {
		return "[]"
	}

	data, err := json.Marshal(records)
	if err != nil {
		return "[]"
	}
	return string(data)
}

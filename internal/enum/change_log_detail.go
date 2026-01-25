package enum

import (
	"github.com/xelarion/go-layout/pkg/errs"
)

// ChangeLogDetailFieldType 变更详情字段类型
type ChangeLogDetailFieldType int8

const (
	ChangeLogDetailFieldTypeNormal    ChangeLogDetailFieldType = iota + 1 // 普通字段
	ChangeLogDetailFieldTypeComposite                                     // 组合字段
	ChangeLogDetailFieldTypeRelation                                      // 关联字段
)

var ChangeLogDetailFieldTypeSlug = map[ChangeLogDetailFieldType]string{
	ChangeLogDetailFieldTypeNormal:    "normal",
	ChangeLogDetailFieldTypeComposite: "composite",
	ChangeLogDetailFieldTypeRelation:  "relation",
}

var ChangeLogDetailFieldTypeText = map[ChangeLogDetailFieldType]string{
	ChangeLogDetailFieldTypeNormal:    "普通字段",
	ChangeLogDetailFieldTypeComposite: "组合字段",
	ChangeLogDetailFieldTypeRelation:  "关联字段",
}

func GetChangeLogDetailFieldType(slug string) (ChangeLogDetailFieldType, error) {
	for k, v := range ChangeLogDetailFieldTypeSlug {
		if v == slug {
			return k, nil
		}
	}
	return 0, errs.NewBusiness("无效的字段类型")
}

func (e ChangeLogDetailFieldType) Slug() string {
	if s, ok := ChangeLogDetailFieldTypeSlug[e]; ok {
		return s
	}
	return ""
}

func (e ChangeLogDetailFieldType) Text() string {
	if s, ok := ChangeLogDetailFieldTypeText[e]; ok {
		return s
	}
	return ""
}

// ChangeLogDetailRelationType 变更详情关联类型
type ChangeLogDetailRelationType int8

const (
	ChangeLogDetailRelationTypeNone ChangeLogDetailRelationType = iota // 非关联
	ChangeLogDetailRelationTypeOne                                     // 一对一
	ChangeLogDetailRelationTypeMany                                    // 一对多/多对多
)

var ChangeLogDetailRelationTypeSlug = map[ChangeLogDetailRelationType]string{
	ChangeLogDetailRelationTypeNone: "",
	ChangeLogDetailRelationTypeOne:  "one",
	ChangeLogDetailRelationTypeMany: "many",
}

var ChangeLogDetailRelationTypeText = map[ChangeLogDetailRelationType]string{
	ChangeLogDetailRelationTypeNone: "非关联",
	ChangeLogDetailRelationTypeOne:  "一对一",
	ChangeLogDetailRelationTypeMany: "一对多/多对多",
}

func GetChangeLogDetailRelationType(slug string) (ChangeLogDetailRelationType, error) {
	for k, v := range ChangeLogDetailRelationTypeSlug {
		if v == slug {
			return k, nil
		}
	}
	return 0, errs.NewBusiness("无效的关联类型")
}

func (e ChangeLogDetailRelationType) Slug() string {
	if s, ok := ChangeLogDetailRelationTypeSlug[e]; ok {
		return s
	}
	return ""
}

func (e ChangeLogDetailRelationType) Text() string {
	if s, ok := ChangeLogDetailRelationTypeText[e]; ok {
		return s
	}
	return ""
}

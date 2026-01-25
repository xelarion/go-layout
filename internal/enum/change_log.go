package enum

import (
	"github.com/xelarion/go-layout/pkg/errs"
)

// ChangeLogAction 操作类型
type ChangeLogAction int8

const (
	ChangeLogActionUnknown ChangeLogAction = iota // 未知
	ChangeLogActionCreate                         // 创建
	ChangeLogActionUpdate                         // 更新
	ChangeLogActionDelete                         // 删除
)

var ChangeLogActionSlug = map[ChangeLogAction]string{
	ChangeLogActionUnknown: "unknown",
	ChangeLogActionCreate:  "create",
	ChangeLogActionUpdate:  "update",
	ChangeLogActionDelete:  "delete",
}

var ChangeLogActionText = map[ChangeLogAction]string{
	ChangeLogActionUnknown: "未知",
	ChangeLogActionCreate:  "创建",
	ChangeLogActionUpdate:  "更新",
	ChangeLogActionDelete:  "删除",
}

func GetChangeLogAction(slug string) (ChangeLogAction, error) {
	for k, v := range ChangeLogActionSlug {
		if v == slug {
			return k, nil
		}
	}
	return 0, errs.NewBusiness("无效的操作类型")
}

func (e ChangeLogAction) Slug() string {
	if s, ok := ChangeLogActionSlug[e]; ok {
		return s
	}
	return ""
}

func (e ChangeLogAction) Text() string {
	if s, ok := ChangeLogActionText[e]; ok {
		return s
	}
	return ""
}

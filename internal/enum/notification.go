package enum

import (
	"github.com/xelarion/go-layout/pkg/errs"
)

// NotificationType 通知类型
type NotificationType int8

const (
	NotificationTypeAnnouncement     NotificationType = iota + 1 // 公告
	NotificationTypeProductionReport                             // 生产报工
)

var NotificationTypeSlug = map[NotificationType]string{
	NotificationTypeAnnouncement:     "announcement",
	NotificationTypeProductionReport: "production_report",
}

var NotificationTypeText = map[NotificationType]string{
	NotificationTypeAnnouncement:     "公告",
	NotificationTypeProductionReport: "生产报工",
}

func GetNotificationType(slug string) (NotificationType, error) {
	for k, v := range NotificationTypeSlug {
		if v == slug {
			return k, nil
		}
	}
	return 0, errs.NewBusiness("无效的通知类型")
}

type NotificationCategory int8

const (
	NotificationCategoryNotice  NotificationCategory = iota + 1 // 通知
	NotificationCategoryMessage                                 // 消息
)

var NotificationCategorySlug = map[NotificationCategory]string{
	NotificationCategoryNotice:  "notice",
	NotificationCategoryMessage: "message",
}

var NotificationCategoryText = map[NotificationCategory]string{
	NotificationCategoryNotice:  "通知",
	NotificationCategoryMessage: "消息",
}

func GetNotificationCategory(slug string) (NotificationCategory, error) {
	for k, v := range NotificationCategorySlug {
		if v == slug {
			return k, nil
		}
	}
	return 0, errs.NewBusiness("无效的通知类别")
}

// NotificationScope 范围
type NotificationScope int8

const (
	NotificationScopeAllUsers NotificationScope = iota + 1 // 全部用户
	NotificationScopeUser                                  // 单个用户
)

var NotificationScopeSlug = map[NotificationScope]string{
	NotificationScopeAllUsers: "all_users",
	NotificationScopeUser:     "user",
}

var NotificationScopeText = map[NotificationScope]string{
	NotificationScopeAllUsers: "全部用户",
	NotificationScopeUser:     "单个用户",
}

func GetNotificationScope(slug string) (NotificationScope, error) {
	for k, v := range NotificationScopeSlug {
		if v == slug {
			return k, nil
		}
	}
	return 0, errs.NewBusiness("无效的通知范围")
}

// TodoStatType 统计类型
// 用于标识不同类型的待办统计，如工单等
type TodoStatType int8

const (
	TodoStatTypeWorkOrder        TodoStatType = iota + 1 // 工单
	TodoStatTypeProductionReport                         // 生产报工
)

var TodoStatTypeText = map[TodoStatType]string{
	TodoStatTypeWorkOrder:        "工单任务",
	TodoStatTypeProductionReport: "生产报工",
}

var TodoStatTypeSlug = map[TodoStatType]string{
	TodoStatTypeWorkOrder:        "work_order",
	TodoStatTypeProductionReport: "production_report",
}

func GetTodoStatType(slug string) (TodoStatType, error) {
	for k, v := range TodoStatTypeSlug {
		if v == slug {
			return k, nil
		}
	}
	return 0, errs.NewBusiness("无效的待办统计类型")
}

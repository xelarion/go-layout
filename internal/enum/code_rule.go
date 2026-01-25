package enum

import (
	"github.com/xelarion/go-layout/pkg/errs"
)

// CodeRuleTimeRule 时间规则
type CodeRuleTimeRule int8

const (
	CodeRuleTimeRuleNone                         CodeRuleTimeRule = iota // 无
	CodeRuleTimeRuleYear                                                 // 年
	CodeRuleTimeRuleYearMonth                                            // 年月
	CodeRuleTimeRuleYearMonthDay                                         // 年月日
	CodeRuleTimeRuleYearMonthDayHour                                     // 年月日时
	CodeRuleTimeRuleYearMonthDayHourMinute                               // 年月日时分
	CodeRuleTimeRuleYearMonthDayHourMinuteSecond                         // 年月日时分秒
)

var CodeRuleTimeRuleSlug = map[CodeRuleTimeRule]string{
	CodeRuleTimeRuleNone:                         "none",
	CodeRuleTimeRuleYear:                         "YYYY",
	CodeRuleTimeRuleYearMonth:                    "YYYYMM",
	CodeRuleTimeRuleYearMonthDay:                 "YYYYMMDD",
	CodeRuleTimeRuleYearMonthDayHour:             "YYYYMMDDHH",
	CodeRuleTimeRuleYearMonthDayHourMinute:       "YYYYMMDDHHMM",
	CodeRuleTimeRuleYearMonthDayHourMinuteSecond: "YYYYMMDDHHMMSS",
}

var CodeRuleTimeRuleText = map[CodeRuleTimeRule]string{
	CodeRuleTimeRuleNone:                         "/",
	CodeRuleTimeRuleYear:                         "年-YYYY",
	CodeRuleTimeRuleYearMonth:                    "年月-YYYYMM",
	CodeRuleTimeRuleYearMonthDay:                 "年月日-YYYYMMDD",
	CodeRuleTimeRuleYearMonthDayHour:             "年月日时-YYYYMMDDHH",
	CodeRuleTimeRuleYearMonthDayHourMinute:       "年月日时分-YYYYMMDDHHMM",
	CodeRuleTimeRuleYearMonthDayHourMinuteSecond: "年月日时分秒-YYYYMMDDHHMMSS",
}

func GetCodeRuleTimeRule(slug string) (CodeRuleTimeRule, error) {
	for k, v := range CodeRuleTimeRuleSlug {
		if v == slug {
			return k, nil
		}
	}
	return 0, errs.NewBusiness("invalid CodeRuleTimeRule")
}

// CodeRuleGenerateType 生成类型
type CodeRuleGenerateType int8

const (
	CodeRuleGenerateTypeDaily  CodeRuleGenerateType = iota + 1 // 每天生成。每天从初始值开始计数。
	CodeRuleGenerateTypeSerial                                 // 递增，流水号一直增加。
)

var CodeRuleGenerateTypeSlug = map[CodeRuleGenerateType]string{
	CodeRuleGenerateTypeDaily:  "daily",
	CodeRuleGenerateTypeSerial: "serial",
}

var CodeRuleGenerateTypeText = map[CodeRuleGenerateType]string{
	CodeRuleGenerateTypeDaily:  "每天",
	CodeRuleGenerateTypeSerial: "递增",
}

func GetCodeRuleGenerateType(slug string) (CodeRuleGenerateType, error) {
	for k, v := range CodeRuleGenerateTypeSlug {
		if v == slug {
			return k, nil
		}
	}
	return 0, errs.NewBusiness("invalid CodeRuleGenerateType")
}

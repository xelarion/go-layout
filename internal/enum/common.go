package enum

import (
	"github.com/xelarion/go-layout/pkg/errs"
)

// DurationUnit 时间单位
type DurationUnit int8

const (
	DurationUnitDay     DurationUnit = iota + 1 // 天
	DurationUnitWeek                            // 周
	DurationUnitMonth                           // 月
	DurationUnitQuarter                         // 季
	DurationUnitYear                            // 年
)

var DurationUnitSlug = map[DurationUnit]string{
	DurationUnitDay:     "day",
	DurationUnitWeek:    "week",
	DurationUnitMonth:   "month",
	DurationUnitQuarter: "quarter",
	DurationUnitYear:    "year",
}

var DurationUnitText = map[DurationUnit]string{
	DurationUnitDay:     "天",
	DurationUnitWeek:    "周",
	DurationUnitMonth:   "月",
	DurationUnitQuarter: "季",
	DurationUnitYear:    "年",
}

func GetDurationUnit(slug string) (DurationUnit, error) {
	for k, v := range DurationUnitSlug {
		if v == slug {
			return k, nil
		}
	}
	return 0, errs.NewBusiness("无效的时间单位")
}

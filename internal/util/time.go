package util

import (
	"fmt"
	"time"

	"gorm.io/datatypes"
)

var TimeLocal = time.Local
var DefaultDateTimeFormat = "2006-01-02 15:04:05"
var DefaultDateFormat = "2006-01-02"

// DbMinTimePrecision 数据库字段最小操作精度
var DbMinTimePrecision = time.Microsecond

// DbMaxTimeNanoseconds 字段操作最大的纳秒数(超过此值999_999_499秒数会进位)
var DbMaxTimeNanoseconds = 999_999_000

// 未特别说明的方法，默认处理 TimeLocal 本地时区时间

// ===== FORMAT =====

// FormatBaseTime format string 基础方法
func FormatBaseTime(t time.Time, layout string) string {
	return t.In(TimeLocal).Format(layout)
}

// FormatDate return "2006-01-02" format string
func FormatDate(t time.Time) string {
	return FormatBaseTime(t, DefaultDateFormat)
}

// FormatDatePointer return "2006-01-02" format string
func FormatDatePointer(t *time.Time) string {
	if t == nil {
		return ""
	}

	return FormatBaseTime(*t, DefaultDateFormat)
}

// FormatDatatypesDate return "2006-01-02" format string for datatypes.Date
func FormatDatatypesDate(d datatypes.Date) string {
	t := time.Time(d)
	if t.IsZero() {
		return ""
	}
	return FormatDate(t)
}

// FormatDatatypesDatePointer return "2006-01-02" format string for *datatypes.Date
func FormatDatatypesDatePointer(d *datatypes.Date) string {
	if d == nil {
		return ""
	}
	t := time.Time(*d)
	if t.IsZero() {
		return ""
	}
	return FormatDate(t)
}

// FormatDateTime return "2006-01-02 15:04:05" format string
func FormatDateTime(t time.Time) string {
	return FormatBaseTime(t, DefaultDateTimeFormat)
}

// FormatDateTimePointer return "2006-01-02 15:04:05" format string
func FormatDateTimePointer(t *time.Time) string {
	if t == nil {
		return ""
	}

	return FormatBaseTime(*t, DefaultDateTimeFormat)
}

// FormatDuration return "x时x分x秒"
func FormatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	result := ""
	if hours > 0 {
		result += fmt.Sprintf("%d时", hours)
	}
	if minutes > 0 {
		result += fmt.Sprintf("%d分", minutes)
	}
	result += fmt.Sprintf("%d秒", seconds)

	return result
}

// ===== FORMAT END =====

// ===== PARSE =====

// ParseStringToBaseTime parse string to time 基础方法
func ParseStringToBaseTime(timeStr, layout string) time.Time {
	t, _ := time.ParseInLocation(layout, timeStr, TimeLocal)
	return t
}

// ParseStringToDate return "2021-01-01 00:00:00 +0800 CST"(local time)
func ParseStringToDate(timeStr string) time.Time {
	return ParseStringToBaseTime(timeStr, DefaultDateFormat)
}

// ParseStringToDateTime return "2021-01-01 01:01:01 +0800 CST"(local time)
func ParseStringToDateTime(timeStr string) time.Time {
	return ParseStringToBaseTime(timeStr, DefaultDateTimeFormat)
}

// ParseStringToDatatypesDate parse string to datatypes.Date
func ParseStringToDatatypesDate(dateStr string) datatypes.Date {
	t := ParseStringToDate(dateStr)
	return datatypes.Date(t)
}

// ===== PARSE END =====

// Year returns the year in which local time t occurs.
func Year(t time.Time) int {
	return t.In(TimeLocal).Year()
}

// Day returns the day of the month specified by local time of t.
func Day(t time.Time) int {
	return t.In(TimeLocal).Day()
}

// Hour returns the hour within the day specified by local time of t,
// in the range [0, 23].
func Hour(t time.Time) int {
	return t.In(TimeLocal).Hour()
}

// BeginningOfMinute beginning of minute(local time)
func BeginningOfMinute(t time.Time) time.Time {
	t = t.In(TimeLocal)
	year, month, day := t.Date()
	hour, min, _ := t.Clock()
	return time.Date(year, month, day, hour, min, 0, 0, TimeLocal)
}

// BeginningOfHour beginning of hour(local time)
func BeginningOfHour(t time.Time) time.Time {
	t = t.In(TimeLocal)
	year, month, day := t.Date()
	hour, _, _ := t.Clock()
	return time.Date(year, month, day, hour, 0, 0, 0, TimeLocal)
}

// BeginningOfDay beginning of day(local time)
func BeginningOfDay(t time.Time) time.Time {
	year, month, day := t.In(TimeLocal).Date()
	return time.Date(year, month, day, 0, 0, 0, 0, TimeLocal)
}

// BeginningOfMonth beginning of month(local time)
func BeginningOfMonth(t time.Time) time.Time {
	year, month, _ := t.In(TimeLocal).Date()
	return time.Date(year, month, 1, 0, 0, 0, 0, TimeLocal)
}

// BeginningOfYear beginning of year(local time)
func BeginningOfYear(t time.Time) time.Time {
	year, _, _ := t.In(TimeLocal).Date()
	return time.Date(year, 1, 1, 0, 0, 0, 0, TimeLocal)
}

// EndOfMinute end of minute(local time)
func EndOfMinute(t time.Time) time.Time {
	t = t.In(TimeLocal)
	year, month, day := t.Date()
	hour, min, _ := t.Clock()
	return time.Date(year, month, day, hour, min, 59, DbMaxTimeNanoseconds, TimeLocal)
}

// EndOfHour end of hour(local time)
func EndOfHour(t time.Time) time.Time {
	localTime := t.In(TimeLocal)
	year, month, day := localTime.Date()
	return time.Date(year, month, day, localTime.Hour(), 59, 59, DbMaxTimeNanoseconds, TimeLocal)
}

// EndOfDay end of day(local time)
func EndOfDay(t time.Time) time.Time {
	year, month, day := t.In(TimeLocal).Date()
	return time.Date(year, month, day, 23, 59, 59, DbMaxTimeNanoseconds, TimeLocal)
}

// EndOfMonth end of month(local time)
func EndOfMonth(t time.Time) time.Time {
	// gorm 1.0 不支持自动转化成数据库精度(应该是直接传给数据库), 2.0 没有问题
	// error: gorm update 2020-02-04 15:59:59.999999999 +0000 UTC to 2020-02-04 16:00:00 +0000 UTC
	return BeginningOfMonth(t).AddDate(0, 1, 0).Add(-DbMinTimePrecision)
}

// EndOfYear end of year(local time)
func EndOfYear(t time.Time) time.Time {
	return BeginningOfYear(t).AddDate(1, 0, 0).Add(-DbMinTimePrecision)
}

// LastDayOfMonth returns the last day of the month specified by local time of t.
func LastDayOfMonth(t time.Time) int {
	return EndOfMonth(t).Day()
}

// AddDate 时间增减
// 类似于ruby中的时间增减，和 time.AddDate 不同
// 如：
// loc, _ := time.LoadLocation("Asia/Shanghai")
// t := time.Date(2010, 3, 31, 12, 0, 0, 0, loc)
// utils.AddDate(t, 0, 1)
// => 2010-04-30 12:00:00 +0800 CST
// 不会因为4月没有31号，而变成5月1号
func AddDate(t time.Time, years int, months int) time.Time {
	year, month, day := t.Date()
	hour, min, sec := t.Clock()

	// firstDayOfMonthAfterAddDate: years 年，months 月后的 那个月份的1号
	firstDayOfMonthAfterAddDate := time.Date(year+years, month+time.Month(months), 1,
		hour, min, sec, t.Nanosecond(), t.Location())
	// firstDayOfMonthAfterAddDate 月份的最后一天
	lastDay := LastDayOfMonth(firstDayOfMonthAfterAddDate)

	// 如果 t 的天 > lastDay，则设置为lastDay
	// 如：t 为 2020-03-31 12:00:00 +0800，增加1个月，为4月31号
	// 但是4月没有31号，则设置为4月最后一天lastDay（30号）
	if day > lastDay {
		day = lastDay
	}

	return time.Date(year+years, month+time.Month(months), day,
		hour, min, sec, t.Nanosecond(), t.Location())
}

// DBTimeNow database locate now time
func DBTimeNow() time.Time {
	return time.Now().UTC()
}

// DBTimeNowPointer database locate pointer of now time
func DBTimeNowPointer() *time.Time {
	t := time.Now().UTC()
	return &t
}

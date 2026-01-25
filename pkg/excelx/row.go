// Package excelx provides utilities for Excel data processing.
package excelx

import (
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Common validation patterns
var (
	phonePattern = regexp.MustCompile(`^1[3-9]\d{9}$`)
	emailPattern = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
)

// Row wraps a row of Excel data for convenient access and validation.
//
// Usage:
//
//	row := excelx.NewRow(data)
//	code := row.Col(0).Validate("客户编号", excelx.Required(), excelx.MaxLen(100)).String()
//	age := row.Col(1).Validate("年龄", excelx.MinInt(0), excelx.MaxInt(150)).Int()
//	email := row.Col(2).Validate("邮箱", excelx.Email()).String()
//	phone := row.Col(3).Validate("手机号", excelx.Phone()).String()
//	enabled := row.Col(4).Validate("状态", excelx.In(map[string]any{"启用": true, "禁用": false})).Bool()
//	customerType := excelx.Val(row.Col(5).Validate("客户类型", excelx.In(map[string]any{"企业": TypeA, "个人": TypeB})), TypeA)
//	if row.HasError() {
//	    return row.Error()
//	}
type Row struct {
	data   []string
	errors []string
}

// NewRow creates a new Row wrapper.
func NewRow(data []string) *Row {
	return &Row{data: data}
}

// Col returns a column accessor for the given index.
func (r *Row) Col(index int) *ColAccessor {
	var value string
	if index >= 0 && index < len(r.data) {
		value = strings.TrimSpace(r.data[index])
	}
	return &ColAccessor{
		row:   r,
		value: value,
		index: index,
	}
}

// HasError returns true if any validation errors occurred.
func (r *Row) HasError() bool {
	return len(r.errors) > 0
}

// Error returns the first validation error message.
func (r *Row) Error() string {
	if len(r.errors) > 0 {
		return r.errors[0]
	}
	return ""
}

// Errors returns all validation error messages.
func (r *Row) Errors() []string {
	return r.errors
}

// addError adds a validation error.
func (r *Row) addError(msg string) {
	r.errors = append(r.errors, msg)
}

// ColAccessor provides fluent API for column value access and validation.
type ColAccessor struct {
	row       *Row
	value     string
	index     int
	hasError  bool
	mappedVal any // stores the actual value from In(map[string]any{...}) for type conversion
}

// Validator is a validation function that checks a value.
// It receives the ColAccessor to allow setting mapped values.
type Validator func(c *ColAccessor, fieldName string) (ok bool, errMsg string)

// Validate applies validators to the column value with field name.
func (c *ColAccessor) Validate(fieldName string, validators ...Validator) *ColAccessor {
	if c.hasError {
		return c
	}
	for _, v := range validators {
		if ok, errMsg := v(c, fieldName); !ok {
			c.hasError = true
			c.row.addError(errMsg)
			return c
		}
	}
	return c
}

// --- Validators ---

// Required returns a validator that checks if value is not empty.
func Required() Validator {
	return func(c *ColAccessor, fieldName string) (bool, string) {
		if c.value == "" {
			return false, fieldName + "不能为空"
		}
		return true, ""
	}
}

// MaxLen returns a validator that checks maximum character length.
func MaxLen(max int) Validator {
	return func(c *ColAccessor, fieldName string) (bool, string) {
		if c.value == "" {
			return true, ""
		}
		if utf8.RuneCountInString(c.value) > max {
			return false, fieldName + "长度不能超过" + strconv.Itoa(max) + "个字符"
		}
		return true, ""
	}
}

// MinLen returns a validator that checks minimum character length.
func MinLen(min int) Validator {
	return func(c *ColAccessor, fieldName string) (bool, string) {
		if c.value == "" {
			return true, ""
		}
		if utf8.RuneCountInString(c.value) < min {
			return false, fieldName + "长度不能少于" + strconv.Itoa(min) + "个字符"
		}
		return true, ""
	}
}

// Phone returns a validator for Chinese mainland mobile phone format (11 digits starting with 1).
func Phone() Validator {
	return func(c *ColAccessor, fieldName string) (bool, string) {
		if c.value == "" {
			return true, ""
		}
		if !phonePattern.MatchString(c.value) {
			return false, fieldName + "必须为11位手机号"
		}
		return true, ""
	}
}

// Email returns a validator for email format.
func Email() Validator {
	return func(c *ColAccessor, fieldName string) (bool, string) {
		if c.value == "" {
			return true, ""
		}
		if !emailPattern.MatchString(c.value) {
			return false, fieldName + "格式不正确"
		}
		return true, ""
	}
}

// Match returns a validator that checks value against a regex pattern.
func Match(pattern *regexp.Regexp, errSuffix string) Validator {
	return func(c *ColAccessor, fieldName string) (bool, string) {
		if c.value == "" {
			return true, ""
		}
		if !pattern.MatchString(c.value) {
			return false, fieldName + errSuffix
		}
		return true, ""
	}
}

// In returns a validator that checks if value is one of the allowed values.
// If mapping is provided (map[string]any), it also stores the mapped value for later retrieval.
// Usage:
//   - In("a", "b", "c") - simple string validation
//   - In(map[string]any{"启用": true, "禁用": false}) - with value mapping
func In(args ...any) Validator {
	if len(args) == 0 {
		panic("In requires at least one argument")
	}

	// Check if first arg is a map
	if mapping, ok := args[0].(map[string]any); ok {
		texts := make([]string, 0, len(mapping))
		for k := range mapping {
			texts = append(texts, k)
		}
		return func(c *ColAccessor, fieldName string) (bool, string) {
			if c.value == "" {
				return true, ""
			}
			if v, ok := mapping[c.value]; ok {
				c.mappedVal = v
				return true, ""
			}
			return false, fieldName + "必须是以下值之一: " + strings.Join(texts, ", ")
		}
	}

	// Otherwise treat as string list
	allowed := make([]string, len(args))
	for i, arg := range args {
		s, ok := arg.(string)
		if !ok {
			panic("In arguments must be strings or a single map[string]any")
		}
		allowed[i] = s
	}
	return func(c *ColAccessor, fieldName string) (bool, string) {
		if c.value == "" {
			return true, ""
		}
		for _, v := range allowed {
			if c.value == v {
				return true, ""
			}
		}
		return false, fieldName + "必须是以下值之一: " + strings.Join(allowed, ", ")
	}
}

// MinInt returns a validator for minimum int value.
func MinInt(min int) Validator {
	return func(c *ColAccessor, fieldName string) (bool, string) {
		if c.value == "" {
			return true, ""
		}
		v, err := strconv.Atoi(c.value)
		if err != nil {
			return false, fieldName + "必须为整数"
		}
		if v < min {
			return false, fieldName + "不能小于" + strconv.Itoa(min)
		}
		return true, ""
	}
}

// MaxInt returns a validator for maximum int value.
func MaxInt(max int) Validator {
	return func(c *ColAccessor, fieldName string) (bool, string) {
		if c.value == "" {
			return true, ""
		}
		v, err := strconv.Atoi(c.value)
		if err != nil {
			return false, fieldName + "必须为整数"
		}
		if v > max {
			return false, fieldName + "不能大于" + strconv.Itoa(max)
		}
		return true, ""
	}
}

// MinFloat returns a validator for minimum float64 value.
func MinFloat(min float64) Validator {
	return func(c *ColAccessor, fieldName string) (bool, string) {
		if c.value == "" {
			return true, ""
		}
		v, err := strconv.ParseFloat(c.value, 64)
		if err != nil {
			return false, fieldName + "必须为数字"
		}
		if v < min {
			return false, fieldName + "不能小于" + strconv.FormatFloat(min, 'f', -1, 64)
		}
		return true, ""
	}
}

// MaxFloat returns a validator for maximum float64 value.
func MaxFloat(max float64) Validator {
	return func(c *ColAccessor, fieldName string) (bool, string) {
		if c.value == "" {
			return true, ""
		}
		v, err := strconv.ParseFloat(c.value, 64)
		if err != nil {
			return false, fieldName + "必须为数字"
		}
		if v > max {
			return false, fieldName + "不能大于" + strconv.FormatFloat(max, 'f', -1, 64)
		}
		return true, ""
	}
}

// Positive returns a validator that checks if numeric value is positive (> 0).
func Positive() Validator {
	return func(c *ColAccessor, fieldName string) (bool, string) {
		if c.value == "" {
			return true, ""
		}
		v, err := strconv.ParseFloat(c.value, 64)
		if err != nil {
			return false, fieldName + "必须为数字"
		}
		if v <= 0 {
			return false, fieldName + "必须为正数"
		}
		return true, ""
	}
}

// NonNegative returns a validator that checks if numeric value is non-negative (>= 0).
func NonNegative() Validator {
	return func(c *ColAccessor, fieldName string) (bool, string) {
		if c.value == "" {
			return true, ""
		}
		v, err := strconv.ParseFloat(c.value, 64)
		if err != nil {
			return false, fieldName + "必须为数字"
		}
		if v < 0 {
			return false, fieldName + "不能为负数"
		}
		return true, ""
	}
}

// IsInt returns a validator that checks if value is a valid integer.
func IsInt() Validator {
	return func(c *ColAccessor, fieldName string) (bool, string) {
		if c.value == "" {
			return true, ""
		}
		if _, err := strconv.Atoi(c.value); err != nil {
			return false, fieldName + "必须为整数"
		}
		return true, ""
	}
}

// IsFloat returns a validator that checks if value is a valid number.
func IsFloat() Validator {
	return func(c *ColAccessor, fieldName string) (bool, string) {
		if c.value == "" {
			return true, ""
		}
		if _, err := strconv.ParseFloat(c.value, 64); err != nil {
			return false, fieldName + "必须为数字"
		}
		return true, ""
	}
}

// Custom returns a custom validator.
func Custom(fn func(value, fieldName string) (bool, string)) Validator {
	return func(c *ColAccessor, fieldName string) (bool, string) {
		if c.value == "" {
			return true, ""
		}
		return fn(c.value, fieldName)
	}
}

// --- Value Getters ---

// String returns the trimmed string value.
func (c *ColAccessor) String() string {
	return c.value
}

// Int parses and returns the value as int. Returns 0 if empty or parse fails.
func (c *ColAccessor) Int() int {
	if c.value == "" {
		return 0
	}
	v, _ := strconv.Atoi(c.value)
	return v
}

// Int64 parses and returns the value as int64. Returns 0 if empty or parse fails.
func (c *ColAccessor) Int64() int64 {
	if c.value == "" {
		return 0
	}
	v, _ := strconv.ParseInt(c.value, 10, 64)
	return v
}

// Uint parses and returns the value as uint. Returns 0 if empty or parse fails.
func (c *ColAccessor) Uint() uint {
	if c.value == "" {
		return 0
	}
	v, _ := strconv.ParseUint(c.value, 10, 64)
	return uint(v)
}

// Float64 parses and returns the value as float64. Returns 0 if empty or parse fails.
func (c *ColAccessor) Float64() float64 {
	if c.value == "" {
		return 0
	}
	v, _ := strconv.ParseFloat(c.value, 64)
	return v
}

// Bool returns the mapped bool value if available, otherwise parses the string value.
// For mapped values: returns the bool from In(map[string]any{...})
// For string values: accepts "true", "t", "1", "是", "yes", "on" as true
func (c *ColAccessor) Bool() bool {
	if c.mappedVal != nil {
		if v, ok := c.mappedVal.(bool); ok {
			return v
		}
	}
	switch strings.ToLower(c.value) {
	case "true", "t", "1", "是", "yes", "on":
		return true
	default:
		return false
	}
}

// Val returns the mapped value with type assertion.
// The zero parameter is used to infer the return type.
// Usage: customerType := Val(row.Col(2).Validate("类型", In(mapping)), enum.CustomerType(0))
func Val[T any](c *ColAccessor, zero T) T {
	if c.mappedVal != nil {
		if v, ok := c.mappedVal.(T); ok {
			return v
		}
	}
	return zero
}

// IsEmpty returns true if the value is empty.
func (c *ColAccessor) IsEmpty() bool {
	return c.value == ""
}

// IsNotEmpty returns true if the value is not empty.
func (c *ColAccessor) IsNotEmpty() bool {
	return c.value != ""
}

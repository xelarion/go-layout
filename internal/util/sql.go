package util

import "strings"

func EscapeFullLike(key string) string {
	return escape("%", key, "%")
}

func EscapeLeftLike(key string) string {
	return escape("%", key, "")
}

func EscapeRightLike(key string) string {
	return escape("", key, "%")
}

func Escape(key string) string {
	return escape("", key, "")
}

func escape(prefix, key, suffix string) string {
	var buf strings.Builder
	buf.Grow(len(key) + 10 + len(prefix) + len(suffix))
	buf.WriteString(prefix)

	for _, c := range key {
		switch c {
		case '%', '_', '\\': // 特殊字符需要转义
			buf.WriteRune('\\')
		}
		buf.WriteRune(c)
	}
	buf.WriteString(suffix)
	return buf.String()
}

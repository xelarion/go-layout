package util

import "strings"

func EscapeLike(key string) string {
	var buf strings.Builder
	buf.Grow(len(key) + 10)
	buf.WriteRune('%')

	for _, c := range key {
		switch c {
		case '%', '_', '\\': // 特殊字符需要转义
			buf.WriteRune('\\')
		}
		buf.WriteRune(c)
	}
	buf.WriteRune('%')
	return buf.String()
}

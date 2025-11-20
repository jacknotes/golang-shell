package utils

import (
	"strings"
)

// TrimSpaces 去除字符串两端的空白字符
func TrimSpaces(s string) string {
	return strings.TrimSpace(s)
}

// ToUpper 将字符串转换为大写
func ToUpper(s string) string {
	return strings.ToUpper(s)
}

// ToLower 将字符串转换为小写
func ToLower(s string) string {
	return strings.ToLower(s)
}

// Contains 判断字符串是否包含子串
func Contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

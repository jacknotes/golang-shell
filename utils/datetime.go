package utils

import (
	"time"
)

// FormatTime 格式化时间
func FormatTime(t time.Time, layout string) string {
	return t.Format(layout)
}

// ParseTime 解析时间字符串
func ParseTime(value string, layout string) (time.Time, error) {
	return time.Parse(layout, value)
}

// Now 获取当前时间
func Now() time.Time {
	return time.Now()
}

// Unix 获取当前时间的Unix时间戳
func Unix() int64 {
	return time.Now().Unix()
}

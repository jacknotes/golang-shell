package utils

import (
	"errors"
	"fmt"
)

// WrapError 包装错误信息
func WrapError(err error, msg string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", msg, err)
}

// IsError 检查错误是否为指定类型
func IsError(err error, target error) bool {
	return errors.Is(err, target)
}

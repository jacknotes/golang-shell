package utils

import (
	"encoding/json"
)

// Marshal 将对象转换为JSON字符串
func Marshal(v interface{}) (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Unmarshal 将JSON字符串转换为对象
func Unmarshal(data string, v interface{}) error {
	return json.Unmarshal([]byte(data), v)
}

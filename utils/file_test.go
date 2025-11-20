package utils

import (
	"io/ioutil"
	"os"
	"testing"
)

// TestWriteFile 测试 WriteFile 函数
func TestWriteFile(t *testing.T) {
	testFile := "testfile.txt"
	testData := []byte("test data")
	testPerm := os.FileMode(0644)

	err := WriteFile(testFile, testData, testPerm)
	if err != nil {
		t.Errorf("WriteFile failed: %v", err)
	} else {
		t.Log("WriteFile succeeded")
	}

	content, err := ioutil.ReadFile(testFile)
	if err != nil {
		t.Errorf("ReadFile failed: %v", err)
	}

	if string(content) != string(testData) {
		t.Errorf("File content mismatch: expected %q, got %q", testData, content)
	}

	// 删除测试文件
	err = os.Remove(testFile)
	if err != nil {
		t.Errorf("Remove file failed: %v", err)
	}
}

// TestFileExists 测试 FileExists 函数
func TestFileExists(t *testing.T) {
	testFile := "testfile.txt"
	testData := []byte("test data")
	testPerm := os.FileMode(0644)

	// 创建测试文件
	err := WriteFile(testFile, testData, testPerm)
	if err != nil {
		t.Errorf("WriteFile failed: %v", err)
	}

	if !FileExists(testFile) {
		t.Errorf("FileExists returned false for existing file")
	} else {
		t.Log("FileExists returned true for existing file")
	}

	// 删除测试文件
	err = os.Remove(testFile)
	if err != nil {
		t.Errorf("Remove file failed: %v", err)
	}

	if FileExists(testFile) {
		t.Errorf("FileExists returned true for non-existing file")
	}
}

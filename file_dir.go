package goutils

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
)

// DestExists 检查目标对象（文件、目录、符号链接，等）是否存在。
func DestExists(pth string) bool {
	_, err := os.Stat(pth)
	if err != nil {
		return os.IsExist(err)
	}

	return true
}

// CreateDirIfNotExist creates target directory with mode `0700` if it
// does not exist.
func CreateDirIfNotExist(pth string, mode os.FileMode) error {
	info, err := os.Stat(pth)

	if err != nil {
		if os.IsNotExist(err) {
			// Destination doesn't exist. Create it.
			err := os.MkdirAll(pth, mode)

			if err != nil {
				return fmt.Errorf("failed in creating directory %s. error=%v", pth, err)
			}
		} else {
			return err
		}
	} else {
		// 目标路径存在，但不是目录。
		if !info.IsDir() {
			return fmt.Errorf("%s exists, but not a directory", pth)
		}
	}

	return nil
}

// CreateFileIfNotExist creates target file with mode `0700` if it doesn't exist.
func CreateFileIfNotExist(pth string, content []byte, mode os.FileMode) error {
	info, err := os.Stat(pth)
	if err == nil {
		if info.IsDir() {
			return fmt.Errorf("%s is a directory (which should be a regular file)", pth)
		}

		return nil
	}

	if os.IsNotExist(err) {
		// Check and create (if not exist) parent directory
		dir := filepath.Dir(pth)
		if err := CreateDirIfNotExist(dir, 0700); err != nil {
			return err
		}

		// 创建文件
		if err := os.WriteFile(pth, content, mode); err != nil {
			return fmt.Errorf("failed in creating file %s: %v", pth, err)
		}

		return nil
	} else {
		// 其它错误
		return fmt.Errorf("failed in checking stat of file %s: %v", pth, err)
	}
}

// ReadFullFileContent 读取指定文件的所有内容，并去除首尾的空白字符。
func ReadFullFileContent(pth string) (content []byte, err error) {
	content, err = os.ReadFile(pth)
	if err != nil {
		return
	}

	content = bytes.TrimSpace(content)

	return
}

// ReadFullFileContentInString 读取指定文件的所有内容，并去除首尾的空白字符，以 string 类型返回文件内容。
func ReadFullFileContentInString(pth string) (content string, err error) {
	b, err := os.ReadFile(pth)
	if err != nil {
		return
	}

	b = bytes.TrimSpace(b)

	return string(b), nil
}

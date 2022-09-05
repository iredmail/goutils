package goutils

import (
	"fmt"
	"os"
	"path/filepath"
)

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

		return os.ErrExist
	} else if os.IsNotExist(err) {
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

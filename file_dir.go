package goutils

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"syscall"
)

type FileStat struct {
	Exists    bool
	IsLink    bool // symbol link
	IsRegular bool // regular file
	IsDir     bool // directory
	Owner     string
	Group     string
	Mode      os.FileMode
	Uid       uint32
	Gid       uint32
}

func (fs *FileStat) String() string {
	return fmt.Sprintf(
		"Exists: %v, IsLink: %v, IsRegular: %v, IsDir: %v",
		fs.Exists, fs.IsLink, fs.IsRegular, fs.IsDir,
	)
}

func GetFileStat(pth string) (*FileStat, error) {
	fs := new(FileStat)

	stat, err := os.Lstat(pth)
	if err != nil {
		// 不能用 os.IsNotExist(err) 来判断文件是否存在
		if e, ok := err.(*os.PathError); ok {
			// no such file or directory
			if errors.Is(e.Err, syscall.ENOENT) {
				return fs, nil
			}
		}

		return fs, fmt.Errorf("failed in checking stat of %s: %v", pth, err)
	}

	fs.Exists = true
	fs.Mode = stat.Mode()

	// Get uid / gid and owner / group names
	ss := stat.Sys().(*syscall.Stat_t)
	fs.Uid = ss.Uid
	fs.Gid = ss.Gid

	usr, err := user.LookupId(fmt.Sprintf("%d", fs.Uid))
	if err == nil {
		fs.Owner = usr.Username
	}

	group, err := user.LookupGroupId(fmt.Sprintf("%d", fs.Gid))
	if err == nil {
		fs.Group = group.Name
	}

	if stat.IsDir() {
		fs.IsDir = true

		return fs, nil
	}

	if stat.Mode()&os.ModeSymlink == os.ModeSymlink {
		fs.IsLink = true

		return fs, nil
	}

	fs.IsRegular = true

	return fs, nil
}

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
	if !DestExists(pth) {
		err = fmt.Errorf("file %s does not exist", pth)

		return
	}

	b, err := os.ReadFile(pth)
	if err != nil {
		return
	}

	b = bytes.TrimSpace(b)

	return string(b), nil
}

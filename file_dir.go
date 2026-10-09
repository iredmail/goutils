package goutils

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"syscall"
)

type FileStat struct {
	Name      string
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

func GetFileStat(pth string) (FileStat, error) {
	fs := FileStat{}

	stat, err := os.Lstat(pth)
	if err != nil {
		if os.IsNotExist(err) {
			return fs, nil
		}

		return fs, fmt.Errorf("failed in checking stat of %s: %w", pth, err)
	}

	fs.Exists = true
	fs.Name = stat.Name()
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

	return err == nil
}

// CreateDirIfNotExist creates target directory with mode `0700` if it
// does not exist.
func CreateDirIfNotExist(pth string, mode os.FileMode) (err error) {
	var info os.FileInfo

	info, err = os.Stat(pth)

	if err != nil {
		if os.IsNotExist(err) {
			// Destination doesn't exist. Create it.
			err = os.MkdirAll(pth, mode)

			if err != nil {
				return fmt.Errorf("failed in creating directory %s. error=%v", pth, err)
			}
		} else {
			return
		}
	} else {
		// 目标路径存在，但不是目录。
		if !info.IsDir() {
			return fmt.Errorf("%s exists, but not a directory", pth)
		}
	}

	return
}

func CreateDirWithOGMIfNotExist(pth string, mode os.FileMode, ownerUid, ownerGid int) (err error) {
	err = CreateDirIfNotExist(pth, mode)
	if err != nil {
		return
	}

	return os.Chown(pth, ownerUid, ownerGid)
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
		// Check and create (if not exist) parent directory with permission 0755.
		dir := filepath.Dir(pth)
		if err2 := CreateDirIfNotExist(dir, 0755); err2 != nil {
			return err2
		}

		// 创建文件
		if err2 := os.WriteFile(pth, content, mode); err2 != nil {
			return fmt.Errorf("failed in creating file %s: %v", pth, err2)
		}

		return nil
	}

	// 其它错误
	return fmt.Errorf("failed in checking stat of file %s: %v", pth, err)
}

// WriteFileAtomic 以「先写临时文件、再原子替换」的方式写入文件。
//
// 为什么不直接用 os.WriteFile / O_TRUNC 原地覆盖：
// 这类写法会先把目标文件截断为 0 字节，再写入新内容。
// 若在两步之间进程被中断（断电、OOM、被 kill、容器被驱逐），
// 目标文件就会停留在**空**或**半截**状态，造成不可逆的数据丢失。
// 而被写入的往往是关键配置或状态文件（/etc/fstab、/etc/hosts、
// /etc/sysctl.conf、crontab 等），一旦丢失会直接影响系统启动或服务可用性。
//
// 本函数改为：
//  1. 在目标文件所在目录创建临时文件（同目录，保证可原子 rename）；
//  2. 写入内容并 fsync，确保数据真正落盘；
//  3. 用 os.Rename 原子替换目标文件。
//
// rename 在同一文件系统内是原子操作，因此观察者要么看到旧内容、
// 要么看到新内容，不会看到中间状态；失败时原文件保持不变。
//
// perm 用于新建文件时的权限。注意：若目标文件已存在，
// 其原有权限会被保留（通过先把权限复制到临时文件实现），
// 因此显式传入的 perm 只对「文件原本不存在」的情况生效。
func WriteFileAtomic(pth string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(pth)

	// 若目标文件已存在，沿用它的权限，避免一次改写意外改变文件权限。
	if info, err := os.Stat(pth); err == nil {
		perm = info.Mode().Perm()
	}

	tmp, err := os.CreateTemp(dir, "."+filepath.Base(pth)+".tmp-*")
	if err != nil {
		return fmt.Errorf("failed in creating temporary file in %s: %w", dir, err)
	}

	tmpPath := tmp.Name()

	// 清理函数：任何提前返回都要移除临时文件，避免残留。
	// 成功 rename 之后该路径已不存在，重复删除是无害的。
	defer func() {
		_ = os.Remove(tmpPath)
	}()

	if err = tmp.Chmod(perm); err != nil {
		_ = tmp.Close()

		return fmt.Errorf("failed in setting permission of %s: %w", tmpPath, err)
	}

	if _, err = tmp.Write(data); err != nil {
		_ = tmp.Close()

		return fmt.Errorf("failed in writing to %s: %w", tmpPath, err)
	}

	// fsync：确保内容在 rename 之前已真正写入磁盘。
	// 否则断电时可能出现「rename 已生效但数据未落盘」的空文件。
	if err = tmp.Sync(); err != nil {
		_ = tmp.Close()

		return fmt.Errorf("failed in syncing %s: %w", tmpPath, err)
	}

	if err = tmp.Close(); err != nil {
		return fmt.Errorf("failed in closing %s: %w", tmpPath, err)
	}

	if err = os.Rename(tmpPath, pth); err != nil {
		return fmt.Errorf("failed in replacing %s: %w", pth, err)
	}

	return nil
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

// MoveDir moves a directory from src to dst.
// It first attempts an atomic rename. If that fails due to cross-device
// boundary issues (EXDEV), it falls back to a manual copy and delete.
func MoveDir(src, dst string) error {
	err := os.Rename(src, dst)
	if err == nil {
		return nil
	}

	// Check if the error is due to moving across different partitions/filesystems.
	// We check for EXDEV (cross-device link error).
	if linkErr, ok := errors.AsType[*os.LinkError](err); ok {
		if errors.Is(linkErr.Err, syscall.EXDEV) {
			// Fallback: Copy the directory and then remove the source
			if err = copyDir(src, dst); err != nil {
				return err
			}

			return os.RemoveAll(src)
		}
	}

	return err
}

// copyDir recursively copies a directory tree.
func copyDir(src, dst string) error {
	stat, err := os.Stat(src)
	if err != nil {
		return err
	}

	sys := stat.Sys().(*syscall.Stat_t)

	// Create the destination directory with the same permissions
	if err = os.MkdirAll(dst, stat.Mode()); err != nil {
		return err
	}

	// Set the owner and group of the destination directory
	if err = os.Chown(dst, int(sys.Uid), int(sys.Gid)); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err = copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err = copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// copyFile copies a single file from src to dst.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err = io.Copy(out, in); err != nil {
		return err
	}

	// Sync file content to disk
	if err = out.Sync(); err != nil {
		return err
	}

	// Copy file permissions and ownership
	stat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err = os.Chmod(dst, stat.Mode()); err != nil {
		return err
	}

	sys := stat.Sys().(*syscall.Stat_t)

	return os.Chown(dst, int(sys.Uid), int(sys.Gid))
}

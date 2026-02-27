package goutils

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"syscall"
)

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
	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	// Create the destination directory with the same permissions
	if err = os.MkdirAll(dst, info.Mode()); err != nil {
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

	// Sync to disk and copy file permissions
	info, err := os.Stat(src)
	if err == nil {
		return os.Chmod(dst, info.Mode())
	}

	return nil
}

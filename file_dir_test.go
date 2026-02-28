package goutils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDestExists(t *testing.T) {
	pth1 := filepath.Join(os.TempDir(), "1.txt")
	pth2 := filepath.Join(os.TempDir(), "2.txt")

	// Make sure files are absent.
	_ = os.Remove(pth1)
	_ = os.Remove(pth2)

	assert.False(t, DestExists(pth1))

	err := os.WriteFile(pth2, []byte("test"), 0700)
	assert.Nil(t, err)
	assert.True(t, DestExists(pth2))

	err = os.Remove(pth2)
	assert.Nil(t, err)
	assert.False(t, DestExists(pth2))
}

func TestReadFullFileContent(t *testing.T) {
	var content []byte
	var s string
	var err error

	pth := filepath.Join(os.TempDir(), "1.txt")
	err = os.WriteFile(pth, []byte("test"), 0700)
	assert.Nil(t, err)

	content, err = ReadFullFileContent(pth)
	assert.Nil(t, err)
	assert.Equal(t, []byte("test"), content)

	s, err = ReadFullFileContentInString(pth)
	assert.Nil(t, err)
	assert.Equal(t, "test", s)

	err = os.WriteFile(pth, []byte("\n\ttest\r\n"), 0700)
	assert.Nil(t, err)
	assert.Equal(t, []byte("test"), content)

	s, err = ReadFullFileContentInString(pth)
	assert.Nil(t, err)
	assert.Equal(t, "test", s)

	_ = os.Remove(pth)
	assert.False(t, DestExists(pth))
}

func TestMoveDir(t *testing.T) {
	tmpDir := os.TempDir()

	t.Run("basic move within same filesystem", func(t *testing.T) {
		src := filepath.Join(tmpDir, "test_move_src")
		dst := filepath.Join(tmpDir, "test_move_dst")

		// Clean up
		_ = os.RemoveAll(src)
		_ = os.RemoveAll(dst)
		defer os.RemoveAll(dst)

		// Create source directory with files
		err := os.MkdirAll(src, 0755)
		assert.Nil(t, err)

		err = os.WriteFile(filepath.Join(src, "file1.txt"), []byte("content1"), 0644)
		assert.Nil(t, err)

		err = os.WriteFile(filepath.Join(src, "file2.txt"), []byte("content2"), 0644)
		assert.Nil(t, err)

		// Move directory
		err = MoveDir(src, dst)
		assert.Nil(t, err)

		// Verify source is removed
		assert.False(t, DestExists(src))

		// Verify destination exists with correct content
		assert.True(t, DestExists(dst))
		content, err := os.ReadFile(filepath.Join(dst, "file1.txt"))
		assert.Nil(t, err)
		assert.Equal(t, []byte("content1"), content)

		content, err = os.ReadFile(filepath.Join(dst, "file2.txt"))
		assert.Nil(t, err)
		assert.Equal(t, []byte("content2"), content)
	})

	t.Run("move with nested directories", func(t *testing.T) {
		src := filepath.Join(tmpDir, "test_move_nested_src")
		dst := filepath.Join(tmpDir, "test_move_nested_dst")

		// Clean up
		_ = os.RemoveAll(src)
		_ = os.RemoveAll(dst)
		defer os.RemoveAll(dst)

		// Create nested directory structure
		subdir := filepath.Join(src, "subdir1", "subdir2")
		err := os.MkdirAll(subdir, 0755)
		assert.Nil(t, err)

		err = os.WriteFile(filepath.Join(src, "root.txt"), []byte("root"), 0644)
		assert.Nil(t, err)

		err = os.WriteFile(filepath.Join(src, "subdir1", "level1.txt"), []byte("level1"), 0644)
		assert.Nil(t, err)

		err = os.WriteFile(filepath.Join(subdir, "level2.txt"), []byte("level2"), 0644)
		assert.Nil(t, err)

		// Move directory
		err = MoveDir(src, dst)
		assert.Nil(t, err)

		// Verify source is removed
		assert.False(t, DestExists(src))

		// Verify nested structure is preserved
		assert.True(t, DestExists(dst))
		assert.True(t, DestExists(filepath.Join(dst, "subdir1")))
		assert.True(t, DestExists(filepath.Join(dst, "subdir1", "subdir2")))

		content, err := os.ReadFile(filepath.Join(dst, "root.txt"))
		assert.Nil(t, err)
		assert.Equal(t, []byte("root"), content)

		content, err = os.ReadFile(filepath.Join(dst, "subdir1", "level1.txt"))
		assert.Nil(t, err)
		assert.Equal(t, []byte("level1"), content)

		content, err = os.ReadFile(filepath.Join(dst, "subdir1", "subdir2", "level2.txt"))
		assert.Nil(t, err)
		assert.Equal(t, []byte("level2"), content)
	})

	t.Run("move preserves permissions", func(t *testing.T) {
		src := filepath.Join(tmpDir, "test_move_perm_src")
		dst := filepath.Join(tmpDir, "test_move_perm_dst")

		// Clean up
		_ = os.RemoveAll(src)
		_ = os.RemoveAll(dst)
		defer os.RemoveAll(dst)

		// Create directory and file with specific permissions
		err := os.MkdirAll(src, 0755)
		assert.Nil(t, err)

		filePath := filepath.Join(src, "exec.sh")
		err = os.WriteFile(filePath, []byte("#!/bin/bash\necho test"), 0755)
		assert.Nil(t, err)

		// Get original permissions
		srcStat, err := os.Stat(filePath)
		assert.Nil(t, err)
		srcMode := srcStat.Mode()

		// Move directory
		err = MoveDir(src, dst)
		assert.Nil(t, err)

		// Verify permissions are preserved
		dstFilePath := filepath.Join(dst, "exec.sh")
		dstStat, err := os.Stat(dstFilePath)
		assert.Nil(t, err)
		assert.Equal(t, srcMode, dstStat.Mode())
	})

	t.Run("error when source does not exist", func(t *testing.T) {
		src := filepath.Join(tmpDir, "nonexistent_src")
		dst := filepath.Join(tmpDir, "test_move_err_dst")

		// Clean up
		_ = os.RemoveAll(src)
		_ = os.RemoveAll(dst)
		defer os.RemoveAll(dst)

		// Try to move non-existent directory
		err := MoveDir(src, dst)
		assert.NotNil(t, err)
	})

	t.Run("move empty directory", func(t *testing.T) {
		src := filepath.Join(tmpDir, "test_move_empty_src")
		dst := filepath.Join(tmpDir, "test_move_empty_dst")

		// Clean up
		_ = os.RemoveAll(src)
		_ = os.RemoveAll(dst)
		defer os.RemoveAll(dst)

		// Create empty directory
		err := os.MkdirAll(src, 0755)
		assert.Nil(t, err)

		// Move directory
		err = MoveDir(src, dst)
		assert.Nil(t, err)

		// Verify
		assert.False(t, DestExists(src))
		assert.True(t, DestExists(dst))
	})
}

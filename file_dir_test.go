package goutils

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

// TestWriteFileAtomicCreatesFile 验证文件不存在时能正常创建。
func TestWriteFileAtomicCreatesFile(t *testing.T) {
	dir := t.TempDir()
	pth := filepath.Join(dir, "new.conf")

	require.NoError(t, WriteFileAtomic(pth, []byte("hello\n"), 0644))

	got, err := os.ReadFile(pth)
	require.NoError(t, err)
	assert.Equal(t, "hello\n", string(got))
}

// TestWriteFileAtomicReplacesContent 验证内容被完整替换而非追加。
func TestWriteFileAtomicReplacesContent(t *testing.T) {
	dir := t.TempDir()
	pth := filepath.Join(dir, "target.conf")

	require.NoError(t, os.WriteFile(pth, []byte("a-very-long-original-content\n"), 0644))
	require.NoError(t, WriteFileAtomic(pth, []byte("SHORT\n"), 0644))

	got, err := os.ReadFile(pth)
	require.NoError(t, err)
	assert.Equal(t, "SHORT\n", string(got), "旧内容应被完整替换，不留残余")
}

// TestWriteFileAtomicPreservesExistingPermission 验证改写不会意外改变文件权限。
//
// 为什么需要这条断言：本函数通过「创建临时文件再 rename」实现，
// 而临时文件的权限若未按目标文件设置，一次改写就可能把 0600 的密钥文件
// 变成 0644，从而泄露内容。因此要求已存在文件的权限被完整沿用。
func TestWriteFileAtomicPreservesExistingPermission(t *testing.T) {
	dir := t.TempDir()
	pth := filepath.Join(dir, "secret.conf")

	require.NoError(t, os.WriteFile(pth, []byte("old"), 0600))

	// 故意传入一个更宽松的权限，函数应沿用文件原有权限。
	require.NoError(t, WriteFileAtomic(pth, []byte("new"), 0644))

	info, err := os.Stat(pth)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm(),
		"已存在文件的权限应被保留，不应被 perm 参数覆盖")
}

// TestWriteFileAtomicAppliesPermForNewFile 验证新建文件时使用传入的权限。
func TestWriteFileAtomicAppliesPermForNewFile(t *testing.T) {
	dir := t.TempDir()
	pth := filepath.Join(dir, "brand-new.conf")

	require.NoError(t, WriteFileAtomic(pth, []byte("x"), 0600))

	info, err := os.Stat(pth)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm())
}

// TestWriteFileAtomicLeavesNoTempFiles 验证成功写入后不残留临时文件。
func TestWriteFileAtomicLeavesNoTempFiles(t *testing.T) {
	dir := t.TempDir()
	pth := filepath.Join(dir, "target.conf")

	require.NoError(t, WriteFileAtomic(pth, []byte("data\n"), 0644))

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	assert.Len(t, entries, 1, "目录中应只剩目标文件，实际：%v", entries)
	assert.Equal(t, "target.conf", entries[0].Name())
}

// TestWriteFileAtomicReplacesViaRenameWithoutTruncatingTarget 验证写入采用
// 「写临时文件 + 原子 rename」，而不是原地截断。
//
// 这里用**结构性质**来断言，而不是「让写入失败再看文件是否受损」：
// 后者无法体现差异——无论哪种实现，在目录或文件不可写时都会在打开阶段
// 就失败，因而都不会留下被截断的文件。
//
// inode 变化是「通过 rename 替换」的直接证据：它意味着写入换了一个新的
// 文件实体，因此任一时刻的观察者要么看到完整旧内容、要么看到完整新内容，
// 不存在「已截断但尚未写入」的中间态。
func TestWriteFileAtomicReplacesViaRenameWithoutTruncatingTarget(t *testing.T) {
	dir := t.TempDir()
	pth := filepath.Join(dir, "target.conf")

	require.NoError(t, os.WriteFile(pth, []byte("original\n"), 0644))

	before, err := os.Stat(pth)
	require.NoError(t, err)

	require.NoError(t, WriteFileAtomic(pth, []byte("replaced\n"), 0644))

	after, err := os.Stat(pth)
	require.NoError(t, err)

	// 内容已替换。
	got, err := os.ReadFile(pth)
	require.NoError(t, err)
	assert.Equal(t, "replaced\n", string(got))

	// rename 会引入新的 inode；原地截断重写则不会。
	// 注意：这里比较 inode 号，而不是整个 FileInfo.Sys()——
	// 后者包含 atime/ctime 等易变字段，读取文件本身就会改变它们，会造成误判。
	assert.NotEqual(t, inodeOf(before), inodeOf(after),
		"目标文件的 inode 应发生变化，说明是通过 rename 替换而非原地重写")
}

// inodeOf 返回文件的 inode 号，用于判断文件是否被「替换」而非「原地改写」。
func inodeOf(info os.FileInfo) uint64 {
	if st, ok := info.Sys().(*syscall.Stat_t); ok {
		return st.Ino
	}

	return 0
}

// TestWriteFileAtomicFailureKeepsOriginalIntact 验证创建临时文件失败时，
// 原文件保持完好。
//
// 说明：如上一个用例所述，本用例与 os.WriteFile 的差异无法通过「失败时是否
// 截断」来体现（两者在打开阶段即失败）；这里断言的是本函数的行为正确性：
// 失败路径不应触碰目标文件。
func TestWriteFileAtomicFailureKeepsOriginalIntact(t *testing.T) {
	// 该用例依赖类 Unix 的目录权限语义，且 root 会绕过权限检查。
	//
	// 跳过不会留下覆盖缺口：本函数存在的核心目的是「先写临时文件、再原子替换」，
	// 该性质已由 TestWriteFileAtomicReplacesViaRenameWithoutTruncatingTarget
	// （断言 inode 发生变化，证明用了 rename）无条件覆盖。
	if os.Geteuid() == 0 {
		t.Skip("以 root 运行时目录权限不生效，跳过；核心原子性由 inode 断言覆盖")
	}

	parent := t.TempDir()
	dir := filepath.Join(parent, "readonly")
	require.NoError(t, os.Mkdir(dir, 0755))

	pth := filepath.Join(dir, "important.conf")
	const original = "critical-content-must-survive\n"
	require.NoError(t, os.WriteFile(pth, []byte(original), 0644))

	// 让目录不可写：创建临时文件将失败。
	require.NoError(t, os.Chmod(dir, 0555))

	t.Cleanup(func() { _ = os.Chmod(dir, 0755) })

	err := WriteFileAtomic(pth, []byte("replacement\n"), 0644)
	require.Error(t, err, "在只读目录中写入应失败")

	got, readErr := os.ReadFile(pth)
	require.NoError(t, readErr)
	assert.Equal(t, original, string(got), "写入失败时原文件应保持完好")
}

// TestWriteFileAtomicWorksInDeepDirectory 验证多级目录下的写入。
func TestWriteFileAtomicWorksInDeepDirectory(t *testing.T) {
	dir := t.TempDir()
	pth := filepath.Join(dir, "a", "b", "c")

	require.NoError(t, os.MkdirAll(pth, 0755))

	target := filepath.Join(pth, "deep.conf")
	require.NoError(t, WriteFileAtomic(target, []byte("deep\n"), 0644))

	got, err := os.ReadFile(target)
	require.NoError(t, err)
	assert.Equal(t, "deep\n", string(got))
}

// TestWriteFileAtomicEmptyContent 验证写入空内容。
func TestWriteFileAtomicEmptyContent(t *testing.T) {
	dir := t.TempDir()
	pth := filepath.Join(dir, "empty.conf")

	require.NoError(t, os.WriteFile(pth, []byte("something"), 0644))
	require.NoError(t, WriteFileAtomic(pth, []byte{}, 0644))

	got, err := os.ReadFile(pth)
	require.NoError(t, err)
	assert.Empty(t, got)
}

// 验证 copyFile 不再跟随符号链接
func TestVerifyCopyFileNoSymlinkFollow(t *testing.T) {
	dir := t.TempDir()

	src := filepath.Join(dir, "src.txt")
	require.NoError(t, os.WriteFile(src, []byte("source-data"), 0600))

	victim := filepath.Join(dir, "victim.txt")
	require.NoError(t, os.WriteFile(victim, []byte("victim-original"), 0600))

	link := filepath.Join(dir, "link.txt")
	require.NoError(t, os.Symlink(victim, link))

	// 模拟 copyDir 把 src 复制到 link 位置
	err := copyFile(src, link)
	require.Error(t, err, "目标已是符号链接时应拒绝")

	got, _ := os.ReadFile(victim)
	assert.Equal(t, "victim-original", string(got), "受害文件不应被写入")
}

// 验证 GetFileStat 不再把特殊文件误判为普通文件
func TestVerifyGetFileStatSpecialFiles(t *testing.T) {
	st, err := GetFileStat("/dev/null")
	require.NoError(t, err)
	require.True(t, st.Exists)

	assert.False(t, st.IsRegular, "/dev/null 是字符设备，不应被标记为普通文件")
	assert.False(t, st.IsDir)
	assert.False(t, st.IsLink)
}

// 验证 CreateFileIfNotExist 拒绝符号链接
func TestVerifyCreateFileIfNotExistRejectsSymlink(t *testing.T) {
	dir := t.TempDir()

	victim := filepath.Join(dir, "victim.txt")
	require.NoError(t, os.WriteFile(victim, []byte("original"), 0600))

	link := filepath.Join(dir, "link.txt")
	require.NoError(t, os.Symlink(victim, link))

	err := CreateFileIfNotExist(link, []byte("new"), 0600)
	require.Error(t, err, "目标为符号链接时应拒绝写入")

	got, _ := os.ReadFile(victim)
	assert.Equal(t, "original", string(got), "受害文件不应被写入")
}

// 验证 CreateDirIfNotExist 拒绝符号链接
func TestVerifyCreateDirIfNotExistRejectsSymlink(t *testing.T) {
	dir := t.TempDir()

	real := filepath.Join(dir, "real")
	require.NoError(t, os.Mkdir(real, 0755))

	link := filepath.Join(dir, "link")
	require.NoError(t, os.Symlink(real, link))

	err := CreateDirIfNotExist(link, 0755)
	require.Error(t, err, "符号链接不应被当作目录使用")
}

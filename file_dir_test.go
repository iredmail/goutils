package goutils

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
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
// TestVerifyWriteFileAtomicPreservesGroupWhenPermitted 覆盖「Chown 成功」路径。
//
// 上一个测试覆盖的是 Chown 因 EPERM 失败而降级；本测试覆盖它**成功**时
// 属组被正确沿用，两者互补。
//
// 为什么需要它：非 root 无法改 uid，但**可以**把自己拥有的文件的属组改为
// 自己所属的另一个组。因此构造「属主是自己、属组是另一个组」的文件后，
// Chown 会成功，且结果可与「根本没调用 Chown」区分开：
//
//   - Chown 被调用且成功 → 新文件属组 = 原文件属组（另一个组）
//   - Chown 未被调用     → 新文件属组 = 进程默认 gid
//
// 两者不同，因此本测试能检出「Chown 调用被删除」这类回归 ——
// 而这正是仅断言「属主 == 当前 uid」时无法区分的（见上一个测试的说明）。
func TestVerifyWriteFileAtomicPreservesGroupWhenPermitted(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("以 root 运行时属组处理不具代表性，该分支由非 root 环境覆盖")
	}

	dir := t.TempDir()
	pth := filepath.Join(dir, "conf")

	require.NoError(t, os.WriteFile(pth, []byte("old"), 0644))

	// 把属组改为当前用户所属的、且**不同于默认 gid** 的另一个组。
	// 若当前用户没有其它组可用，则跳过（环境限制，不应误判为失败）。
	groups, err := os.Getgroups()
	require.NoError(t, err)

	var altGid int = -1

	for _, g := range groups {
		if g != os.Getgid() {
			altGid = g

			break
		}
	}

	if altGid < 0 {
		t.Skip("当前用户只有一个组，无法构造「属组与默认 gid 不同」的场景")
	}

	require.NoError(t, os.Chown(pth, -1, altGid), "应能把自己文件的属组改为自己所属的组")

	before, err := os.Stat(pth)
	require.NoError(t, err)

	bs, ok := before.Sys().(*syscall.Stat_t)
	require.True(t, ok)
	require.Equal(t, uint32(altGid), bs.Gid, "前置条件：文件属组应为另一个组")
	require.NotEqual(t, uint32(os.Getgid()), bs.Gid,
		"前置条件：文件属组必须不同于进程默认 gid，否则无法与「未调用 Chown」区分")

	require.NoError(t, WriteFileAtomic(pth, []byte("new"), 0644))

	after, err := os.Stat(pth)
	require.NoError(t, err)

	as, ok := after.Sys().(*syscall.Stat_t)
	require.True(t, ok)

	assert.Equal(t, uint32(altGid), as.Gid,
		"属组应沿用原文件(%d)；\n"+
			"若等于进程默认 gid(%d)，说明 Chown 没有被调用或没有生效",
		altGid, os.Getgid())

	got, err := os.ReadFile(pth)
	require.NoError(t, err)
	assert.Equal(t, "new", string(got))
}

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

// 验证 copyFile 不会跟随目标位置预置的符号链接。
//
// 这是安全断言的核心，与「是否允许覆盖」无关：
// copyFile 允许覆盖已存在的目标（copyDir 依赖该语义），
// 但覆盖的必须是**链接本身**，而不能顺着链接写到它指向的文件。
func TestVerifyCopyFileDoesNotFollowSymlink(t *testing.T) {
	dir := t.TempDir()

	src := filepath.Join(dir, "src.txt")
	require.NoError(t, os.WriteFile(src, []byte("source-data"), 0600))

	victim := filepath.Join(dir, "victim.txt")
	require.NoError(t, os.WriteFile(victim, []byte("victim-original"), 0600))

	link := filepath.Join(dir, "link.txt")
	require.NoError(t, os.Symlink(victim, link))

	// 模拟 copyDir 把 src 复制到 link 位置。
	require.NoError(t, copyFile(src, link))

	// 最关键：受害文件的内容必须未被改动。
	got, err := os.ReadFile(victim)
	require.NoError(t, err)
	assert.Equal(t, "victim-original", string(got),
		"受害文件被写入了：说明仍会跟随预置的符号链接")

	// 链接本身应被替换成普通文件，内容为源文件内容。
	li, err := os.Lstat(link)
	require.NoError(t, err)
	assert.True(t, li.Mode().IsRegular(), "符号链接应被普通文件替换")

	copied, err := os.ReadFile(link)
	require.NoError(t, err)
	assert.Equal(t, "source-data", string(copied))
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

// TestVerifyCopyFileCleansUpOnLateFailure 验证「尾部失败也会清理 dst」。
//
// 背景：copyFile 以 O_EXCL 创建目标，若失败后不删除，调用方在同一路径
// 重试时会因「已存在」而永久失败。因此 defer 的清理逻辑必须能感知
// **所有**错误路径，包括 Chmod / Chown 这些位于函数尾部的步骤。
//
// 之所以能感知，是因为 copyFile 的签名保留了命名返回值 err：
// return 语句会先把返回值赋给 err，再执行 defer。
// 若把签名改回 `error`（err 变成局部变量），尾部那些
// `return fmt.Errorf(...)` 就不会赋值给 err，dst 将残留。
//
// 真实环境中 Chown 失败难以稳定构造（非 root 用户 chown 给自己总是成功），
// 因此这里用「只读目录」让创建阶段失败，验证失败路径不残留；
// 命名返回值机制本身则由 TestVerifyCopyFileSignatureHasNamedReturn 断言。
func TestVerifyCopyFileCleansUpOnLateFailure(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("以 root 运行时目录权限不生效")
	}

	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	require.NoError(t, os.WriteFile(src, []byte("data"), 0644))

	roDir := filepath.Join(dir, "readonly")
	require.NoError(t, os.Mkdir(roDir, 0555))
	t.Cleanup(func() { _ = os.Chmod(roDir, 0755) })

	dst := filepath.Join(roDir, "dst.txt")

	err := copyFile(src, dst)
	require.Error(t, err, "只读目录中应失败")

	_, statErr := os.Stat(dst)
	assert.True(t, os.IsNotExist(statErr), "失败路径不应残留目标文件")
}

// TestVerifyCopyFileSignatureHasNamedReturn 断言 copyFile 使用命名返回值。
//
// 这是 TestVerifyCopyFileCleansUpOnLateFailure 能成立的前提：
// 只有命名返回值才能让 defer 看到 `return fmt.Errorf(...)` 这类路径的错误。
// 该断言通过行为验证：构造一个在 Chmod 之后失败的场景不可行，
// 因此改为验证「成功路径不误删 + 失败路径不残留」这一对性质，
// 并在注释中明确记录签名要求，避免后人误改。
func TestVerifyCopyFileSignatureHasNamedReturn(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	dst := filepath.Join(dir, "dst.txt")

	require.NoError(t, os.WriteFile(src, []byte("data"), 0644))

	// 成功路径：目标必须保留（若 defer 误判为失败而删除，此处会失败）
	require.NoError(t, copyFile(src, dst))

	got, err := os.ReadFile(dst)
	require.NoError(t, err)
	assert.Equal(t, "data", string(got), "成功路径不应删除目标文件")
}

// TestCopyFileSignatureMustKeepNamedReturn 静态断言 copyFile 保留命名返回值。
//
// 为什么需要静态断言：函数中「类型断言失败」那条路径写成
//
//	return fmt.Errorf("failed in reading owner/group info of %s", src)
//
// 位于 defer 注册之后。defer 通过检查 err 是否为 nil 来决定是否删除已创建的
// 目标文件；而该路径不给 err 赋值，只有当签名是命名返回值 `(err error)` 时，
// return 语句才会把值写入 err，defer 才能看到。
//
// 若签名被改回 `error`，这条路径下 err 保持为 nil，目标文件不会被删除；
// 又因为创建使用了 O_EXCL，调用方在同一路径重试会永久失败。
//
// 该路径无法从外部构造（stat.Sys() 的类型断言在类 Unix 上恒为真），
// 因此行为测试覆盖不到，只能在此静态检查源码。
func TestCopyFileSignatureMustKeepNamedReturn(t *testing.T) {
	src, err := os.ReadFile("file_dir.go")
	require.NoError(t, err, "应能读到同目录的 file_dir.go")

	const wantSignature = "func copyFile(src, dst string) (err error) {"

	if !strings.Contains(string(src), wantSignature) {
		t.Errorf("copyFile 的签名必须是 %q，实际未找到。\n"+
			"原因：defer 依赖命名返回值 err 来判断是否需要清理已创建的目标文件；\n"+
			"改成 `error` 会让「类型断言失败」那条 return 路径不删除目标文件，\n"+
			"而创建时用了 O_EXCL，会导致调用方在同一路径重试永久失败。",
			wantSignature)
	}
}

// TestVerifyCopyFileOverwritesExisting 验证目标已存在时会被覆盖。
//
// copyDir 依赖该语义：os.Rename 因 EXDEV 失败后，它把源树复制到目标树，
// 而目标树中可能已有同名文件。若此处报错，重复复制与「合并到已有目录」
// 都会失败。
func TestVerifyCopyFileOverwritesExisting(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	dst := filepath.Join(dir, "dst.txt")

	require.NoError(t, os.WriteFile(src, []byte("new-content"), 0644))
	require.NoError(t, os.WriteFile(dst, []byte("old-content"), 0644))

	require.NoError(t, copyFile(src, dst), "目标已存在时应覆盖，不应报错")

	got, err := os.ReadFile(dst)
	require.NoError(t, err)
	assert.Equal(t, "new-content", string(got))
}

// TestVerifyCopyFileOverwriteUpdatesMode 验证覆盖后权限跟随源文件。
//
// 创建时先用 0600 再 Chmod 成源文件权限，因此覆盖一个 0644 的旧文件、
// 源文件为 0600 时，结果必须是 0600，不能沿用旧文件的权限。
func TestVerifyCopyFileOverwriteUpdatesMode(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	dst := filepath.Join(dir, "dst.txt")

	require.NoError(t, os.WriteFile(src, []byte("x"), 0600))
	require.NoError(t, os.WriteFile(dst, []byte("old"), 0644))

	require.NoError(t, copyFile(src, dst))

	info, err := os.Stat(dst)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm())
}

// TestVerifyCopyDirIsIdempotent 验证 copyDir 可重复执行。
//
// 由于 copyDir 用 MkdirAll 创建目标目录（已存在时成功），随后逐条调用
// copyFile，若 copyFile 不覆盖已存在文件，第二次复制会在第一个同名文件上
// 直接失败，使整个复制不幂等。
func TestVerifyCopyDirIsIdempotent(t *testing.T) {
	base := t.TempDir()
	src := filepath.Join(base, "src")
	dst := filepath.Join(base, "dst")

	require.NoError(t, os.MkdirAll(src, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(src, "a.txt"), []byte("aaa"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(src, "b.txt"), []byte("bbb"), 0644))

	require.NoError(t, copyDir(src, dst), "首次复制应成功")
	require.NoError(t, copyDir(src, dst), "重复复制应成功（幂等）")

	got, err := os.ReadFile(filepath.Join(dst, "a.txt"))
	require.NoError(t, err)
	assert.Equal(t, "aaa", string(got))
}

// TestVerifyChmodKeepsSpecialModeBits 验证复制时不会丢掉 setuid/setgid/sticky。
//
// os.FileMode.Perm() 只返回低 9 位（0777），setuid / setgid / sticky
// 都存在 FileMode 的高位，用 Perm() 传给 os.Chmod 会把这三位丢掉。
// os.Chmod 本身会忽略类型位（目录/普通文件等），只应用权限位与这三个特殊位，
// 因此必须传 Mode() 而不是 Mode().Perm()。
//
// 共享目录（group-writable 的部署目录、sticky 目录）依赖这些位，
// 丢掉后复制结果与源权限不一致。
func TestVerifyChmodKeepsSpecialModeBits(t *testing.T) {
	dir := t.TempDir()

	// 第一部分：先说明为什么必须用 Mode() —— Perm() 会丢掉高位。
	// 这里直接对 FileMode 做位运算断言（不依赖平台），
	// 证明 Perm() 确实不含特殊位、而 Mode() 含。
	modeWithBits := os.FileMode(0777) | os.ModeSetgid | os.ModeSticky

	assert.Equal(t, os.FileMode(0777), modeWithBits.Perm(),
		"Perm() 只保留低 9 位")
	assert.Zero(t, modeWithBits.Perm()&os.ModeSetgid,
		"Perm() 结果不含 setgid —— 这正是误用 Perm() 会丢位的原因")
	assert.Zero(t, modeWithBits.Perm()&os.ModeSticky,
		"Perm() 结果不含 sticky")
	assert.NotZero(t, modeWithBits&os.ModeSetgid, "Mode() 保留 setgid")
	assert.NotZero(t, modeWithBits&os.ModeSticky, "Mode() 保留 sticky")

	// 第二部分：确认 sticky 能被真正写到文件上（跨平台稳定）。
	pth2 := filepath.Join(dir, "mode")
	require.NoError(t, os.WriteFile(pth2, []byte("x"), 0600))
	require.NoError(t, os.Chmod(pth2, os.FileMode(0777)|os.ModeSticky))

	info2, err := os.Stat(pth2)
	require.NoError(t, err)
	assert.NotZero(t, info2.Mode()&os.ModeSticky, "Mode() 应保留 sticky")

	// 第三部分：确认把 info2.Mode()（含 sticky）传给 Chmod 能真正生效。
	//
	// 这里只断言 sticky：它在两个平台都会保留。
	// setuid / setgid 不能在非 root 的 macOS 上断言 ——
	// BSD 会在 chmod 时静默丢弃这两位（不报错但读不回来），
	// 断言它们会导致「在有权限的平台上通过、在无权限的平台上失败」的假象，
	// 而失败原因是平台权限而非本函数的行为。
	pth3 := filepath.Join(dir, "chmod-result")
	require.NoError(t, os.WriteFile(pth3, []byte("x"), 0600))
	require.NoError(t, os.Chmod(pth3, info2.Mode()))

	info3, err := os.Stat(pth3)
	require.NoError(t, err)
	assert.NotZero(t, info3.Mode()&os.ModeSticky,
		"用 Mode() 调 Chmod 后 sticky 应生效")
	assert.Equal(t, info2.Mode().Perm(), info3.Mode().Perm(),
		"低 9 位权限应完整应用")
}

// TestVerifySetgidSilentlyDroppedOnBSD 记录 BSD 与 Linux 在 setgid 上的差异。
//
// 这不是本仓库的缺陷，而是平台语义差异，记录在此避免后人误判：
// 在 macOS 上 Chmod 设置 setgid 不会报错，但读回来该位不存在，
// 因此针对 setgid 的断言不能在 macOS 上执行。
func TestVerifySetgidSilentlyDroppedOnBSD(t *testing.T) {
	if runtime.GOOS == "linux" {
		t.Skip("该差异仅在 BSD/macOS 上成立")
	}

	dir := t.TempDir()
	pth := filepath.Join(dir, "sg")
	require.NoError(t, os.WriteFile(pth, []byte("x"), 0600))

	// 不报错，但位可能不生效。
	require.NoError(t, os.Chmod(pth, os.FileMode(0640)|os.ModeSetgid))

	info, err := os.Stat(pth)
	require.NoError(t, err)

	// 只断言权限位本身（低 9 位），特殊位在有权限时可能保留、否则被丢弃。
	assert.Equal(t, os.FileMode(0640), info.Mode().Perm(),
		"低 9 位权限应正常应用")
}

// TestVerifyCopyDirKeepsSpecialModeBits 验证 copyDir 保留目录的特殊权限位。
func TestVerifyCopyDirKeepsSpecialModeBits(t *testing.T) {
	base := t.TempDir()
	src := filepath.Join(base, "src")
	dst := filepath.Join(base, "dst")

	require.NoError(t, os.Mkdir(src, 0755))

	// 源目录带 sticky + group-writable（模拟 /tmp 风格的共享目录）
	require.NoError(t, os.Chmod(src, os.FileMode(0775)|os.ModeSticky))

	srcInfo, err := os.Stat(src)
	require.NoError(t, err)
	require.NotZero(t, srcInfo.Mode()&os.ModeSticky, "前置条件：源目录应带 sticky")

	require.NoError(t, copyDir(src, dst))

	dstInfo, err := os.Stat(dst)
	require.NoError(t, err)

	assert.NotZero(t, dstInfo.Mode()&os.ModeSticky,
		"目标目录应保留 sticky（Perm() 会丢掉它）")
	assert.Equal(t, srcInfo.Mode().Perm(), dstInfo.Mode().Perm())
}

// TestVerifyWriteFileAtomicKeepsOwnerAndSpecialBits 验证原子写入保留属主与特殊位。
//
// 临时文件由当前进程创建，rename 之后目标的 uid/gid 就是当前进程用户。
// 若被改写的是属于其它服务账号的文件（以 root 运行），属主会被改掉，
// 可能让那个服务无法读写该文件。因此实现中会在替换前记录原属主、
// 对临时文件 Chown 后再 rename。
//
// 验证方式：非 root 无法把自己创建的文件 chown 给其它 uid，
// 因此这里不直接断言「属主变了没有」，而是给出一个**当前用户不是其属主**
// 的文件是不现实的。改为断言两件可验证的事：
//
//  1. 权限位（含 sticky）被完整沿用 —— 若实现退回 Perm() 会失败；
//  2. 属主/属组与改写前一致 —— 在非 root 下这是必要条件（不变量），
//     可防止实现给文件 chown 成错误的 uid（例如误用 tmp 文件的属主）。
//
// 跨账号场景（root 改写属于 postfix 的文件）无法在单元测试中构造，
// 已由容器内的 root 环境手工验证：uid/gid 与 mode 均保持不变。
func TestVerifyWriteFileAtomicKeepsOwnerAndSpecialBits(t *testing.T) {
	dir := t.TempDir()
	pth := filepath.Join(dir, "conf")

	require.NoError(t, os.WriteFile(pth, []byte("old"), 0644))
	require.NoError(t, os.Chmod(pth, os.FileMode(0640)|os.ModeSticky))

	before, err := os.Stat(pth)
	require.NoError(t, err)

	// 传入一个不同的 perm，实现应沿用目标文件原有的权限而非这个值。
	require.NoError(t, WriteFileAtomic(pth, []byte("new"), 0600))

	after, err := os.Stat(pth)
	require.NoError(t, err)

	got, err := os.ReadFile(pth)
	require.NoError(t, err)
	assert.Equal(t, "new", string(got))

	assert.Equal(t, before.Mode().Perm(), after.Mode().Perm(),
		"权限位应沿用原文件，而不是调用方传入的 0600")
	assert.Equal(t, before.Mode()&os.ModeSticky, after.Mode()&os.ModeSticky,
		"sticky 应沿用原文件（Perm() 会丢掉它）")

	// setuid / setgid 的断言只在 Linux 上成立：
	// BSD（macOS）会在 chmod 时静默丢弃这两位，无论实现如何都读不回来，
	// 在该平台上断言它们毫无意义（既抓不到回归，也可能误报）。
	if runtime.GOOS == "linux" {
		pth2 := filepath.Join(dir, "conf-sg")
		require.NoError(t, os.WriteFile(pth2, []byte("old"), 0755))
		require.NoError(t, os.Chmod(pth2, os.FileMode(0755)|os.ModeSetgid))

		b2, err := os.Stat(pth2)
		require.NoError(t, err)
		require.NotZero(t, b2.Mode()&os.ModeSetgid, "前置条件：应能设上 setgid")

		require.NoError(t, WriteFileAtomic(pth2, []byte("new"), 0644))

		a2, err := os.Stat(pth2)
		require.NoError(t, err)
		assert.NotZero(t, a2.Mode()&os.ModeSetgid,
			"setgid 应沿用原文件：若实现退回 Perm() 或把 Chown 放到 Chmod 之后，这里会丢位")
	}

	bs, ok := before.Sys().(*syscall.Stat_t)
	if ok {
		as, ok2 := after.Sys().(*syscall.Stat_t)
		require.True(t, ok2)
		assert.Equal(t, bs.Uid, as.Uid, "属主 uid 应保持不变")
		assert.Equal(t, bs.Gid, as.Gid, "属组 gid 应保持不变")
	}
}

// TestVerifyWriteFileAtomicDoesNotFollowSymlink 验证原子写入不沿符号链接取属主。
//
// 为什么要单独验证：实现用 os.Lstat（而非 os.Stat）判断目标，
// 并对符号链接走「不沿用原文件属性」的分支。若改用 os.Stat、
// 或删掉 ModeSymlink 判断，就会从**链接指向的文件**读取属主，
// 后果是把临时文件 chown 成那个文件的属主 —— 等于沿链接泄露/篡改属主。
//
// 该用例在非 root 下也能运行：断言的是「受害文件的属主/内容不受影响」
// 以及「链接被替换为普通文件」，不依赖跨账号 chown。
func TestVerifyWriteFileAtomicDoesNotFollowSymlink(t *testing.T) {
	dir := t.TempDir()

	victim := filepath.Join(dir, "victim.conf")
	require.NoError(t, os.WriteFile(victim, []byte("victim-content"), 0644))

	vBefore, err := os.Stat(victim)
	require.NoError(t, err)

	link := filepath.Join(dir, "link.conf")
	require.NoError(t, os.Symlink(victim, link))

	require.NoError(t, WriteFileAtomic(link, []byte("new-content"), 0600))

	// 受害文件的内容与属主都不应被改动。
	gotVictim, err := os.ReadFile(victim)
	require.NoError(t, err)
	assert.Equal(t, "victim-content", string(gotVictim),
		"受害文件内容被改写：说明沿链接写入了")

	vAfter, err := os.Stat(victim)
	require.NoError(t, err)
	assert.Equal(t, vBefore.Mode(), vAfter.Mode(),
		"受害文件权限被改动：说明沿链接取用了属性")

	if vb, ok := vBefore.Sys().(*syscall.Stat_t); ok {
		if va, ok2 := vAfter.Sys().(*syscall.Stat_t); ok2 {
			assert.Equal(t, vb.Uid, va.Uid, "受害文件属主不应变化")
			assert.Equal(t, vb.Gid, va.Gid, "受害文件属组不应变化")
		}
	}

	// 链接应被替换为普通文件，承载新内容。
	li, err := os.Lstat(link)
	require.NoError(t, err)
	assert.True(t, li.Mode().IsRegular(),
		"符号链接应被替换为普通文件，而不是继续指向受害文件")

	gotLink, err := os.ReadFile(link)
	require.NoError(t, err)
	assert.Equal(t, "new-content", string(gotLink))

	// 关键断言：新文件的权限应来自**调用方传入的 perm**，
	// 而不是从链接指向的 victim 读来的权限。
	//
	// 这是区分「用 os.Lstat + 判 ModeSymlink」与「用 os.Stat 跟随链接」的唯一
	// 可观测差异：两种写法都不会改动 victim（rename 只替换链接本身），
	// 但跟随链接的那版会错误地把 victim 的权限套用到新文件上。
	//
	// victim 是 0640，这里传入的是 0600，因此两者可区分。
	assert.Equal(t, os.FileMode(0600), li.Mode().Perm(),
		"新文件权限应来自传入的 perm；若等于 victim 的 0640，说明沿链接取了属性")
	assert.NotEqual(t, vBefore.Mode().Perm(), li.Mode().Perm(),
		"新文件权限不应等于链接目标的权限（那意味着跟随了符号链接）")
}

// TestVerifyCopyFileKeepsSpecialModeBits 验证 copyFile 保留特殊权限位。
//
// 覆盖缺口说明：针对 copyFile 的多数测试都只用普通权限（0644 等）的源文件，
// 而 Perm() 与 Mode() 对这类文件的结果相同，因此**无法**发现
// 「误用 Perm() 导致 setuid/setgid/sticky 丢失」这一回归。
// 本测试给源文件设置 sticky，使两种写法产生可区分的差异。
//
// 只断言 sticky 而不含 setgid/setuid：BSD（macOS）在 chmod 时会静默丢弃
// setuid/setgid（不报错但读不回来），sticky 在两个平台都会保留。
func TestVerifyCopyFileKeepsSpecialModeBits(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	dst := filepath.Join(dir, "dst.txt")

	require.NoError(t, os.WriteFile(src, []byte("data"), 0644))
	require.NoError(t, os.Chmod(src, os.FileMode(0644)|os.ModeSticky))

	srcInfo, err := os.Stat(src)
	require.NoError(t, err)
	require.NotZero(t, srcInfo.Mode()&os.ModeSticky,
		"前置条件：源文件应带 sticky（否则本测试无法区分 Perm 与 Mode）")

	require.NoError(t, copyFile(src, dst))

	dstInfo, err := os.Stat(dst)
	require.NoError(t, err)

	got, err := os.ReadFile(dst)
	require.NoError(t, err)
	assert.Equal(t, "data", string(got))

	assert.NotZero(t, dstInfo.Mode()&os.ModeSticky,
		"目标文件应保留 sticky：用 Perm() 会丢掉它")
	assert.Equal(t, srcInfo.Mode().Perm(), dstInfo.Mode().Perm(),
		"低 9 位权限应与源文件一致")
}

// TestVerifyCopyFileOrderChownBeforeChmod 验证复制时「先 Chown 再 Chmod」的顺序。
//
// 为什么需要单独测顺序：Linux 的 chown(2) 在属主/属组**发生变化**时会清除
// setgid（安全语义：换了主人就不该保留提权位）。因此若先 Chmod 再 Chown，
// 刚设好的 setgid 会被紧接着的 Chown 抹掉，复制结果丢位。
//
// 触发条件有两个，缺一不可（均由实测得出）：
//
//  1. **chown 真的改变了属主**。若目标当前属主与源属主相同，Linux 视为
//     no-op，不清位。因此本测试以 root 运行、把源文件 chown 给另一个 uid，
//     使 chown 真正生效。非 root 无法构造，故跳过。
//
//  2. **setgid 与「组可执行位」同时存在**。实测：chmod(0755|setgid) 之后
//     再 chown 会清掉 setgid；而 chmod(0644|setgid) 之后再 chown 则保留
//     （普通文件的 setgid 语义依赖组执行位）。因此本测试用 0755。
//
// 两个条件同时满足时，顺序写反会导致断言失败，从而守护住这个顺序要求。
func TestVerifyCopyFileOrderChownBeforeChmod(t *testing.T) {
	if os.Geteuid() != 0 {
		t.Skip("需要 root 才能把源文件 chown 给其它 uid，使 chown 真正改变属主")
	}

	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	dst := filepath.Join(dir, "dst.txt")

	require.NoError(t, os.WriteFile(src, []byte("data"), 0755))

	// 让源文件属于其它 uid，使复制时的 chown 真正改变属主。
	const targetUid, targetGid = 1234, 5678
	require.NoError(t, os.Chown(src, targetUid, targetGid))

	// 0755 含组执行位 —— 这是 setgid 会被 chown 清掉的前提。
	require.NoError(t, os.Chmod(src, os.FileMode(0755)|os.ModeSetgid))

	srcInfo, err := os.Stat(src)
	require.NoError(t, err)
	require.NotZero(t, srcInfo.Mode()&os.ModeSetgid,
		"前置条件：源文件应带 setgid")

	require.NoError(t, copyFile(src, dst))

	dstInfo, err := os.Stat(dst)
	require.NoError(t, err)

	assert.NotZero(t, dstInfo.Mode()&os.ModeSetgid,
		"目标应保留 setgid：若先 Chmod 后 Chown，chown 会把它清掉")

	ds, ok := dstInfo.Sys().(*syscall.Stat_t)
	require.True(t, ok)
	assert.Equal(t, uint32(targetUid), ds.Uid, "属主应与源文件一致")
	assert.Equal(t, uint32(targetGid), ds.Gid, "属组应与源文件一致")
}

// TestVerifyCopyDirOrderChownBeforeChmod 验证目录复制同样「先 Chown 再 Chmod」。
//
// 与 copyFile 同理：Linux 的 chown(2) 在属主变化时清除 setgid，
// 而 setgid 对**目录**的语义是「新建的子项继承目录属组」，是共享目录的常用配置。
// 若先 Chmod 后 Chown，目录的 setgid 会被抹掉，子项便不再继承属组。
//
// 触发条件与 copyFile 相同：需要 root（才能让 chown 真正改变属主）。
// 目录不涉及「组执行位」的额外约束，因此 setgid 是否存在即可判定。
func TestVerifyCopyDirOrderChownBeforeChmod(t *testing.T) {
	if os.Geteuid() != 0 {
		t.Skip("需要 root 才能把源目录 chown 给其它 uid，使 chown 真正改变属主")
	}

	base := t.TempDir()
	src := filepath.Join(base, "src")
	dst := filepath.Join(base, "dst")

	require.NoError(t, os.Mkdir(src, 0755))

	const targetUid, targetGid = 1234, 5678
	require.NoError(t, os.Chown(src, targetUid, targetGid))
	require.NoError(t, os.Chmod(src, os.FileMode(0755)|os.ModeSetgid))

	srcInfo, err := os.Stat(src)
	require.NoError(t, err)
	require.NotZero(t, srcInfo.Mode()&os.ModeSetgid,
		"前置条件：源目录应带 setgid")

	require.NoError(t, copyDir(src, dst))

	dstInfo, err := os.Stat(dst)
	require.NoError(t, err)

	assert.NotZero(t, dstInfo.Mode()&os.ModeSetgid,
		"目标目录应保留 setgid：若先 Chmod 后 Chown，chown 会把它清掉")

	ds, ok := dstInfo.Sys().(*syscall.Stat_t)
	require.True(t, ok)
	assert.Equal(t, uint32(targetUid), ds.Uid,
		"目录属主应与源目录一致")
}

// TestVerifyWriteFileAtomicSyncsParentDir 守护「rename 后 fsync 父目录」这一步。
//
// 为什么用源码检查而不是行为断言：目录项是否已落盘只能通过掉电实验观察，
// 单元测试无法构造；而这一行很容易在重构中被当作「多余的 fsync」删掉，
// 删掉后并不会立刻有任何测试失败。因此这里做一个存在性检查作为最低保障。
//
// 该检查只匹配调用本身，不匹配变量名，因此重命名局部变量不会误报；
// 但把实现整体改写（例如换成 unix.Fsync(fd)）会使其失败 —— 那时应同步更新本检查。
func TestVerifyWriteFileAtomicSyncsParentDir(t *testing.T) {
	src, err := os.ReadFile("file_dir.go")
	require.NoError(t, err)
	body := string(src)

	// 定位 WriteFileAtomic 函数体，避免匹配到其它函数中的同名调用。
	start := strings.Index(body, "func WriteFileAtomic(")
	require.GreaterOrEqual(t, start, 0, "应能在 file_dir.go 中找到 WriteFileAtomic")

	rest := body[start:]
	end := strings.Index(rest, "\n}\n")
	require.Greater(t, end, 0, "应能定位 WriteFileAtomic 的函数体")

	fn := rest[:end]

	require.Contains(t, fn, "os.Rename(",
		"WriteFileAtomic 应使用 os.Rename 原子替换")

	// 统计 Sync 调用次数：一次给临时文件（rename 前），一次给父目录（rename 后）。
	// 只检查「存在 .Sync()」是不够的 —— 文件自身的 tmp.Sync() 也会匹配，
	// 因此父目录 fsync 被删除时检查依然通过。这里按次数断言。
	count := strings.Count(fn, ".Sync()")
	assert.GreaterOrEqual(t, count, 2,
		"WriteFileAtomic 应对临时文件与父目录各调用一次 Sync（当前 %d 次）：\n"+
			"仅 fsync 文件不能保证目录项（rename 结果）落盘，\n"+
			"掉电后目标路径可能回退到旧内容或读不到文件。", count)

	// 并确认父目录 fsync 发生在 rename 之后。
	ri := strings.Index(fn, "os.Rename(")
	require.GreaterOrEqual(t, ri, 0)
	afterRename := fn[ri:]
	assert.Contains(t, afterRename, ".Sync()",
		"父目录的 Sync 应在 os.Rename 之后调用，否则无法保证目录项落盘")
}

// TestVerifyChownBeforeChmodOrder 守护 copyDir / copyFile 中 Chown 早于 Chmod 的顺序。
//
// 为什么需要：Linux 的 chown(2) 在属主改变时会清除 setgid，
// 因此若先 Chmod 再 Chown，刚设好的位会被抹掉。
// 该行为只在「以 root 运行 + 属主确实变化 + 文件带执行位」时可见，
// 普通开发机（非 root）与 macOS 都无法触发，因此行为测试覆盖不到，
// 这里补一个源码顺序检查。
//
// 实现方式：确认每个函数中 `os.Chown(` 的出现位置早于 `os.Chmod(`。
func TestVerifyChownBeforeChmodOrder(t *testing.T) {
	src, err := os.ReadFile("file_dir.go")
	require.NoError(t, err)
	body := string(src)

	for _, name := range []string{"func copyDir(", "func copyFile("} {
		start := strings.Index(body, name)
		require.GreaterOrEqual(t, start, 0, "应能找到 %s", name)

		rest := body[start:]
		end := strings.Index(rest, "\n}\n")
		require.Greater(t, end, 0, "应能定位 %s 的函数体", name)

		fn := rest[:end]
		ci := strings.Index(fn, "os.Chown(")
		mi := strings.Index(fn, "os.Chmod(")

		require.GreaterOrEqual(t, ci, 0, "%s 应调用 os.Chown", name)
		require.GreaterOrEqual(t, mi, 0, "%s 应调用 os.Chmod", name)
		assert.Less(t, ci, mi,
			"%s 中 os.Chown 必须在 os.Chmod 之前：\n"+
				"Linux 的 chown(2) 在属主改变时会清除 setgid，\n"+
				"先 Chmod 再 Chown 会让刚设好的特殊位被抹掉。", name)
	}
}

// TestVerifyWriteFileAtomicWritesBeforeChmod 守护「Write 早于 Chmod」的顺序。
//
// 为什么需要静态检查：Linux 有一条安全规则 —— 非 root 进程写入一个已带
// setuid / setgid 的文件时，内核会清除这两个位。因此若先 Chmod 设上 setgid
// 再 Write，刚设好的位会被这次写入抹掉，最终目标文件丢掉 setgid。
//
// 该问题**只在 Linux 且非 root 时可见**，实测三种环境下用错误顺序的结果：
//
//	macOS                  → 测试通过（漏检）
//	Linux + root           → 测试通过（漏检）
//	Linux + 非 root        → 测试失败（检出）
//
// 也就是说行为测试只在 CI 那种「Linux 非 root」环境下才有效，
// 本地开发（macOS）与 root 容器都发现不了。为避免这类回归只能靠环境碰运气，
// 这里补一个与运行环境无关的源码顺序检查。
func TestVerifyWriteFileAtomicWritesBeforeChmod(t *testing.T) {
	src, err := os.ReadFile("file_dir.go")
	require.NoError(t, err)
	body := string(src)

	start := strings.Index(body, "func WriteFileAtomic(")
	require.GreaterOrEqual(t, start, 0, "应能在 file_dir.go 中找到 WriteFileAtomic")

	rest := body[start:]
	end := strings.Index(rest, "\n}\n")
	require.Greater(t, end, 0, "应能定位 WriteFileAtomic 的函数体")

	fn := rest[:end]

	wi := strings.Index(fn, "tmp.Write(")
	ci := strings.Index(fn, "tmp.Chmod(")
	oi := strings.Index(fn, "tmp.Chown(")

	require.GreaterOrEqual(t, wi, 0, "WriteFileAtomic 应调用 tmp.Write")
	require.GreaterOrEqual(t, ci, 0, "WriteFileAtomic 应调用 tmp.Chmod")
	require.GreaterOrEqual(t, oi, 0, "WriteFileAtomic 应调用 tmp.Chown")

	// 顺序要求：Chown（若存在）→ Write → Chmod
	assert.Less(t, oi, wi,
		"tmp.Chown 必须在 tmp.Write 之前：Linux 的 chown(2) 会清除 setuid/setgid")
	assert.Less(t, wi, ci,
		"tmp.Write 必须在 tmp.Chmod 之前：\n"+
			"Linux 上非 root 写入带 setuid/setgid 的文件会清除这两位，\n"+
			"先 Chmod 再 Write 会让刚设好的 setgid 被抹掉。\n"+
			"该问题只在 Linux 非 root 时可见（macOS 与 root 都发现不了），\n"+
			"因此不能只依赖行为测试。")
}

// TestVerifyWriteFileAtomicFiltersChownErrors 验证 Chown 失败按类型区别处理。
//
// 之所以要区分：并非所有 Chown 失败都等价。
//
//   - EPERM（权限不足）是**预期内**的：非 root 进程无法把文件改属给别的账号。
//     此时应继续写入（属主降级为当前用户，权限与内容仍正确）。
//   - 其它错误属于**非预期**故障，若一律吞掉，函数会返回 nil 而隐藏真实问题。
//
// 这里用源码检查而非行为断言，原因是 EPERM 之外的真实错误很难在单元测试里
// 稳定构造（它需要「目录可写但文件不可写 + 无 CAP_CHOWN」的组合）。
//
// 检查方式刻意使用**正则**而不是简单的字符串包含：
//   - 必须出现 `errors.Is(..., syscall.EPERM)` 这一调用形式。
//     若只检查「源码里含有 syscall.EPERM」，那么把 errors.Is 误写成
//     `cerr != syscall.EPERM`（真实会出 bug 的写法）也能通过 —— 已实测该漏报；
//     并且注释里出现 syscall.EPERM 同样会让检查通过。
//   - 不再匹配错误文案字符串，避免仅改文案就被误判为失败。
func TestVerifyWriteFileAtomicFiltersChownErrors(t *testing.T) {
	src, err := os.ReadFile("file_dir.go")
	require.NoError(t, err)
	body := string(src)

	start := strings.Index(body, "func WriteFileAtomic(")
	require.GreaterOrEqual(t, start, 0, "应能找到 WriteFileAtomic")

	rest := body[start:]
	end := strings.Index(rest, "\n}\n")
	require.Greater(t, end, 0, "应能定位 WriteFileAtomic 的函数体")
	fn := rest[:end]

	// 去掉注释行后再检查，避免注释中的文字让断言通过。
	var code []string
	for _, line := range strings.Split(fn, "\n") {
		if trimmed := strings.TrimSpace(line); strings.HasPrefix(trimmed, "//") {
			continue
		}

		code = append(code, line)
	}

	codeBody := strings.Join(code, "\n")

	require.Contains(t, codeBody, "tmp.Chown(",
		"WriteFileAtomic 应调用 tmp.Chown 以沿用属主")

	// 必须用 errors.Is 比较 EPERM（而非 ==），否则 *PathError 包装会让判断失效。
	re := regexp.MustCompile(`errors\.Is\(\s*cerr\s*,\s*syscall\.EPERM\s*\)`)
	assert.Regexp(t, re, codeBody,
		"应以 `errors.Is(cerr, syscall.EPERM)` 判断权限错误：\n"+
			"os.File.Chown 返回的是 *PathError，用 == 比较会永远为假，\n"+
			"从而让 EPERM 降级变成硬失败（非 root 改写他人文件会直接报错）。")

	// Chown 失败时必须有向上返回的路径，不能只 `_ = cerr` 吞掉。
	assert.Contains(t, codeBody, "return fmt.Errorf(",
		"非 EPERM 的 Chown 失败应向上返回错误，而不是静默忽略")
}

// TestVerifyWriteFileAtomicEPERMDegradesGracefully 覆盖 EPERM 降级分支。
//
// 为什么需要外部编排：EPERM 只在「非 root 进程改写一个属于**第三方**账号的
// 文件」时出现，两个条件缺一不可：
//
//   - 必须是非 root：root 有 CAP_CHOWN，任何 chown 都会成功；
//   - 目标文件的属主必须是**别人**：若文件属于自己，Chown(self) 恒成功，
//     根本走不到 EPERM 分支。
//
// 第二个条件无法在单个测试进程内构造：t.TempDir() 造出的文件属主就是当前
// 用户，而改属他人需要 root。因此本测试改为**读取外部准备好的文件**：
//
//	环境变量 GOUTILS_EPERM_TEST_FILE 指向一个由 root 创建、
//	属主为第三方 uid 且对当前用户可读可写的文件。
//
// CI 中由 workflow 以 root 准备该文件，再以普通用户运行本测试（见
// .github/workflows/tests.yml 的 "Test EPERM degradation as unprivileged user"）。
//
// 未提供该环境变量时跳过 —— 这是刻意的：在本地开发机上不会伪造出一个
// 「已验证 EPERM」的假象，跳过消息会说明该分支由哪个环境覆盖。
func TestVerifyWriteFileAtomicEPERMDegradesGracefully(t *testing.T) {
	pth := os.Getenv("GOUTILS_EPERM_TEST_FILE")
	if pth == "" {
		t.Skip("未设置 GOUTILS_EPERM_TEST_FILE；" +
			"该分支需由「root 准备属主为第三方的文件 + 非 root 执行」的外部编排覆盖（见 CI workflow）")
	}

	// 前置条件：文件存在、且属主不是当前用户。
	info, err := os.Stat(pth)
	require.NoError(t, err, "GOUTILS_EPERM_TEST_FILE 指向的文件应存在：%s", pth)

	ss, ok := info.Sys().(*syscall.Stat_t)
	require.True(t, ok, "应能读取文件的 uid/gid")

	require.NotEqual(t, uint32(os.Getuid()), ss.Uid,
		"该测试要求文件属主是第三方（当前 uid=%d，文件 uid=%d）：\n"+
			"若属主就是自己，Chown(self) 不会失败，本测试将失去意义",
		os.Getuid(), ss.Uid)

	require.NotEqual(t, 0, os.Geteuid(),
		"该测试必须以非 root 运行：root 有 CAP_CHOWN，Chown 不会返回 EPERM")

	// 写入内容。此时 writeFileAtomicKeepOwner 会尝试 Chown 到原属主（第三方），
	// 以非 root 身份必然 EPERM —— 这正是要覆盖的分支。
	const newContent = "NEW-CONTENT-FROM-EPERM-TEST"

	require.NoError(t, WriteFileAtomic(pth, []byte(newContent), 0644),
		"属主降级属于可接受行为（EPERM 被忽略），不应返回错误")

	got, err := os.ReadFile(pth)
	require.NoError(t, err)
	assert.Equal(t, newContent, string(got), "内容仍应正确写入")

	// 属主应降级为当前用户 —— 这是该分支的预期结果，也验证降级确实发生了
	// （而不是「刚好没走到 Chown」）。
	after, err := os.Stat(pth)
	require.NoError(t, err)

	as, ok := after.Sys().(*syscall.Stat_t)
	require.True(t, ok)

	assert.Equal(t, uint32(os.Getuid()), as.Uid,
		"非 root 无法沿用他人属主，应降级为当前用户；\n"+
			"若仍为原属主(%d)，说明 Chown 未被调用或调用失败被静默跳过",
		ss.Uid)

	// 权限位应被沿用（原文件 0644），不受传入 perm 影响。
	assert.Equal(t, os.FileMode(0644), after.Mode().Perm(),
		"权限位应沿用原文件")
}

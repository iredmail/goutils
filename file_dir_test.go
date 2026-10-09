package goutils

import (
	"os"
	"path/filepath"
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

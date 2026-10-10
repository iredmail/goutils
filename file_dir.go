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

// String 返回 FileStat 的可读描述，便于日志与调试。
//
// 除类型标记外，这里也带上属主/属组与权限，
// 因为这些是排查“文件为何未被修改”时最常用的信息。
func (fs *FileStat) String() string {
	return fmt.Sprintf(
		"Name: %s, Exists: %v, IsLink: %v, IsRegular: %v, IsDir: %v, Mode: %04o, Owner: %s, Group: %s",
		fs.Name, fs.Exists, fs.IsLink, fs.IsRegular, fs.IsDir, fs.Mode.Perm(), fs.Owner, fs.Group,
	)
}

// GetFileStat 获取目标路径的信息（不跟随符号链接）。
//
// 注意：使用 os.Lstat，因此若 pth 本身是符号链接，
// 返回的是**链接自身**的信息（IsLink = true），而不是它指向的目标。
//
// 路径不存在时返回零值 FileStat（Exists = false）且 err 为 nil，
// 调用方应以 fs.Exists 判断存在性。
//
// 类型判定说明（重要）：
// IsRegular **仅在文件确实是普通文件**时为 true，使用 stat.Mode().IsRegular() 判断。
// 四个类型标记是互斥的；FIFO、socket、字符/块设备（如 /dev/null）
// 三者皆为 false。调用方（file / template / copy2 等模块）应先确认
// IsRegular 为 true，再对文件执行 chmod / chown / 覆盖等操作，
// 避免把特殊文件当作普通文件处理。
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

	// 取 uid/gid 与属主/属组名。
	//
	// 注意：stat.Sys() 在不同操作系统上的具体类型不同，
	// 这里假定为类 Unix 平台的 *syscall.Stat_t（本仓库仅支持类 Unix）。
	ss, ok := stat.Sys().(*syscall.Stat_t)
	if !ok {
		return fs, fmt.Errorf("unsupported file info type for %s", pth)
	}

	fs.Uid = ss.Uid
	fs.Gid = ss.Gid

	// 查不到名称时保留空字符串即可（例如 uid 没有对应的 /etc/passwd 条目），
	// 这不影响 Exists / 类型等关键信息，因此不视为错误。
	if usr, err := user.LookupId(fmt.Sprintf("%d", fs.Uid)); err == nil {
		fs.Owner = usr.Username
	}

	if group, err := user.LookupGroupId(fmt.Sprintf("%d", fs.Gid)); err == nil {
		fs.Group = group.Name
	}

	// 按真实类型分类，互斥且完整：
	// 目录 → IsDir；符号链接 → IsLink；普通文件 → IsRegular；其余（FIFO、
	// socket、设备等）三者皆 false，调用方应据此拒绝操作，而不是当成普通文件。
	switch {
	case stat.IsDir():
		fs.IsDir = true
	case stat.Mode()&os.ModeSymlink != 0:
		fs.IsLink = true
	case stat.Mode().IsRegular():
		fs.IsRegular = true
	}

	return fs, nil
}

// DestExists 检查目标对象（文件、目录、符号链接，等）是否存在。
//
// 注意：使用 os.Stat，会**跟随符号链接**，因此：
//   - 指向有效目标的符号链接 → true；
//   - **悬空**符号链接（目标不存在）→ false。
//
// 若判断的是“链接本身是否存在”，应改用 os.Lstat。
func DestExists(pth string) bool {
	_, err := os.Stat(pth)

	return err == nil
}

// CreateDirIfNotExist 创建目标目录（若不存在）。
//
// 注意：
//   - MkdirAll 的 perm 参数会**受 umask 影响**（例如 umask 022 时 0777 实际为 0755）。
//     若调用方要求精确权限，应在调用后自行 Chmod。
//   - 使用 os.Lstat 判断存在性，不跟随符号链接：
//     符号链接即使是悬空的也不会被当成“不存在”，避免顺着链接去创建目录。
//   - 已存在但不是目录（普通文件、符号链接等）时返回明确错误。
func CreateDirIfNotExist(pth string, mode os.FileMode) error {
	info, err := os.Lstat(pth)
	if err != nil {
		if !os.IsNotExist(err) {
			// 除“不存在”外的错误（如权限不足、路径中间层不是目录）原样返回，
			// 交由调用方判断；此处不吞掉错误。
			return fmt.Errorf("failed in checking stat of %s: %w", pth, err)
		}

		if err = os.MkdirAll(pth, mode); err != nil {
			return fmt.Errorf("failed in creating directory %s: %w", pth, err)
		}

		return nil
	}

	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("%s exists and is a symbol link (refusing to use it as a directory)", pth)
	}

	if !info.IsDir() {
		return fmt.Errorf("%s exists, but not a directory", pth)
	}

	return nil
}

// CreateDirWithOGMIfNotExist 创建目录并设置属主/属组（若目录不存在）。
//
// 参数 ownerUid / ownerGid 为 -1 时，对应项保持不变
// （与 os.Chown 的语义一致）；传入其它负值会被 os.Chown 拒绝。
func CreateDirWithOGMIfNotExist(pth string, mode os.FileMode, ownerUid, ownerGid int) error {
	if err := CreateDirIfNotExist(pth, mode); err != nil {
		return err
	}

	if err := os.Chown(pth, ownerUid, ownerGid); err != nil {
		return fmt.Errorf("failed in setting owner/group of %s: %w", pth, err)
	}

	return nil
}

// CreateFileIfNotExist creates target file with given mode if it doesn't exist.
//
// 注意事项：
//   - 使用 os.Lstat 判断存在性，**不跟随符号链接**。
//     若用 os.Stat，一个指向不存在目标的「悬空符号链接」会被判为“不存在”，
//     随后的写入会顺着该链接落到它指向的真实路径上（若父目录存在），
//     使调用方在不知情的情况下写到别处。
//   - 若目标已是符号链接（无论是否悬空），直接报错而不是跟随写入。
//   - 目标已存在且是普通文件时直接返回 nil（不覆盖既有内容）。
func CreateFileIfNotExist(pth string, content []byte, mode os.FileMode) error {
	info, err := os.Lstat(pth)
	if err == nil {
		if info.IsDir() {
			return fmt.Errorf("%s is a directory (which should be a regular file)", pth)
		}

		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("%s is a symbol link (refusing to write through it)", pth)
		}

		if !info.Mode().IsRegular() {
			return fmt.Errorf("%s is not a regular file (mode=%v)", pth, info.Mode().Type())
		}

		// 已存在且是普通文件：不覆盖。
		return nil
	}

	if !os.IsNotExist(err) {
		// 其它错误（如权限不足）不应被当成“文件不存在”。
		return fmt.Errorf("failed in checking stat of file %s: %w", pth, err)
	}

	// 文件不存在：确保父目录存在（权限 0755），再创建。
	dir := filepath.Dir(pth)
	if err2 := CreateDirIfNotExist(dir, 0755); err2 != nil {
		return err2
	}

	// 以 O_EXCL 创建：若在 Lstat 与本次创建之间有人抢先创建了该路径
	// （可能是符号链接），O_EXCL 会让创建失败，而不是跟随写入。
	f, err2 := os.OpenFile(pth, os.O_WRONLY|os.O_CREATE|os.O_EXCL, mode)
	if err2 != nil {
		return fmt.Errorf("failed in creating file %s: %w", pth, err2)
	}

	if _, err2 = f.Write(content); err2 != nil {
		_ = f.Close()
		_ = os.Remove(pth)

		return fmt.Errorf("failed in writing to file %s: %w", pth, err2)
	}

	if err2 = f.Close(); err2 != nil {
		_ = os.Remove(pth)

		return fmt.Errorf("failed in closing file %s: %w", pth, err2)
	}

	return nil
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
// 其原有权限与属主/属组会被沿用（见下方实现），
// 因此显式传入的 perm 只对「文件原本不存在」的情况生效。
//
// 为什么要沿用属主：临时文件由当前进程创建，rename 之后目标文件的
// uid/gid 是当前进程用户。若被改写的是属于其它服务账号的配置或状态文件
// （例如以 root 运行、去改写属于 postfix 的文件），属主会被改掉，
// 可能导致那个服务读不到或写不了该文件。因此这里在替换前记录原属主，
// 并对临时文件 Chown 后再 rename。
//
// 沿用属主需要相应权限（非 root 只能 chown 给自己所属的组）；
// 若无权限则忽略该错误，此时行为退化为「只沿用权限、属主变为当前用户」。
//
// 关于符号链接：本函数用 os.Rename 替换目标路径，
// 因此若 pth 本身是符号链接，**链接会被替换成普通文件**，
// 而不会沿链接写入它指向的目标。这个行为是刻意的（可避免被预置链接劫持），
// 但调用方若期望“写入链接指向的真实文件”，需先自行解析链接（filepath.EvalSymlinks）。
func WriteFileAtomic(pth string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(pth)

	// 若目标文件已存在，沿用它的权限与属主，避免一次改写意外改变它们。
	//
	// 用 os.Lstat 而非 os.Stat：目标若是符号链接，则取到的是链接自身的权限
	// （符号链接权限恒为 0777，无实际意义），此时应保留调用方传入的 perm。
	var (
		keepOwner bool
		ownerUid  int
		ownerGid  int
	)

	if info, err := os.Lstat(pth); err == nil && info.Mode()&os.ModeSymlink == 0 {
		// 用 Mode() 而非 Perm()：后者会丢掉 setuid / setgid / sticky。
		perm = info.Mode()

		if ss, ok := info.Sys().(*syscall.Stat_t); ok {
			keepOwner = true
			ownerUid = int(ss.Uid)
			ownerGid = int(ss.Gid)
		}
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

	// 先把属主设好再写入：这样文件一出现就属于正确的账号，
	// 不会存在“短暂属于当前用户”的窗口。
	if keepOwner {
		if cerr := tmp.Chown(ownerUid, ownerGid); cerr != nil {
			// 无权限时忽略：保留权限位仍然有意义，
			// 属主退化为当前进程用户。
			_ = cerr
		}
	}

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

	// 再 fsync 父目录，使目录项（即上面的 rename 结果）也真正落盘。
	//
	// 为什么文件自身的 fsync 不够：fsync(文件) 只保证该文件的数据与元数据落盘，
	// 不保证「目录中新增/改名了一个条目」这件事落盘。断电时可能出现
	// 数据已写入磁盘、但目录项仍指向旧文件（甚至指向已删除的 inode）的情况，
	// 目标路径因此可能回退到旧内容或读不到文件。
	//
	// 目录 fsync 失败不视为写入失败：此时文件内容已正确替换，
	// 只是崩溃后的持久性保证打了折扣，不值得让调用方认为写入失败。
	if d, derr := os.Open(dir); derr == nil {
		_ = d.Sync()
		_ = d.Close()
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
//
// 目录权限的复制方式说明：MkdirAll 的 perm 会受 umask 影响，
// 直接用源目录权限调用 MkdirAll 得到的权限会被 umask 削减
// （例如源目录 0777、umask 022 时实际得到 0755），与原目录不一致。
// 因此这里先用 0700 创建，再用 Chmod 显式设置 —— Chmod 不受 umask 影响。
func copyDir(src, dst string) error {
	stat, err := os.Stat(src)
	if err != nil {
		return err
	}

	sys, ok := stat.Sys().(*syscall.Stat_t)
	if !ok {
		return fmt.Errorf("unsupported file info type for %s", src)
	}

	// 先用最小权限创建，随后再设置属主与权限。
	//
	// 顺序很重要：必须**先 Chown 再 Chmod**。
	// Linux 的 chown(2) 在属主/属组发生变化时会清除 setuid / setgid
	// （安全语义：换了主人就不该保留提权位），因此若先 Chmod 再 Chown，
	// 刚设好的 setuid / setgid 会被紧接着的 Chown 抹掉。
	// 实测（Linux 容器）：源目录带 setgid 时，
	//   Chmod → Chown 的结果是 setgid 丢失；Chown → Chmod 则保留。
	if err = os.MkdirAll(dst, 0700); err != nil {
		return fmt.Errorf("failed in creating directory %s: %w", dst, err)
	}

	if err = os.Chown(dst, int(sys.Uid), int(sys.Gid)); err != nil {
		return fmt.Errorf("failed in setting owner/group of %s: %w", dst, err)
	}

	// 这里必须用 stat.Mode() 而不是 stat.Mode().Perm()。
	// Perm() 只保留低 9 位（0777），会丢掉 setuid / setgid / sticky；
	// os.Chmod 会忽略 FileMode 中的类型位、正确处理这三个特殊位。
	// 共享目录（group-writable 的部署目录、sticky 的目录等）依赖这些位，
	// 丢掉后目标目录权限就与源目录不一致了。
	if err = os.Chmod(dst, stat.Mode()); err != nil {
		return fmt.Errorf("failed in setting permission of %s: %w", dst, err)
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		// 目录按目录递归；其余（普通文件、符号链接等）按文件复制。
		//
		// 注意：这里用 entry.IsDir()，它基于 ReadDir 返回的 DirEntry 类型，
		// 对符号链接返回 false（不会被误当成目录递归进去）。
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
//
// 安全要点：不使用 os.Create —— 后者会**跟随符号链接**：
// 若攻击者预先在 dst 处放置一个指向 /etc/sudoers 之类的符号链接，
// 复制操作就会把内容写进该目标文件，形成任意文件写。
//
// 因此不采用「打开已有路径再写入」的方式，而是：
//  1. 若 dst 已存在则先 os.Remove（作用于链接自身，不跟随）；
//  2. 再以 O_CREATE|O_EXCL|O_NOFOLLOW 新建。
//
// 这样既保持了「目标已存在时覆盖」的语义（os.Rename 因 EXDEV 失败后，
// copyDir 依赖它把源树复制到目标树；目标树中可能已有同名文件），
// 又不会顺着预置的符号链接写入。
//
// 写入过程中若发生错误，会删除已创建的 dst，避免留下半截文件；
// 因为创建使用了 O_EXCL，残留文件会让调用方在同一路径重试时永久失败，
// 所以函数签名保留了命名返回值 err（详见 defer 处的说明）。
func copyFile(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	stat, err := os.Stat(src)
	if err != nil {
		return err
	}

	// 目标已存在时先移除：os.Remove 作用于路径本身（不跟随符号链接），
	// 因此删除的是预置的链接而不是它指向的文件。
	//
	// 必须先删除再创建，而不能直接打开已有路径写入（os.Create 就是这样），
	// 因为那样会跟随符号链接。同时这一步也保证了「目标已存在时覆盖」的语义：
	// copyDir 在 os.Rename 因 EXDEV 失败后被调用，目标树中可能已有同名文件，
	// 此时应覆盖而不是报错，否则重复复制会失败。
	if _, err = os.Lstat(dst); err == nil {
		if err = os.Remove(dst); err != nil {
			return fmt.Errorf("failed in removing existing destination %s: %w", dst, err)
		}
	}

	// 初建权限先取 0600（最小权限），随后再按源文件权限修正 ——
	// 避免“创建到 chmod 之间”存在一段权限过宽的窗口。
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_EXCL|syscall.O_NOFOLLOW, 0600)
	if err != nil {
		return fmt.Errorf("failed in creating destination %s: %w", dst, err)
	}

	// 任何一步失败都删除已创建的 dst，避免留下半截文件。
	//
	// 这里判断的是**命名返回值 err**（函数签名中写的是 `(err error)`）。
	// return 语句会先把返回值赋给 err，再执行 defer，因此无论写成本函数
	// 末尾那种 `return fmt.Errorf(...)`（类型断言失败分支），
	// 还是 `err = ...; return err`，defer 都能看到错误。
	//
	// 反过来说，若把签名改回 `error`（err 变成局部变量），
	// 末尾那条 `return fmt.Errorf(...)` 就不会再赋值给 err，
	// defer 看到的仍是 nil，dst 便不会被删除；而创建使用了 O_EXCL，
	// 调用方在同一路径重试会永久失败。因此命名返回值是必要的，不要改掉。
	// 这一点由 TestCopyFileSignatureMustKeepNamedReturn 静态守护。
	defer func() {
		if err != nil {
			_ = os.Remove(dst)
		}
	}()

	if _, err = io.Copy(out, in); err != nil {
		_ = out.Close()

		return err
	}

	// Sync file content to disk
	if err = out.Sync(); err != nil {
		_ = out.Close()

		return err
	}

	if err = out.Close(); err != nil {
		return err
	}

	// 先复制属主/属组，再复制权限位。
	//
	// 顺序很重要：必须**先 Chown 再 Chmod**。
	// Linux 的 chown(2) 在属主/属组发生变化时会清除 setuid / setgid
	// （安全语义：换了主人就不该保留提权位），因此若先 Chmod 再 Chown，
	// 刚设好的 setuid / setgid 会被紧接着的 Chown 抹掉，这步修复就失效了。
	// 实测（Linux 容器）：源文件带 setgid 时，
	//   Chmod → Chown 的结果是 setgid 丢失；Chown → Chmod 则保留。
	sys, ok := stat.Sys().(*syscall.Stat_t)
	if !ok {
		return fmt.Errorf("failed in reading owner/group info of %s", src)
	}

	if err = os.Chown(dst, int(sys.Uid), int(sys.Gid)); err != nil {
		return err
	}

	// 这里必须用 stat.Mode() 而不是 stat.Mode().Perm()：
	// Perm() 只保留低 9 位，会丢掉 setuid / setgid / sticky。
	// 例如复制一份带 setgid 的可执行文件时，Perm() 会让目标失去该位。
	// os.Chmod 会忽略 FileMode 里的类型位，只应用权限位与这三个特殊位。
	if err = os.Chmod(dst, stat.Mode()); err != nil {
		return err
	}

	return nil
}

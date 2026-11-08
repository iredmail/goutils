# 代码形态示例（goutils）

goutils 是被多个项目依赖的公共工具库，API 的稳定性与可读性优先。本文件配合项目内
`AGENTS.md` 与全局 Go 规则使用：

- 新增/修改导出 API 前先读本文件；示例与代码不一致时以代码为准，并顺手修正示例。
- 示例多为节选；只收录稳定、已评审的写法，不是复制模板。
- 兼容性优先：已导出的签名、错误语义、nil/空切片约定不要随意改变。
- 新增/修改导出 API 时同步更新本文件。

## 1. 导出函数：短、文档注释、标准库优先

```go
// SplitLines 按行拆分字符串，并去除每行的首尾空白字符。
func SplitLines(s string) (lines []string) {
	s = strings.ReplaceAll(s, "\r\n", "\n")

	for i := range strings.SplitSeq(s, "\n") {
		lines = append(lines, strings.TrimSpace(i))
	}

	return
}
```

要点：导出符号必须有文档注释；优先标准库（`strings.SplitSeq`、`slices`、`maps`）；
返回 nil 还是空切片属于契约，改变时要评估所有调用方。

## 2. 类型断言：comma-ok，库代码不 panic

```go
// ErrIsDuplicate 检测 err 是否为插入数据重复。
func ErrIsDuplicate(err error) bool {
	e, ok := err.(*sqlite.Error)
	if !ok {
		return false
	}

	// (2067) SQLITE_CONSTRAINT_UNIQUE
	// (1555) SQLITE_CONSTRAINT_PRIMARYKEY
	return e.Code() == sqlitelib.SQLITE_CONSTRAINT_UNIQUE ||
		e.Code() == sqlitelib.SQLITE_CONSTRAINT_PRIMARYKEY
}
```

要点：错误/接口断言一律 comma-ok；库函数不 panic；注释写清魔法值来源。

## 3. 契约写进注释，错误保留错误链

```go
// GetFileStat 获取目标路径的信息（不跟随符号链接）。
//
// 路径不存在时返回零值 FileStat（Exists = false）且 err 为 nil，
// 调用方应以 fs.Exists 判断存在性。
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

	// 其余字段（IsLink/IsRegular/IsDir、属主/属组、uid/gid）省略，见源文件。

	return fs, nil
}
```

要点：完整文档注释还说明了 `IsRegular` 与特殊文件（FIFO/socket/设备）的判定规则，
修改文件相关工具函数前先读它；错误用 `%w` 保留错误链，上下文写清对象和操作。

## 4. 测试：testify + 边界值

```go
func TestFlattenStrings(t *testing.T) {
	var empty []string

	assert.Equal(t, empty, FlattenStrings(nil))
	assert.Equal(t, []string{"1"}, FlattenStrings("1"))
	assert.Equal(t, []string{"1", "2"}, FlattenStrings([]string{"1", "2", "", "1"}))
}
```

要点（节选）：所有导出函数都要有测试；覆盖 nil、空值、重复值、错误类型；断言用
testify 的 `assert` / `require`（库内已有依赖）。

## 5. 项目特有约束

- `ctxutils` 同时提供 fiber v2（`ctxutils`）和 fiber v3（`ctxutils/v3`）两套 API：
  改一处必须同步另一处，测试两边都跑。
- `logger` 定义小接口并给出多种实现：新代码面向 `logger.Logger` 接口，不依赖具体实现。
- 本库被 `ee`、`gosible`、`spiderd` 等项目直接依赖：改公开 API 前先搜所有使用方，
  并分别跑测试。

## 不要这样做

- 在工具函数里 `panic`、`os.Exit`、`log.Fatal`。
- 把业务逻辑、协议细节放进通用 utils；应放到对应项目。
- 忽略错误（`_ = err`）或用 `%v` 包装导致错误链丢失。
- 为一次性需求新增导出 API；先确认已有函数能否覆盖。

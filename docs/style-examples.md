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

## 6. 组合型 API：薄封装写清委托范围

对底层库做薄封装时，文档注释要写清哪些行为来自底层、哪些由调用方负责，避免把
底层语义误写成新契约。

```go
// HTTPHandler 返回处理 ACME http-01 挑战、并把其余请求交给 fallback 的
// http.Handler，必须由调用方监听在 80 端口；fallback 为 nil 时，其余请求默认
// 重定向到 HTTPS。
//
// 以 /.well-known/acme-challenge/ 开头的请求始终交给 autocert 处理，与 fallback
// 无关；固定证书等未启用 autocert 的情况下通常返回 404。
func (m *Manager) HTTPHandler(fallback http.Handler) http.Handler {
	return m.autocertMgr.HTTPHandler(fallback)
}
```

要点：薄封装优先保持底层语义，不额外发明行为；契约里写清"必须由谁监听端口"、
"哪些请求不经过 fallback"等边界，否则调用方容易误用。

## 7. 临时文件：所有权与清理责任写进注释

```go
// WriteTempFile 把 content 写入系统临时目录下的随机文件并返回其路径。
// 文件已关闭，权限为 0600；调用方负责在使用完毕后删除该文件。
func WriteTempFile(content []byte, pattern ...string) (pth string, err error)
```

要点：临时文件 API 返回路径而不是 `*os.File`，调用方无需再关闭；权限固定 0600；
"调用方负责删除"属于契约，必须写进文档注释，否则容易泄漏临时文件。`pattern` 为
可选参数，语义与 `os.CreateTemp` 一致（最后一个 `*` 会被替换为随机串）；只使用第一
个 pattern，未提供或为空时使用默认前缀 `go-temp-*`。

## 不要这样做

- 在工具函数里 `panic`、`os.Exit`、`log.Fatal`。
- 把业务逻辑、协议细节放进通用 utils；应放到对应项目。
- 忽略错误（`_ = err`）或用 `%v` 包装导致错误链丢失。
- 为一次性需求新增导出 API；先确认已有函数能否覆盖。

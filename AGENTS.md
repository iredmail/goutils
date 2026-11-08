# goutils 模块规则

本文件是 `goutils` 的模块级补充，与 `/Users/zhb/projects/go/AGENTS.md`（工作区规则）和
`~/.dsh/AGENTS.md`（全局规则）冲突时，以更具体、更安全的一条为准。

## 模块信息

- module：`github.com/iredmail/goutils`；Go 1.27.3。
- 公共工具库，被 `ee`、`gosible`、`spiderd` 等项目直接依赖，API 兼容性优先。
- 同时提供 fiber v2（`ctxutils`）与 fiber v3（`ctxutils/v3`）两套 API。
- 主要包：`ctxutils`、`dbutils`、`dnsutils`、`emailutils`、`logger`、`respcode`、
  `sqlutils`、`timeutils`、`sslcert`、`slice`、`smtpclient`、`i18n`、`htmlfunc`、`pwhash`。
- 测试：`testing` + testify。

## 命令

- 测试：`make test`（`go test -cover -count=1 -parallel=10 ./...`）
- 详细测试：`make testv`
- Lint：`golangci-lint run --config .golangci.yml`
- 格式化：`gofmt -w` + `goimports -w`
- 静态检查：`go vet ./...`

## 硬规则

- 导出 API 视为契约：签名、返回值语义（nil/空切片）、错误类型不要随意改变；必须破坏时
  先弃用（文档注释标 `Deprecated:`）并说明迁移方式。
- `ctxutils` 与 `ctxutils/v3` 必须同步修改，两边都跑测试。
- 工具函数不 `panic`、不 `os.Exit`、不 `log.Fatal`；错误用 `%w` 保留错误链。
- 导出符号必须有文档注释；非直观的返回值语义（例如“不存在时 err 为 nil”）写进注释。
- 类型断言一律 comma-ok；尽量不用 `reflect`，确需使用必须有测试覆盖。
- 不把业务逻辑、协议细节放进通用工具包；应放到对应项目。
- 新增导出函数前先确认已有函数能否覆盖；不为一次性需求新增 API。
- 新增/升级依赖前先说明理由并确认；优先标准库。
- 本库被多个项目依赖：改公开 API 后，至少跑 `ee`、`gosible`、`spiderd` 的相关测试。

## 参考

- `docs/style-examples.md`：推荐写法的代码示例。
- `README.md`：各包的用途说明。

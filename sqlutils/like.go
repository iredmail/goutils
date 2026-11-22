package sqlutils

import (
	"fmt"
	"strings"

	"github.com/doug-martin/goqu/v9"
)

// EscapeLikeLiteral 将字面值中的 `!`、`%` 和 `_` 转义，供 SQL LIKE 模式使用。
// 返回值必须搭配 `ESCAPE '!'`。调用方应在转义后添加有意使用的 `%` 通配符，例如：
//
//	pattern := "%@" + EscapeLikeLiteral(domain)
//	goqu.L("? LIKE ? ESCAPE '!'", goqu.I("email"), pattern)        // PostgreSQL、SQLite
//	goqu.L("? LIKE BINARY ? ESCAPE '!'", goqu.I("email"), pattern) // MySQL
//
// MySQL 示例保留 goqu `.Like()` 的 `LIKE BINARY` 语义。改用普通 `LIKE` 前，必须确认
// 列 collation 和大小写匹配行为的变化。本函数只处理 LIKE 模式，不代替 SQL 参数绑定。
func EscapeLikeLiteral(value string) string {
	return strings.NewReplacer("!", "!!", "%", "!%", "_", "!_").Replace(value)
}

// LikeMatchMode 指定字面值在待匹配文本中的位置。
type LikeMatchMode uint8

const (
	// LikeMatchExact 仅匹配完整的字面值，不添加通配符。
	LikeMatchExact LikeMatchMode = iota
	// LikeMatchPrefix 匹配以字面值开头的文本，在字面值后添加 `%`。
	LikeMatchPrefix
	// LikeMatchSuffix 匹配以字面值结尾的文本，在字面值前添加 `%`。
	LikeMatchSuffix
	// LikeMatchContains 匹配包含字面值的文本，在字面值两侧添加 `%`。
	LikeMatchContains
)

// LikeOptions 控制 LIKE 的匹配方式。
type LikeOptions struct {
	// Match 指定字面值在待匹配文本中的位置，默认为 LikeMatchExact。
	Match LikeMatchMode

	// StarAsWildcard 为 true 时，将 value 中的每个 `*` 作为匹配任意字符序列的
	// 通配符；为 false 时，`*` 按字面匹配。
	//
	// 例如匹配 192.168.0.0/16 网段内的地址：
	//
	//	expr, err := Like(
	//		"mysql",
	//		goqu.I("address"),
	//		"192.168.*/*",
	//		LikeOptions{Match: LikeMatchExact, StarAsWildcard: true},
	//	)
	//
	// 生成的 LIKE 模式为 "192.168.%/%"。
	StarAsWildcard bool

	// CaseInsensitive 为 true 时，对 column 使用 SQL LOWER，并在 Go 中将
	// value 转换为小写；为 false 时保留数据库默认的大小写和 collation 行为。
	CaseInsensitive bool
}

// Like 根据 SQL dialect 构造带 `ESCAPE '!'` 的 LIKE 表达式。
// 支持 mysql、mariadb、postgres、pgsql、sqlite、sqlite3，其他 dialect 返回错误。
// MySQL/MariaDB 使用 LIKE BINARY，其余数据库使用 LIKE。
//
// column 可以传入 goqu.I、goqu.Func 等表达式，以支持表别名、函数以及
// PostgreSQL bytea 转换。value 默认按字面匹配，通过 LikeOptions 控制通配符
// 和大小写行为。本函数不代替 SQL 参数绑定。
//
// 匹配域名后缀：
//
//	expr, err := Like(
//		"mysql",
//		goqu.I("policy_name"),
//		"@"+domain,
//		LikeOptions{Match: LikeMatchSuffix},
//	)
//
// 在 SQLite 中进行大小写不敏感的包含匹配：
//
//	expr, err := Like(
//		"sqlite3",
//		goqu.I("description"),
//		keyword,
//		LikeOptions{Match: LikeMatchContains, CaseInsensitive: true},
//	)
//
// 将 `*` 用作通配符：
//
//	expr, err := Like(
//		"mysql",
//		goqu.I("address"),
//		"192.168.*/*",
//		LikeOptions{StarAsWildcard: true},
//	)
//
// 匹配 PostgreSQL bytea 列：
//
//	column := goqu.Func("convert_from", goqu.I("email"), "UTF8")
//	expr, err := Like(
//		"pgsql",
//		column,
//		"@"+domain,
//		LikeOptions{Match: LikeMatchSuffix},
//	)
func Like(dialect string, column goqu.Expression, value string, options LikeOptions) (goqu.Expression, error) {
	if column == nil {
		return nil, fmt.Errorf("build LIKE expression: column is required")
	}

	var operator string
	switch strings.ToLower(strings.TrimSpace(dialect)) {
	case "mysql", "mariadb":
		operator = "LIKE BINARY"
	case "postgres", "pgsql", "sqlite", "sqlite3":
		operator = "LIKE"
	default:
		return nil, fmt.Errorf("build LIKE expression: unsupported dialect %q", dialect)
	}

	pattern, err := buildLikePattern(value, options)
	if err != nil {
		return nil, fmt.Errorf("build LIKE expression: %w", err)
	}

	if options.CaseInsensitive {
		column = goqu.Func("LOWER", column)
	}

	return goqu.L("? "+operator+" ? ESCAPE '!'", column, pattern), nil
}

func buildLikePattern(value string, options LikeOptions) (string, error) {
	if options.CaseInsensitive {
		value = strings.ToLower(value)
	}

	pattern := EscapeLikeLiteral(value)
	if options.StarAsWildcard {
		pattern = strings.ReplaceAll(pattern, "*", "%")
	}

	switch options.Match {
	case LikeMatchExact:
	case LikeMatchPrefix:
		pattern += "%"
	case LikeMatchSuffix:
		pattern = "%" + pattern
	case LikeMatchContains:
		pattern = "%" + pattern + "%"
	default:
		return "", fmt.Errorf("unsupported match mode %d", options.Match)
	}

	return pattern, nil
}

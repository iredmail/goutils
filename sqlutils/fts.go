package sqlutils

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/doug-martin/goqu/v9"
)

// 参考文档：https://www.sqlite.org/fts5.html

const substituteChar rune = 0x1A

// 参考：https://www.sqlite.org/fts5.html#fts5_strings
func isFTS5BarewordRune(r rune) bool {
	if r > 0x7F {
		return true
	}
	if r == substituteChar {
		return true
	}
	if r == '_' {
		return true
	}
	if unicode.IsLetter(r) || unicode.IsDigit(r) {
		return true
	}
	return false
}

// escapeFTSKeyword 会把输入当作一组“字面 token”来处理，避免被 FTS5 解释成查询语法。
// 处理方式很简单：先按空白分词，再把每个 token 直接输出或整体加引号。
// - 纯 bareword 原样保留；
// - `AND`、`OR`、`NOT` 这三个词强制加引号，避免被当成布尔操作符；
// - 含 `@`、`.`、`"`、`;` 之类标点的 token，整体作为一个短语字面量加引号。
// 这样 `u@x.io` 这类邮箱地址会作为完整内容参与匹配，而不是被拆散。
func escapeFTSKeyword(s string) string {
	tokens := strings.Fields(strings.TrimSpace(s))
	if len(tokens) == 0 {
		return ""
	}

	var b strings.Builder
	for i, token := range tokens {
		if i > 0 {
			b.WriteByte(' ')
		}

		if token == "AND" || token == "OR" || token == "NOT" {
			b.WriteByte('"')
			b.WriteString(token)
			b.WriteByte('"')
			continue
		}

		bareword := true
		for _, r := range token {
			if !isFTS5BarewordRune(r) {
				bareword = false
				break
			}
		}

		if bareword {
			b.WriteString(token)
			continue
		}

		b.WriteByte('"')
		b.WriteString(strings.ReplaceAll(token, `"`, `""`))
		b.WriteByte('"')
	}

	return b.String()
}

// FTSSearch 先查询 fts 表，再查询目标表，将目标表里匹配的记录赋值给 inputRows 并返回。
//
// 注意：
//   - ftsTable 是使用 SQLite 的 fts5 扩展通过 external content 方式创建的 virtual table。
//   - destTable 必须有一个 UNIQUE INDEX 的字段 `id`。后期再考虑支持指定这个唯一字段。
//   - inputRows 必须是一个结构体实例的指针，用于映射 destTable 里的 SQL 记录。
func FTSSearch(gdb *goqu.Database, keyword, ftsTable, destTable string, inputRows any) (err error) {
	var rowIDs []int64

	escapedKeyword := escapeFTSKeyword(keyword)
	if escapedKeyword == "" {
		return
	}

	err = gdb.
		From(ftsTable).
		Select("rowid").
		Where(goqu.L(fmt.Sprintf("%s MATCH ?", ftsTable), escapedKeyword)).
		Prepared(true).
		ScanVals(&rowIDs)

	if err != nil {
		return
	}

	if len(rowIDs) == 0 {
		return
	}

	err = gdb.
		From(destTable).
		Where(goqu.Ex{"id": rowIDs}).
		ScanStructs(inputRows)

	return
}

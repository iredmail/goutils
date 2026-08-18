package sqlutils

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/doug-martin/goqu/v9"
)

// 参考文档：https://www.sqlite.org/fts5.html

const substituteChar rune = 0x1A

var (
	ftsBarewordKeyword = regexp.MustCompile(`(^|\s)(AND|OR|NOT)(\s|$)`)
	ftsSpaceCollapser  = regexp.MustCompile(`\s+`)
)

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

// escapeFTSKeyword 按照 SQLite FTS5 "3.1. FTS5 Strings" 的规则转义查询关键字，
// 确保输入的 keyword 全部被当作「字面内容」参与匹配，不会被解释为 FTS5 查询语法元素。
// 官方文档：https://www.sqlite.org/fts5.html#fts5_strings
//
// 具体规则：
//  1. 空白字符合并为单个空格并修剪首尾空白。
//  2. 按字符扫描：
//     - 属于 FTS5 bareword 的字符（>U+007F 的非 ASCII、A-Za-z、0-9、下划线 `_`、U+001A SUB）原样保留；
//     - 空白字符原样保留（作为 term 分隔符）；
//     - 双引号 `"` 写为 `""""`（外层为短语引号，内层 SQL-style 转义后的 `""`）；
//     - 其他所有字符都用双引号包围成短语，例如 `*` 写成 `"*"`。
//  3. 最后按 FTS5 "case sensitive" 的 bareword 禁用列表处理：
//     当 `AND`、`OR`、`NOT`（必须完全匹配大小写、且被空白或字符串边界包围）
//     作为独立 term 出现时，用双引号包裹成 `"AND"` / `"OR"` / `"NOT"`，
//     使它们作为普通单词而不是布尔操作符。注意 `and` / `or` / `not`（小写）
//     本身就是合法 bareword，不需要处理；`NEAR` 也不是 bareword 禁用词，不需处理。
func escapeFTSKeyword(s string) string {
	s = ftsSpaceCollapser.ReplaceAllString(s, " ")
	s = strings.TrimSpace(s)

	var b strings.Builder
	for _, r := range s {
		if isFTS5BarewordRune(r) || unicode.IsSpace(r) {
			b.WriteRune(r)
		} else if r == '"' {
			b.WriteString(`""""`)
		} else {
			b.WriteByte('"')
			b.WriteRune(r)
			b.WriteByte('"')
		}
	}

	result := b.String()
	for ftsBarewordKeyword.MatchString(result) {
		result = ftsBarewordKeyword.ReplaceAllStringFunc(result, func(m string) string {
			sub := ftsBarewordKeyword.FindStringSubmatch(m)
			return sub[1] + `"` + sub[2] + `"` + sub[3]
		})
	}
	return result
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

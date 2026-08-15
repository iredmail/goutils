package sqlutils

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/doug-martin/goqu/v9"
)

var (
	// 移除 fts5 关键字（不区分大小写）：AND, OR, NOT, NEAR。
	ftsKeywordRemover = regexp.MustCompile(`(?i)\b(AND|OR|NOT|NEAR)\b`)
	ftsSpaceCollapser = regexp.MustCompile(`\s+`)
)

// escapeFTSKeyword 转义 SQLite FTS 查询关键字中的特殊字符。包括：
// - 移除 fts5 关键字（不区分大小写）：AND, OR, NOT, NEAR。
// - 除了字母和数字之外的所有字符都用双引号进行包裹，例如：`a*b` 改为 `a"*"b`。
//
// 参考：https://www.sqlite.org/fts5.html#fts5_strings
func escapeFTSKeyword(s string) string {
	s = ftsKeywordRemover.ReplaceAllString(s, "")
	// 移除连续的多个空白字符，只保留一个
	s = ftsSpaceCollapser.ReplaceAllString(s, " ")
	s = strings.TrimSpace(s)

	// 除了字母和数字之外的所有字符都用双引号进行包裹，例如：`a*b` 要改为 `a"*"b`。
	var b strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		} else if r == '"' {
			b.WriteString(`""""`)
		} else {
			b.WriteByte('"')
			b.WriteRune(r)
			b.WriteByte('"')
		}
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

	fmt.Printf("DEBUG old keyword: %q\n", keyword)
	escapedKeyword := escapeFTSKeyword(keyword)
	if escapedKeyword == "" {
		return
	}
	fmt.Printf("DEBUG escaped keyword: %q\n", escapedKeyword)

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

package sqlutils

import (
	"fmt"

	"github.com/doug-martin/goqu/v9"
)

// FTSSearch 先查询 fts 表，再查询目标表，将目标表里匹配的记录赋值给 inputRows 并返回。
//
// 注意：
//   - ftsTable 是使用 SQLite 的 fts5 扩展通过 external content 方式创建的 virtual table。
//   - destTable 必须有一个 UNIQUE INDEX 的字段 `id`。后期再考虑支持指定这个唯一字段。
//   - inputRows 必须是一个结构体实例的指针，用于映射 destTable 里的 SQL 记录。
func FTSSearch(gdb *goqu.Database, keyword, ftsTable, destTable string, inputRows any) (err error) {
	var rowIDs []int64

	// 注意：SQLite FTS 查询关键字如果带有特殊字符或 SQLite 的保留字符，例如 `.`, `@` 等，必须将其用双引号包含起来。
	// 例如：`SELECT * FROM xxx WHERE xxx MATCH '"u@x.io"';`
	err = gdb.
		From(ftsTable).
		Select("rowid").
		Where(goqu.L(fmt.Sprintf(`%s MATCH '"%s"'`, ftsTable, keyword))).
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

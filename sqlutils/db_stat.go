package sqlutils

import (
	"strings"

	"github.com/doug-martin/goqu/v9"
)

type DBStat struct {
	Path string // 数据库名称或文件路径

	Size int64 // PageSize * PageCount

	// Pragma
	PageSize      int64
	PageCount     int64
	JournalMode   string // 日志模式
	AutoVacuum    string
	Synchronous   string
	FreelistCount int64 // number of unused pages in the database file
}

type TableStat struct {
	Name string `db:"name"`       // Table name.
	Size int64  `db:"table_size"` // Table size in bytes
}

// GetDBStat 返回指定数据库的相关信息。
func GetDBStat(pth string, db *goqu.Database) (stat DBStat) {
	var pageSize, pageCount, freelistCount int64
	var journalMode, autoVacuum, synchronous string

	_ = db.QueryRow("PRAGMA page_size").Scan(&pageSize)
	_ = db.QueryRow("PRAGMA page_count").Scan(&pageCount)
	_ = db.QueryRow("PRAGMA freelist_count").Scan(&freelistCount)

	_ = db.QueryRow("PRAGMA journal_mode").Scan(&journalMode)
	_ = db.QueryRow("PRAGMA auto_vacuum").Scan(&autoVacuum)
	switch autoVacuum {
	case "0":
		autoVacuum = "NONE"
	case "1":
		autoVacuum = "FULL"
	case "2":
		autoVacuum = "INCREMENTAL"
	}

	// 0 | OFF | 1 | NORMAL | 2 | FULL | 3 | EXTRA;
	_ = db.QueryRow("PRAGMA synchronous").Scan(&synchronous)
	switch synchronous {
	case "0":
		synchronous = "OFF"
	case "1":
		synchronous = "NORMAL"
	case "2":
		synchronous = "FULL"
	case "3":
		synchronous = "EXTRA"
	}

	return DBStat{
		Path: pth,

		Size: pageSize * pageCount,

		PageCount:     pageCount,
		PageSize:      pageSize,
		JournalMode:   strings.ToUpper(journalMode),
		AutoVacuum:    strings.ToUpper(autoVacuum),
		Synchronous:   strings.ToUpper(synchronous),
		FreelistCount: freelistCount,
	}
}

// GetTableStats 返回数据库中所有表的大小，包含索引。
// FYI <https://www.sqlite.org/dbstat.html>
// WARNING: Go 的 sqlite driver 必须支持 SQLITE_ENABLE_DBSTAT_VTAB 才能访问 `dbstat` 表。
func GetTableStats(db *goqu.Database) (rows []TableStat) {
	_ = db.
		Select("name", goqu.L("SUM(pgsize) table_size")).
		From(goqu.L("dbstat")).
		GroupBy("name").
		Order(goqu.C("table_size").Desc()).
		ScanStructs(&rows)

	return
}

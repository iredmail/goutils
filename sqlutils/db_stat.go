package sqlutils

import (
	"database/sql"
	"strings"

	"github.com/doug-martin/goqu/v9"
)

type DBStat struct {
	Path string `json:"path"` // 数据库名称或文件路径

	Size int64 `json:"size"` // PageSize * PageCount

	// Pragma
	PageSize      int64  `json:"page_size"`
	PageCount     int64  `json:"page_count"`
	JournalMode   string `json:"journal_mode"` // 日志模式
	AutoVacuum    string `json:"auto_vacuum"`
	Synchronous   string `json:"synchronous"`
	FreelistCount int64  `json:"freelist_count"` // number of unused pages in the database file
}

// GetSqliteDBStat 返回指定 sqlite 数据库的相关信息。
func GetSqliteDBStat(pth string, db *sql.DB) (stat DBStat) {
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

type TableStat struct {
	Name string `db:"name"`       // Table name.
	Size int64  `db:"table_size"` // Table size in bytes
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

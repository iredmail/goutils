package sqlutils

import (
	"strings"

	"github.com/doug-martin/goqu/v9"
)

type DBStat struct {
	Path string // 数据库名称或文件路径
	Size int64

	// Pragma
	JournalMode string // 日志模式
	AutoVacuum  string
	Synchronous string
}

// GetDBStat 返回指定数据库的相关信息。
func GetDBStat(pth string, db *goqu.Database) (stat DBStat) {
	var pageSize, pageCount int64
	var journalMode, autoVacuum, synchronous string

	_ = db.QueryRow("PRAGMA page_size").Scan(&pageSize)
	_ = db.QueryRow("PRAGMA page_count").Scan(&pageCount)
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

		JournalMode: strings.ToUpper(journalMode),
		AutoVacuum:  strings.ToUpper(autoVacuum),
		Synchronous: strings.ToUpper(synchronous),
	}
}
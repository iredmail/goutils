package sqlutils

import (
	"github.com/doug-martin/goqu/v9"
)

type DBStat struct {
	Path string // 数据库名称或文件路径
	Size int64

	// Pragma
	JournalMode string // 日志模式
}

// GetDBStat 返回指定数据库的相关信息。
func GetDBStat(pth string, db *goqu.Database) (stat DBStat) {
	var pageSize, pageCount int64
	var journalMode string

	_ = db.QueryRow("PRAGMA page_size").Scan(&pageSize)
	_ = db.QueryRow("PRAGMA page_count").Scan(&pageCount)
	_ = db.QueryRow("PRAGMA journal_mode").Scan(&journalMode)

	return DBStat{
		Path:        pth,
		Size:        pageSize * pageCount,
		JournalMode: journalMode,
	}
}

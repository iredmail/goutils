package sqlutils

import (
	"database/sql"
	"fmt"
	"time"

	// 注册的 driver name 是 `sqlite`，不是 `sqlite3`。
	_ "modernc.org/sqlite" // database/sql driver
)

var (
	// SQLitePragmas 定义打开 SQLite 数据库时的 pragma 参数。
	// 参考：
	// https://www.sqlite.org/pragma.html
	// https://phiresky.github.io/blog/2020/sqlite-performance-tuning/
	SQLitePragmas = [][2]string{
		{"busy_timeout", "10000"},
		{"auto_vacuum", "full"},
		{"journal_mode", "WAL"},
		// WAL mode is always consistent with synchronous=NORMAL, but WAL mode
		// does lose durability. A transaction committed in WAL mode with
		// synchronous=NORMAL might roll back following a power loss or system
		// crash. Transactions are durable across application crashes
		// regardless of the synchronous setting or journal mode. The
		// synchronous=NORMAL setting is a good choice for most applications running in WAL mode.
		// {"schema.synchronous", "full"},
	}
)

// InitSQLiteDB 初始化 pth 参数指定的 SQLite 数据库。
//
//   - pth 指定数据库文件的路径。
//   - pragma 如果是 nil，则使用默认的 pragmas。
//   - maxIdleConns 指定打开数据库时的最大空闲连接数。如果为 0 表示使用默认值（50）。
//   - connMaxLifetime 指定连接的最大存活时间。如果为 0 表示使用默认值（10 分钟）。
func InitSQLiteDB(pth string, pragmas [][2]string, maxIdleConns int, connMaxLifetime time.Duration) (sqliteDB *sql.DB, err error) {
	if len(pragmas) > 0 {
		pth = pth + "?" + GenSQLiteURIPragmas(pragmas)
	}

	sqliteDB, err = sql.Open("sqlite", pth)
	if err != nil {
		return nil, fmt.Errorf("failed in open SQLite database: %s, %v", pth, err)
	}

	// 避免 `database is locked (5) (SQLITE_BUSY)` 错误。
	sqliteDB.SetMaxOpenConns(1)

	if maxIdleConns == 0 {
		maxIdleConns = 50
	}
	sqliteDB.SetMaxIdleConns(maxIdleConns)

	if connMaxLifetime == 0 {
		connMaxLifetime = 10 * time.Minute
	}
	sqliteDB.SetConnMaxLifetime(connMaxLifetime)

	return sqliteDB, nil
}

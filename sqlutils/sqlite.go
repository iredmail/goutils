package sqlutils

import (
	"database/sql"
	"fmt"
	"io/fs"
	"strings"

	"spider/internal/cfg"
	"spider/internal/logger"

	"github.com/doug-martin/goqu/v9"
	"modernc.org/sqlite"
)

const (
	// 内部使用的数据库，用于记录一些系统信息，如当前 SQL 表结构版本。
	tableSystem = "system"

	// keySQLSchemaVersion 记录数据库版本的 key
	keySQLSchemaVersion = "sql_schema_version"
)

func InitSQLiteDB(pth string, pragmas [][2]string) (sqliteDB *sql.DB, err error) {
	if len(pragmas) == 0 {
		pragmas = cfg.SQLitePragmas
	}

	uri := pth + "?" + GenSQLiteURIPragmas(pragmas)

	sqliteDB, err = sql.Open(cfg.SQLiteDriverName, uri)
	if err != nil {
		return nil, fmt.Errorf("failed in open sqlite db: %s, %v", pth, err)
	}

	// 避免 `database is locked (5) (SQLITE_BUSY)` 错误。
	sqliteDB.SetMaxOpenConns(cfg.DefaultSQLiteMaxOpenConns)
	sqliteDB.SetMaxIdleConns(cfg.DefaultSQLiteMaxIdleConns)
	sqliteDB.SetConnMaxLifetime(cfg.DefaultSQLiteConnMaxLifetime)

	return sqliteDB, nil
}

// ErrIsDuplicate 检测 err 是否为插入数据重复
// document: https://www.sqlite.org/rescode.html
func ErrIsDuplicate(err error) bool {
	e, ok := err.(*sqlite.Error)
	if !ok {
		return ok
	}

	// (2067) SQLITE_CONSTRAINT_UNIQUE
	// (1555) SQLITE_CONSTRAINT_PRIMARYKEY
	return e.Code() == 2067 ||
		e.Code() == 1555
}

func GenSQLiteURIPragmas(pragmas [][2]string) string {
	if len(pragmas) == 0 {
		return ""
	}

	var params []string
	for _, p := range pragmas {
		params = append(params, "_pragma="+p[0]+"%3d"+p[1]) // 以 `%3d` 代替 `=`
	}

	return strings.Join(params, "&")
}

func HasTable(gdb *goqu.Database, table string) (bool, error) {
	count, err := gdb.From("sqlite_master").
		Select("name").
		Where(goqu.Ex{
			"type": "table",
			"name": table,
		}).
		Count()

	return count == 1, err
}

func HasSystemTable(gdb *goqu.Database) (bool, error) {
	return HasTable(gdb, tableSystem)
}

// InsertSQLSchemaVersion 插入当前的 SQL 表结构版本
func InsertSQLSchemaVersion(gdb *goqu.Database, version int) (err error) {
	_, err = gdb.Insert(tableSystem).
		Prepared(true).
		Rows(goqu.Record{
			"k": keySQLSchemaVersion,
			"v": version,
		}).
		OnConflict(goqu.DoNothing()).
		Executor().Exec()

	return err
}

// getSQLSchemaVersion 获取当前数据库结构版本
func getSQLSchemaVersion(gdb *goqu.Database) (found bool, version int, err error) {
	var kv KVInt

	found, err = gdb.From(tableSystem).
		Where(goqu.Ex{"k": keySQLSchemaVersion}).
		Limit(1).
		ScanStruct(&kv)

	return found, kv.V, err
}

// updateSQLSchemaVersion 更新本地版本
func updateSQLSchemaVersion(gdb *goqu.Database, version int) error {
	_, err := gdb.
		Update(tableSystem).
		Where(goqu.Ex{"k": keySQLSchemaVersion}).
		Set(goqu.Record{"v": version}).
		Executor().Exec()

	return err
}

// UpgradeSQLSchema 升级 sql 表结构
//
// - `subFSSQLFiles` 是使用 fs.Sub 方法提取需要升级的 sql 文件所在的子目录。
func UpgradeSQLSchema(dbName string, gdb *goqu.Database, subFSSQLFiles fs.FS, latestVersion int) error {
	hasTable, err := HasSystemTable(gdb)
	if err != nil {
		return err
	}

	if !hasTable {
		// 初始安装，数据库尚未初始化，没有 system 表。
		return nil
	}

	// 获取本地表结构版本
	found, localVersion, err := getSQLSchemaVersion(gdb)
	if err != nil {
		return err
	}

	// 未找到版本信息则不更新。
	// 正常情况是必须有一个版本号，但存在表却不存在版本号记录的情况太特殊，暂不做处理。
	if !found {
		return nil
	}

	if localVersion >= latestVersion {
		return nil
	}

	for i := localVersion; i < latestVersion; i++ {
		newVersion := i + 1
		pth := fmt.Sprintf("%d.sql", newVersion)

		sqlRaw, err := fs.ReadFile(subFSSQLFiles, pth)
		if err != nil {
			logger.Error("[SQL] Failed in reading sql update file: database=%s, file=%s, error=%v", dbName, pth, err)

			return err
		}

		if _, err = gdb.Exec(string(sqlRaw)); err != nil {
			logger.Error("[SQL] Failed in updating sql schema: %d, %v", newVersion, err)

			return err
		}

		// 立即更新本地版本
		if err = updateSQLSchemaVersion(gdb, newVersion); err != nil {
			return err
		}
	}

	return nil
}

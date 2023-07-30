package sqlutils

import (
	"fmt"
	"io/fs"

	"github.com/doug-martin/goqu/v9"
)

func hasSystemTable(gdb *goqu.Database, dbName string) (bool, error) {
	dialect := gdb.Dialect()
	var sd *goqu.SelectDataset
	switch dialect {
	case dialectSQLite, dialectSQLite3:
		sd = gdb.From("sqlite_master").
			Select("name").
			Where(goqu.Ex{
				"type": "table",
				"name": tableSystem,
			})
	case dialectMysql:
		sd = gdb.From("information_schema.TABLES").
			Where(goqu.Ex{"table_schema": dbName, "table_name": tableSystem})
	case dialectPostgres:
		sd = gdb.From("pg_class").
			Where(goqu.Ex{"relname": tableSystem})
	default:
		return false, fmt.Errorf("unsupported dialect type: %s", dialect)
	}

	count, err := sd.Count()

	return count == 1, err
}

func createSystemTable(gdb *goqu.Database) error {
	dialect := gdb.Dialect()
	var exec string
	switch dialect {
	case dialectSQLite:
		exec = schemaSystemSqlite
	case dialectMysql:
		exec = schemaSystemMysql
	case dialectPostgres:
		exec = schemaSystemPostgres
	}

	_, err := gdb.Exec(exec)

	return err
}

func insertSQLSchemaVersion(gdb *goqu.Database, version int) error {
	_, err := gdb.Insert(tableSystem).
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

// InsertSQLSchemaVersion 插入当前的 SQL 表结构版本
func InsertSQLSchemaVersion(gdb *goqu.Database, version int) (err error) {
	_, err = gdb.
		Insert(tableSystem).
		Prepared(true).
		Rows(goqu.Record{
			"k": keySQLSchemaVersion,
			"v": version,
		}).
		OnConflict(goqu.DoNothing()).
		Executor().Exec()

	return err
}

// UpgradeSQLSchema 升级 sql 表结构
//
// - `subFSSQLFiles` 是使用 fs.Sub 方法提取需要升级的 sql 文件所在的子目录。
func UpgradeSQLSchema(dbName string, gdb *goqu.Database, subFSSQLFiles fs.FS, latestVersion int) error {
	hasTable, err := hasSystemTable(gdb, dbName)
	if err != nil {
		return err
	}

	// 初始安装，数据库尚未初始化，没有 system 表。
	if !hasTable {
		if err = createSystemTable(gdb); err != nil {
			return err
		}

		return insertSQLSchemaVersion(gdb, latestVersion)
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

	if subFSSQLFiles == nil {
		return nil
	}

	for i := localVersion; i < latestVersion; i++ {
		newVersion := i + 1
		pth := fmt.Sprintf("%d.sql", newVersion)

		sqlRaw, err := fs.ReadFile(subFSSQLFiles, pth)
		if err != nil {
			return err
		}

		if _, err = gdb.Exec(string(sqlRaw)); err != nil {
			return err
		}

		// 立即更新本地版本
		if err = updateSQLSchemaVersion(gdb, newVersion); err != nil {
			return err
		}
	}

	return nil
}

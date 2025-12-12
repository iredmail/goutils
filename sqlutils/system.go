package sqlutils

const (
	// 内部使用的数据库，用于记录一些系统信息，如当前 SQL 表结构版本。
	tableSystem = "system"

	// keySQLSchemaVersion 记录数据库版本的 key
	keySQLSchemaVersion = "sql_schema_version"

	// https://github.com/doug-martin/goqu/blob/master/docs/dialect.md#dialect
	dialectSQLite   = "sqlite"
	dialectSQLite3  = "sqlite3"
	dialectMysql    = "mysql"
	dialectPostgres = "postgres"

	schemaSystemSqlite = `
CREATE TABLE IF NOT EXISTS system (
    id  INTEGER PRIMARY KEY AUTOINCREMENT,
    k   TEXT NOT NULL,
    v   TEXT NOT NULL
) STRICT;

CREATE UNIQUE INDEX IF NOT EXISTS idx_system_k ON system (k);
`

	// MySQL 要求字段名要加反引号，因此合并到一行。
	schemaSystemMysql = "CREATE TABLE IF NOT EXISTS `system` (`id` INT(10) UNSIGNED AUTO_INCREMENT, `k` VARCHAR(255) NOT NULL, `v` VARCHAR(255) NOT NULL, PRIMARY KEY (`id`), UNIQUE INDEX idx_system_k (`k`)) ENGINE=InnoDB;"

	schemaSystemPostgres = `
CREATE TABLE IF NOT EXISTS system (
    id  SERIAL PRIMARY KEY,
    k   VARCHAR(255) NOT NULL,
    v   VARCHAR(255) NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_system_k ON system (k);
`
)

// SQL 表 `system` 保存 key-value 格式的值，这里针对不同数据类型的 value 定义结构体，方便
// SQL 查询时利用 goqu 做自动转换。

type KVInt struct {
	K string
	V int
}

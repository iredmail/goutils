package dbutils

import (
	"fmt"
)

const (
	DBTypeMariaDB  DBType = "mariadb"
	DBTypeMySQL    DBType = "mysql"
	DBTypePGSQL    DBType = "pgsql"
	DBTypeOpenLDAP DBType = "openldap"

	DialectMySQL = "mysql"
	DialectPg    = "postgres"
)

type DBType string

func (t DBType) String() string {
	return string(t)
}

func (t DBType) Dialect() (dialect string) {
	switch t {
	case DBTypeMariaDB, DBTypeMySQL:
		dialect = DialectMySQL
	case DBTypePGSQL:
		dialect = DialectPg
	default:
		// 默认使用 MySQL。
		// iRedMail 服务器在使用 OpenLDAP 作为 backend 的同时也安装了 MySQL/MariaDB。
		dialect = DialectMySQL
	}

	return
}

func ErrUnsupportedDBType(dt DBType) error {
	return fmt.Errorf("unsupported db type: %s", dt.String())
}

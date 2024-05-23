package dbutils

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

func DBTypeToDialect(t DBType) (dialect string) {
	switch t {
	case DBTypePGSQL:
		dialect = DialectPg
	default:
		dialect = DialectMySQL
	}

	return
}

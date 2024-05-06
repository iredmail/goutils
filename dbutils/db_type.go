package dbutils

const (
	DBTypeMariaDB  DBType = "mariadb"
	DBTypeMySQL    DBType = "mysql"
	DBTypePGSQL    DBType = "pgsql"
	DBTypeOpenLDAP DBType = "openldap"
)

type DBType string

func (t DBType) String() string {
	return string(t)
}

package dbutils

const (
	DBTypeMariadb  DBType = "mariadb"
	DBTypePGSQL    DBType = "pgsql"
	DBTypeOpenLDAP DBType = "openldap"
)

type DBType string

func (dt DBType) String() string {
	return string(dt)
}

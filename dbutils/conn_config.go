package dbutils

type ConnConfig struct {
	DBType DBType

	SQLConnConfig
	LDAPConnConfig
}

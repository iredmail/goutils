package dbutils

import (
	"time"
)

const (
	DBTypeMariadb DBType = "mariadb"
	DBTypePgSQL   DBType = "pgsql"
	DBTypeLDAP    DBType = "openldap"
)

type DBConfig struct {
	Type DBType
	SQLConfig
	LDAPConfig
}

type DBType string

func (dt DBType) String() string {
	return string(dt)
}

type SQLConfig struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	// For SQL driver
	MaxLifetime  time.Duration
	MaxOpenConns int
	MaxIdleConns int
}

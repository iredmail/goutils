package database

import (
	"database/sql"
	"time"

	"github.com/go-ldap/ldap/v3"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

const (
	DBTypeMySQL DBType = "mysql"
	DBTypePgSQL DBType = "postgres"
	DBTypeLDAP  DBType = "openldap"
)

type DBType string

func (dt DBType) String() string {
	return string(dt)
}

type Conn interface {
	*sql.DB | *ldap.Client
}

type Config interface {
	SQLConfig | LDAPConfig
}

type Connection[C Config, T Conn] interface {
	GetDBType() DBType
	GetConfig() C
	Connect() (conn T, err error)
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

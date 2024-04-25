package database

import (
	"crypto/tls"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/iredmail/ldappool"
	_ "github.com/lib/pq"
)

const (
	MySQLType Type = "mysql"
	PgSQLType Type = "postgres"
	LDAPType  Type = "openldap"
)

type Type string

type Connection[C any, T any] interface {
	GetType() string
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

type Mariadb[C, T any] struct {
	SQLConfig
}

func (m *Mariadb[C, T]) GetType() string {
	return string(MySQLType)
}

func (m *Mariadb[C, T]) GetConfig() C {
	return any(m.SQLConfig)
}

func (m *Mariadb[C, T]) Connect() (conn T, err error) {
	// supported params：
	// https://github.com/go-sql-driver/mysql#parameters
	//
	//	- parseTime=true format time.
	//	- timeout: Timeout for establishing connections, aka dial timeout.
	//	- writeTimeout: I/O write timeout.
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		m.DBUser,
		m.DBPassword,
		m.DBHost,
		m.DBPort,
		m.DBName,
	)

	db, err := sql.Open(string(MySQLType), dsn)
	if err != nil {
		return
	}

	if m.MaxLifetime == 0 {
		m.MaxLifetime = time.Minute * 10
	}
	if m.MaxIdleConns == 0 {
		m.MaxIdleConns = 10
	}
	if m.MaxOpenConns == 0 {
		m.MaxOpenConns = 10
	}

	db.SetConnMaxLifetime(m.MaxLifetime)
	db.SetMaxOpenConns(m.MaxOpenConns)
	db.SetMaxIdleConns(m.MaxIdleConns)

	conn = any(db)

	return
}

type PgSQL[C, T any] struct {
	SQLConfig
}

func (p *PgSQL[C, T]) GetType() string {
	return string(PgSQLType)
}

func (p *PgSQL[C, T]) GetConfig() C {
	return any(p.SQLConfig)
}

func (p *PgSQL[C, T]) Connect() (conn T, err error) {
	// supported params：
	// https://pkg.go.dev/github.com/lib/pq#hdr-Connection_String_Parameters
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		p.DBUser,
		p.DBPassword,
		p.DBHost,
		p.DBPort,
		p.DBName,
	)

	db, err := sql.Open(string(PgSQLType), dsn)
	if err != nil {
		return
	}

	if p.MaxLifetime == 0 {
		p.MaxLifetime = time.Minute * 10
	}
	if p.MaxIdleConns == 0 {
		p.MaxIdleConns = 10
	}
	if p.MaxOpenConns == 0 {
		p.MaxOpenConns = 10
	}

	db.SetConnMaxLifetime(p.MaxLifetime)
	db.SetMaxOpenConns(p.MaxOpenConns)
	db.SetMaxIdleConns(p.MaxIdleConns)

	conn = any(db)

	return
}

type LDAPConfig struct {
	URI                string
	Suffix             string // dc=xx
	BaseDN             string // o=domains,dc=xx
	DomainAdminsBaseDN string // o=domainAdmins,dc=xx,dc=xx
	BindDN             string
	BindPassword       string
	StartTLS           bool
}

type LDAP[C, T any] struct {
	LDAPConfig
}

func (l *LDAP[C, T]) GetType() string {
	return string(LDAPType)
}

func (l *LDAP[C, T]) GetConfig() C {
	return any(l.LDAPConfig)
}

func (l *LDAP[C, T]) Connect() (conn T, err error) {
	opts := []ldappool.Option{
		ldappool.WithBindCredentials(l.BindDN, l.BindPassword),
	}

	if l.StartTLS {
		opts = append(opts, ldappool.WithTLSConfig(&tls.Config{InsecureSkipVerify: true}))
	}

	pool, err := ldappool.New(l.URI, opts...)
	if err != nil {
		return
	}

	conn = any(pool)

	return
}

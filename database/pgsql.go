package database

import (
	"database/sql"
	"fmt"
	"time"
)

type PgSQL[C Config, T Conn] struct {
	SQLConfig
}

func (p *PgSQL[C, T]) GetDBType() DBType {
	return DBTypePgSQL
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

	db, err := sql.Open(string(DBTypePgSQL), dsn)
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

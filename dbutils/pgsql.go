package dbutils

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

func NewPgSQL(c SQLConnConfig) (db *sql.DB, err error) {
	// supported params：
	// https://pkg.go.dev/github.com/lib/pq#hdr-Connection_String_Parameters
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		c.DBUser,
		c.DBPassword,
		c.DBHost,
		c.DBPort,
		c.DBName,
	)

	db, err = sql.Open("postgres", dsn)
	if err != nil {
		return
	}

	if c.MaxLifetime == 0 {
		c.MaxLifetime = time.Minute * 10
	}
	if c.MaxIdleConns == 0 {
		c.MaxIdleConns = 10
	}
	if c.MaxOpenConns == 0 {
		c.MaxOpenConns = 10
	}

	db.SetConnMaxLifetime(c.MaxLifetime)
	db.SetMaxOpenConns(c.MaxOpenConns)
	db.SetMaxIdleConns(c.MaxIdleConns)

	return
}

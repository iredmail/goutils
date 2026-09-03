package dbutils

import (
	"database/sql"
	"net"
	"net/url"
	"time"

	_ "github.com/lib/pq"
)

func NewPgSQL(c SQLConnConfig) (db *sql.DB, err error) {
	// supported params：
	// https://pkg.go.dev/github.com/lib/pq#hdr-Connection_String_Parameters
	u := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(c.DBUser, c.DBPassword),
		Host:   net.JoinHostPort(c.DBHost, c.DBPort),
		Path:   c.DBName,
	}

	sslmode := "disable"
	if c.UseSSL && c.VerifyCert {
		sslmode = "require"
	}

	q := u.Query()
	q.Set("sslmode", sslmode)
	u.RawQuery = q.Encode()
	dsn := u.String()

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

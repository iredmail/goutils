package dbutils

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func NewMySQLConn(c SQLConnConfig) (db *sql.DB, err error) {
	// supported params：
	// https://github.com/go-sql-driver/mysql#parameters
	//
	//	- parseTime=true format time.
	//	- timeout: Timeout for establishing connections, aka dial timeout.
	//	- writeTimeout: I/O write timeout.
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&multiStatements=true",
		c.DBUser,
		c.DBPassword,
		c.DBHost,
		c.DBPort,
		c.DBName,
	)

	// Add tls=skip-verify if UseSSL is true.
	if c.UseSSL {
		dsn += "&tls=skip-verify"
	}

	db, err = sql.Open("mysql", dsn)
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

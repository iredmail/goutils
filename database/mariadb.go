package database

import (
	"database/sql"
	"fmt"
	"time"
)

type Mariadb[C Config, T Conn] struct {
	SQLConfig
}

func (m *Mariadb[C, T]) GetDBType() DBType {
	return DBTypeMySQL
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

	db, err := sql.Open(string(DBTypeMySQL), dsn)
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

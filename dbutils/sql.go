package dbutils

import (
	"time"
)

type SQLConnConfig struct {
	DBHost     string
	DBPort     string
	UseSSL     bool
	VerifyCert bool // force or skip ssl cert verification
	DBUser     string
	DBPassword string
	DBName     string

	// For SQL driver
	MaxLifetime  time.Duration
	MaxOpenConns int
	MaxIdleConns int
}

type DBSize struct {
	Name string `json:"name" db:"name"`
	Size int64  `json:"size" db:"size"` // in bytes
}

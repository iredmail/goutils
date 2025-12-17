package dbutils

import (
	"time"
)

type SQLConnConfig struct {
	DBHost string
	DBPort string
	UseSSL bool
	// TODO Add new field `VerifyCert bool` to force or skip ssl cert verification.
	// VerifyCert bool
	DBUser     string
	DBPassword string
	DBName     string

	// For SQL driver
	MaxLifetime  time.Duration
	MaxOpenConns int
	MaxIdleConns int
}

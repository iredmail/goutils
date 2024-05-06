package dbutils

import (
	"time"
)

type SQLConnConfig struct {
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

package sslcert

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/acme/autocert"
)

// ErrEmptyTableName is returned when given table name is empty
var ErrEmptyTableName = errors.New("table name must not be empty")

// Making sure that we're adhering to the autocert.Cache interface.
var _ autocert.Cache = (*Cache)(nil)

// Cache provides a SQL backend to the autocert cache.
type Cache struct {
	conn *sql.DB

	// SQL statements for the cache operations.
	getQuery    string // Get()
	putQuery    string // Put()
	deleteQuery string // Delete()
}

// Get returns a certificate data for the specified key.
// If there's no such key, Get returns ErrCacheMiss.
func (c *Cache) Get(ctx context.Context, key string) (data []byte, err error) {
	err = c.conn.
		QueryRowContext(ctx, c.getQuery, key).
		Scan(&data)

	if errors.Is(err, sql.ErrNoRows) {
		err = autocert.ErrCacheMiss
	}

	return
}

// Put stores the data in the cache under the specified key.
func (c *Cache) Put(ctx context.Context, key string, data []byte) (err error) {
	_, err = c.conn.ExecContext(ctx, c.putQuery, key, data)

	return
}

// Delete removes a certificate data from the cache under the specified key.
// If there's no such key in the cache, Delete returns nil.
func (c *Cache) Delete(ctx context.Context, key string) (err error) {
	_, err = c.conn.ExecContext(ctx, c.deleteQuery, key)

	return
}

// NewSQLiteCache creates an cache instance that can be used with autocert.Cache.
// It returns any errors that could happen while connecting to SQL.
func NewSQLiteCache(conn *sql.DB, tableName string) (cache *Cache, err error) {
	tableName = strings.TrimSpace(tableName)
	if tableName == "" {
		err = ErrEmptyTableName

		return
	}

	// Create SQLite table if not exists.
	_, err = conn.Exec(fmt.Sprintf(`
        CREATE TABLE IF NOT EXISTS %s (
            key        VARCHAR(255) NOT NULL DEFAULT '' PRIMARY KEY,
            data       BLOB,
            created_at INTETER DEFAULT 0,
            updated_at INTETER DEFAULT 0
        );
        CREATE INDEX IF NOT EXISTS idx_%s_created_at ON %s (key);`,
		tableName, tableName, tableName,
	))

	if err != nil {
		return
	}

	cache = &Cache{
		conn:     conn,
		getQuery: fmt.Sprintf(`SELECT data FROM %s WHERE key = $1`, tableName),
		putQuery: fmt.Sprintf(`
			INSERT INTO %s (key, data, created_at)
		            VALUES ($1, $2, unixepoch())
			ON CONFLICT (key) DO UPDATE SET data = $2, updated_at = unixepoch()
		`, tableName),
		deleteQuery: fmt.Sprintf(`DELETE FROM %s WHERE key = $1`, tableName),
	}

	return
}

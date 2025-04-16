package sslcert

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/acme/autocert"
	_ "modernc.org/sqlite"
)

var conn *sql.DB

func init() {
	var err error

	conn, err = sql.Open("sqlite", "file:test.db?cache=shared&mode=memory")
	if err != nil {
		panic(err)
	}
}

func TestNewSQLiteCache(t *testing.T) {
	tbl := "NewSQLiteCache"

	cache, err := NewSQLiteCache(conn, tbl)
	assert.NotNil(t, cache)
	assert.Nil(t, err)

	var count int
	err = conn.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", tbl)).Scan(&count)
	assert.Nil(t, err)
	assert.Equal(t, count, 0)
}

func TestGetUnknownKey(t *testing.T) {
	cache, _ := NewSQLiteCache(conn, "GetUnknownKey")
	data, err := cache.Get(context.Background(), "my-key")
	assert.Equal(t, err, autocert.ErrCacheMiss)
	assert.Equal(t, len(data), 0)
}

func TestGetAfterPut(t *testing.T) {
	tbl := "GetAfterPut"

	cache, _ := NewSQLiteCache(conn, tbl)

	actual := []byte{1}
	err := cache.Put(context.Background(), "my-key", actual)
	assert.Nil(t, err)

	data, err := cache.Get(context.Background(), "my-key")
	assert.Nil(t, err)
	assert.Equal(t, data, actual)

	var count int
	err = conn.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", tbl)).Scan(&count)
	assert.Nil(t, err)
	assert.Equal(t, count, 1)
}

func TestGetAfterDelete(t *testing.T) {
	cache, _ := NewSQLiteCache(conn, "GetAfterDelete")

	actual := []byte{1, 2, 3, 4}
	err := cache.Put(context.Background(), "my-key", actual)
	assert.Nil(t, err)

	err = cache.Delete(context.Background(), "my-key")
	assert.Nil(t, err)

	data, err := cache.Get(context.Background(), "my-key")
	assert.Equal(t, err, autocert.ErrCacheMiss)
	assert.Equal(t, len(data), 0)
}

func TestDeleteUnknownKey(t *testing.T) {
	cache, _ := NewSQLiteCache(conn, "DeleteUnknownKey")

	var err error

	err = cache.Delete(context.Background(), "my-key1")
	assert.Nil(t, err)
	err = cache.Delete(context.Background(), "other-key")
	assert.Nil(t, err)
	err = cache.Delete(context.Background(), "hello-world")
	assert.Nil(t, err)
}

func TestPutOverwrite(t *testing.T) {
	cache, _ := NewSQLiteCache(conn, "PutOverwrite")

	data1 := []byte{1, 2, 3, 4}
	err := cache.Put(context.Background(), "thekey", data1)
	assert.Nil(t, err)
	data, err := cache.Get(context.Background(), "thekey")
	assert.Equal(t, data, data1)

	data2 := []byte{5, 6, 7, 8}
	err = cache.Put(context.Background(), "thekey", data2)
	assert.Nil(t, err)
	data, err = cache.Get(context.Background(), "thekey")
	assert.Equal(t, data, data2)
}

func TestDifferentTableName(t *testing.T) {
	tbl := "DifferentTableName"

	cache, _ := NewSQLiteCache(conn, tbl)

	actual := []byte{1, 2, 3, 4}
	err := cache.Put(context.Background(), "thekey.hi", actual)
	assert.Nil(t, err)

	var count int
	err = conn.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", tbl)).Scan(&count)
	assert.Nil(t, err)
	assert.Equal(t, count, 1)

	// 查询不存在的 sql 表
	err = conn.QueryRow("SELECT COUNT(*) FROM cert_store").Scan(&count)
	assert.NotNil(t, err)
}

func TestGetCancelledContext(t *testing.T) {
	tbl := "GetCancelledContext"

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	cache, _ := NewSQLiteCache(conn, tbl)
	data, err := cache.Get(ctx, "my-key")
	assert.Equal(t, err, context.Canceled)
	assert.Equal(t, len(data), 0)
}

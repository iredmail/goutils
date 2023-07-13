package ctxutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenPagination(t *testing.T) {
	var total int64
	var limit, page int

	total = 0
	limit = 10
	page = 1

	p := GenPagination(page, total, limit)
	assert.Equal(t, p.TotalItems, int64(0))
	assert.Equal(t, p.TotalPages, 0)
	assert.Equal(t, p.CurrentPage, 1)
	assert.Equal(t, p.Limit, 10)
	assert.Equal(t, p.PageNumbers, []int{})

	// 10 pages
	total = 100
	limit = 10
	page = 2

	p = GenPagination(page, total, limit)
	assert.Equal(t, p.TotalItems, int64(100))
	assert.Equal(t, p.TotalPages, 10)
	assert.Equal(t, p.CurrentPage, 2)
	assert.Equal(t, p.Limit, 10)
	assert.Equal(t, p.PageNumbers, []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})

	// 6 pages
	total = 53
	limit = 10
	page = 7

	p = GenPagination(page, total, limit)
	assert.Equal(t, p.TotalItems, int64(53))
	assert.Equal(t, p.TotalPages, 6)
	assert.Equal(t, p.CurrentPage, 7)
	assert.Equal(t, p.Limit, 10)
	assert.Equal(t, p.PageNumbers, []int{1, 2, 3, 4, 5, 6})

	// 12 pages
	total = 120
	limit = 10
	page = 8

	p = GenPagination(page, total, limit)
	assert.Equal(t, p.TotalItems, int64(120))
	assert.Equal(t, p.TotalPages, 12)
	assert.Equal(t, p.CurrentPage, 8)
	assert.Equal(t, p.Limit, 10)
	assert.Equal(t, p.PageNumbers, []int{1, 0, 5, 6, 7, 8, 9, 10, 11, 12})

	// 13 pages
	total = 130
	limit = 10
	page = 8

	p = GenPagination(page, total, limit)
	assert.Equal(t, p.TotalItems, int64(130))
	assert.Equal(t, p.TotalPages, 13)
	assert.Equal(t, p.CurrentPage, 8)
	assert.Equal(t, p.Limit, 10)
	assert.Equal(t, p.PageNumbers, []int{1, 0, 5, 6, 7, 8, 9, 10, 11, 0, 13})
}

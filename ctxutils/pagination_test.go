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
	assert.Equal(t, p.PageBeginNum, 0)
	assert.Equal(t, p.PageLastNum, 0)

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
	assert.Equal(t, p.PageBeginNum, 11)
	assert.Equal(t, p.PageLastNum, 20)

	// 6 pages
	total = 53
	limit = 10
	page = 5

	p = GenPagination(page, total, limit)
	assert.Equal(t, p.TotalItems, int64(53))
	assert.Equal(t, p.TotalPages, 6)
	assert.Equal(t, p.CurrentPage, 5)
	assert.Equal(t, p.Limit, 10)
	assert.Equal(t, p.PageNumbers, []int{1, 2, 3, 4, 5, 6})
	assert.Equal(t, p.PageBeginNum, 41)
	assert.Equal(t, p.PageLastNum, 50)

	// 测试最后一页的最后一条和最后一条的序号
	page = 6
	p = GenPagination(page, total, limit)
	assert.Equal(t, p.PageBeginNum, 51)
	assert.Equal(t, p.PageLastNum, 53)

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

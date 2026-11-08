package ctxutils

import (
	"math"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenPagination(t *testing.T) {
	tests := []struct {
		name      string
		page      int
		uriPrefix string
		total     int64
		limit     int
		queries   map[string]string
		want      Pagination
	}{
		{
			name:  "limit 为 0 时返回零值",
			page:  1,
			total: 100,
			limit: 0,
			want:  Pagination{},
		},
		{
			name:  "limit 为负数时返回零值",
			page:  1,
			total: 100,
			limit: -5,
			want:  Pagination{},
		},
		{
			name:      "total 为 0",
			page:      1,
			uriPrefix: "/abc",
			total:     0,
			limit:     10,
			want: Pagination{
				TotalItems:   0,
				TotalPages:   0,
				CurrentPage:  1,
				Limit:        10,
				PageNumbers:  []int{},
				URIPrefix:    "/abc?limit=10",
				PageBeginNum: 0,
				PageLastNum:  0,
			},
		},
		{
			name:      "total 为 0 时越界页码收敛到第 1 页",
			page:      999,
			uriPrefix: "/",
			total:     0,
			limit:     10,
			want: Pagination{
				TotalItems:   0,
				TotalPages:   0,
				CurrentPage:  1,
				Limit:        10,
				PageNumbers:  []int{},
				URIPrefix:    "/?limit=10",
				PageBeginNum: 0,
				PageLastNum:  0,
			},
		},
		{
			name:      "单页且刚好整除",
			page:      1,
			uriPrefix: "/",
			total:     10,
			limit:     10,
			want: Pagination{
				TotalItems:   10,
				TotalPages:   1,
				CurrentPage:  1,
				Limit:        10,
				PageNumbers:  []int{},
				URIPrefix:    "/?limit=10",
				PageBeginNum: 1,
				PageLastNum:  10,
			},
		},
		{
			name:      "单页且不满一页",
			page:      1,
			uriPrefix: "/",
			total:     1,
			limit:     10,
			want: Pagination{
				TotalItems:   1,
				TotalPages:   1,
				CurrentPage:  1,
				Limit:        10,
				PageNumbers:  []int{},
				URIPrefix:    "/?limit=10",
				PageBeginNum: 1,
				PageLastNum:  1,
			},
		},
		{
			name:      "10 页以内全部显示",
			page:      2,
			uriPrefix: "/",
			total:     100,
			limit:     10,
			want: Pagination{
				TotalItems:   100,
				TotalPages:   10,
				CurrentPage:  2,
				Limit:        10,
				PageNumbers:  []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
				URIPrefix:    "/?limit=10",
				PageBeginNum: 11,
				PageLastNum:  20,
			},
		},
		{
			name:      "11 页第 1 页",
			page:      1,
			uriPrefix: "/",
			total:     110,
			limit:     10,
			want: Pagination{
				TotalItems:   110,
				TotalPages:   11,
				CurrentPage:  1,
				Limit:        10,
				PageNumbers:  []int{1, 2, 3, 4, 0, 11},
				URIPrefix:    "/?limit=10",
				PageBeginNum: 1,
				PageLastNum:  10,
			},
		},
		{
			name:      "11 页第 4 页",
			page:      4,
			uriPrefix: "/",
			total:     110,
			limit:     10,
			want: Pagination{
				TotalItems:   110,
				TotalPages:   11,
				CurrentPage:  4,
				Limit:        10,
				PageNumbers:  []int{1, 2, 3, 4, 5, 6, 7, 0, 11},
				URIPrefix:    "/?limit=10",
				PageBeginNum: 31,
				PageLastNum:  40,
			},
		},
		{
			name:      "11 页第 5 页不产生冗余省略号",
			page:      5,
			uriPrefix: "/",
			total:     110,
			limit:     10,
			want: Pagination{
				TotalItems:   110,
				TotalPages:   11,
				CurrentPage:  5,
				Limit:        10,
				PageNumbers:  []int{1, 2, 3, 4, 5, 6, 7, 8, 0, 11},
				URIPrefix:    "/?limit=10",
				PageBeginNum: 41,
				PageLastNum:  50,
			},
		},
		{
			name:      "11 页第 6 页",
			page:      6,
			uriPrefix: "/",
			total:     110,
			limit:     10,
			want: Pagination{
				TotalItems:   110,
				TotalPages:   11,
				CurrentPage:  6,
				Limit:        10,
				PageNumbers:  []int{1, 0, 3, 4, 5, 6, 7, 8, 9, 0, 11},
				URIPrefix:    "/?limit=10",
				PageBeginNum: 51,
				PageLastNum:  60,
			},
		},
		{
			name:      "11 页第 7 页显示尾部连续页码",
			page:      7,
			uriPrefix: "/",
			total:     110,
			limit:     10,
			want: Pagination{
				TotalItems:   110,
				TotalPages:   11,
				CurrentPage:  7,
				Limit:        10,
				PageNumbers:  []int{1, 0, 4, 5, 6, 7, 8, 9, 10, 11},
				URIPrefix:    "/?limit=10",
				PageBeginNum: 61,
				PageLastNum:  70,
			},
		},
		{
			name:      "11 页最后一页",
			page:      11,
			uriPrefix: "/",
			total:     110,
			limit:     10,
			want: Pagination{
				TotalItems:   110,
				TotalPages:   11,
				CurrentPage:  11,
				Limit:        10,
				PageNumbers:  []int{1, 0, 8, 9, 10, 11},
				URIPrefix:    "/?limit=10",
				PageBeginNum: 101,
				PageLastNum:  110,
			},
		},
		{
			name:      "12 页中间页",
			page:      8,
			uriPrefix: "/",
			total:     120,
			limit:     10,
			want: Pagination{
				TotalItems:   120,
				TotalPages:   12,
				CurrentPage:  8,
				Limit:        10,
				PageNumbers:  []int{1, 0, 5, 6, 7, 8, 9, 10, 11, 12},
				URIPrefix:    "/?limit=10",
				PageBeginNum: 71,
				PageLastNum:  80,
			},
		},
		{
			name:      "13 页中间页",
			page:      8,
			uriPrefix: "/",
			total:     130,
			limit:     10,
			want: Pagination{
				TotalItems:   130,
				TotalPages:   13,
				CurrentPage:  8,
				Limit:        10,
				PageNumbers:  []int{1, 0, 5, 6, 7, 8, 9, 10, 11, 0, 13},
				URIPrefix:    "/?limit=10",
				PageBeginNum: 71,
				PageLastNum:  80,
			},
		},
		{
			name:      "最后一页且刚好整除",
			page:      12,
			uriPrefix: "/",
			total:     120,
			limit:     10,
			want: Pagination{
				TotalItems:   120,
				TotalPages:   12,
				CurrentPage:  12,
				Limit:        10,
				PageNumbers:  []int{1, 0, 9, 10, 11, 12},
				URIPrefix:    "/?limit=10",
				PageBeginNum: 111,
				PageLastNum:  120,
			},
		},
		{
			name:      "最后一页且不满一页",
			page:      6,
			uriPrefix: "/",
			total:     53,
			limit:     10,
			want: Pagination{
				TotalItems:   53,
				TotalPages:   6,
				CurrentPage:  6,
				Limit:        10,
				PageNumbers:  []int{1, 2, 3, 4, 5, 6},
				URIPrefix:    "/?limit=10",
				PageBeginNum: 51,
				PageLastNum:  53,
			},
		},
		{
			name:      "越界页码收敛到最后一页",
			page:      999,
			uriPrefix: "/",
			total:     53,
			limit:     10,
			want: Pagination{
				TotalItems:   53,
				TotalPages:   6,
				CurrentPage:  6,
				Limit:        10,
				PageNumbers:  []int{1, 2, 3, 4, 5, 6},
				URIPrefix:    "/?limit=10",
				PageBeginNum: 51,
				PageLastNum:  53,
			},
		},
		{
			name:      "页码 0 收敛到第 1 页",
			page:      0,
			uriPrefix: "/",
			total:     53,
			limit:     10,
			want: Pagination{
				TotalItems:   53,
				TotalPages:   6,
				CurrentPage:  1,
				Limit:        10,
				PageNumbers:  []int{1, 2, 3, 4, 5, 6},
				URIPrefix:    "/?limit=10",
				PageBeginNum: 1,
				PageLastNum:  10,
			},
		},
		{
			name:      "URIPrefix 已含 query 时用 & 连接",
			page:      1,
			uriPrefix: "/abc?x=1",
			total:     20,
			limit:     10,
			want: Pagination{
				TotalItems:   20,
				TotalPages:   2,
				CurrentPage:  1,
				Limit:        10,
				PageNumbers:  []int{1, 2},
				URIPrefix:    "/abc?x=1&limit=10",
				PageBeginNum: 1,
				PageLastNum:  10,
			},
		},
		{
			name:      "ctxQueries 追加普通参数",
			page:      1,
			uriPrefix: "/abc",
			total:     20,
			limit:     10,
			queries:   map[string]string{"foo": "bar"},
			want: Pagination{
				TotalItems:   20,
				TotalPages:   2,
				CurrentPage:  1,
				Limit:        10,
				PageNumbers:  []int{1, 2},
				URIPrefix:    "/abc?foo=bar&limit=10",
				PageBeginNum: 1,
				PageLastNum:  10,
			},
		},
		{
			name:      "ctxQueries 移除空值参数",
			page:      1,
			uriPrefix: "/abc",
			total:     20,
			limit:     10,
			queries:   map[string]string{"foo": ""},
			want: Pagination{
				TotalItems:   20,
				TotalPages:   2,
				CurrentPage:  1,
				Limit:        10,
				PageNumbers:  []int{1, 2},
				URIPrefix:    "/abc?limit=10",
				PageBeginNum: 1,
				PageLastNum:  10,
			},
		},
		{
			name:      "ctxQueries 中的 limit 被实际 limit 覆盖",
			page:      1,
			uriPrefix: "/abc",
			total:     20,
			limit:     10,
			queries:   map[string]string{"limit": "999"},
			want: Pagination{
				TotalItems:   20,
				TotalPages:   2,
				CurrentPage:  1,
				Limit:        10,
				PageNumbers:  []int{1, 2},
				URIPrefix:    "/abc?limit=10",
				PageBeginNum: 1,
				PageLastNum:  10,
			},
		},
		{
			name:      "ctxQueries 参数值按 URL 编码",
			page:      1,
			uriPrefix: "/abc",
			total:     20,
			limit:     10,
			queries:   map[string]string{"foo": "a b"},
			want: Pagination{
				TotalItems:   20,
				TotalPages:   2,
				CurrentPage:  1,
				Limit:        10,
				PageNumbers:  []int{1, 2},
				URIPrefix:    "/abc?foo=a+b&limit=10",
				PageBeginNum: 1,
				PageLastNum:  10,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenPagination(tt.page, tt.uriPrefix, tt.total, tt.limit, tt.queries)

			assert.Equal(t, tt.want, got)
		})
	}
}

// TestGenPaginationLargeTotal 覆盖超大 total 的分页计算。
// 分页计算使用整数除法，不应受 float64 精度影响。
func TestGenPaginationLargeTotal(t *testing.T) {
	if strconv.IntSize < 64 {
		t.Skip("需要 64 位 int 平台")
	}

	total := int64(math.MaxInt64)
	// MaxInt64 除以 10 的余数为 7，因此需要向上取整。
	wantPages := int(total/10 + 1)

	p := GenPagination(1, "/", total, 10, nil)

	assert.Equal(t, total, p.TotalItems)
	assert.Equal(t, wantPages, p.TotalPages)
	assert.Equal(t, 1, p.CurrentPage)
	assert.Equal(t, 1, p.PageBeginNum)
	assert.Equal(t, 10, p.PageLastNum)
}

// TestGenPaginationMaxPages 覆盖总页数达到 int 最大值时的页码列表生成，
// 确保不因页码递增溢出而死循环。
func TestGenPaginationMaxPages(t *testing.T) {
	if strconv.IntSize < 64 {
		t.Skip("需要 64 位 int 平台")
	}

	total := int64(math.MaxInt64)
	// 32 位平台会在上面的 Skip 处返回；运行时转换避免编译期常量溢出。
	maxInt := int(total)

	p := GenPagination(maxInt-4, "/", total, 1, nil)

	assert.Equal(t, maxInt, p.TotalPages)
	assert.Equal(t, maxInt-4, p.CurrentPage)
	assert.Equal(t, []int{
		1,
		0,
		maxInt - 7,
		maxInt - 6,
		maxInt - 5,
		maxInt - 4,
		maxInt - 3,
		maxInt - 2,
		maxInt - 1,
		maxInt,
	}, p.PageNumbers)
}

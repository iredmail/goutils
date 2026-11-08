package ctxutils

import (
	"fmt"
	"net/url"
	"strings"
)

// Pagination 描述分页结果：总条目数、当前页、页码列表与记录序号区间。
// PageNumbers 中的 0 表示省略的页码范围；URIPrefix 为分页链接的前缀。
type Pagination struct {
	TotalItems  int64
	TotalPages  int
	CurrentPage int
	Limit       int    // Page size limit
	PageNumbers []int  // 数字为 0 表示省略的范围，可以以省略号表示。
	URIPrefix   string // 分页数字对应链接的前缀 URI。分页数字将以 query `&page=<num>` 格式追加其后。参数在其后。

	// 当前页第一条和最后一条项目的序号
	PageBeginNum int
	PageLastNum  int
}

// GenPagination 根据当前页 `page`，总条目数 `total`，每页条目数 `limit` 生成分页链接。
// limit <= 0 时返回零值；页码会被收敛到 [1, 总页数] 范围内。
func GenPagination(page int, URIPrefix string, total int64, limit int, ctxQueries map[string]string) (p Pagination) {
	if limit <= 0 {
		return
	}

	// 用整数除法向上取整，避免超大 total 在浮点转换时丢失精度。
	pages := 0
	if total > 0 {
		pages = int(total / int64(limit))
		if total%int64(limit) != 0 {
			pages++
		}
	}

	// 页码收敛到 [1, 总页数]，避免越界页码产生错误的记录区间和页码列表。
	// total 为 0 时没有有效页，按第 1 页处理。
	page = max(min(page, pages), 1)

	p = Pagination{
		TotalItems:   total,
		TotalPages:   pages,
		CurrentPage:  page,
		Limit:        limit,
		PageNumbers:  []int{},
		URIPrefix:    URIPrefix,
		PageBeginNum: (page-1)*limit + 1,
		PageLastNum:  page * limit,
	}

	uv := url.Values{}

	for k, v := range ctxQueries {
		// 移除空值参数，避免生成类似 `&foo=` 的链接。
		if v == "" {
			continue
		}

		uv.Set(k, v)
	}

	// 总是覆盖 `limit` 参数
	if uv.Has("limit") {
		uv.Set("limit", fmt.Sprintf("%d", limit))
	} else {
		uv.Add("limit", fmt.Sprintf("%d", limit))
	}

	sep := "?"
	if strings.Contains(p.URIPrefix, sep) {
		sep = "&"
	}
	p.URIPrefix += sep + uv.Encode()

	if total == 0 {
		p.PageBeginNum = 0
		p.PageLastNum = 0

		return
	}

	if pages <= 1 {
		// 不需要分页
		p.PageLastNum = int(total)

		return
	}

	if page == pages {
		// 最后一页
		p.PageLastNum = (page-1)*limit + int(total-int64((page-1)*limit))
	}

	var nums []int
	if pages <= 10 {
		// 10 页以内直接全部显示
		for i := 1; i <= pages; i++ {
			nums = append(nums, i)
		}
	} else {
		//
		// 第一页、当前页及其前后3页，最后一页
		//

		// 第一页
		nums = append(nums, 1)

		if page <= 4 {
			// 前面几页全部显示
			for i := 2; i <= page; i++ {
				nums = append(nums, i)
			}

			// 后面3页及最后一页
			nums = append(nums,
				page+1,
				page+2,
				page+3,
				0,     // 省略号
				pages, // 最后一页
			)
		} else if page < pages-4 {
			// 当前页及其前后 3 页。页码 1 与 page-3 相邻时不需要省略号。
			if page-3 > 2 {
				nums = append(nums, 0) // 省略号
			}

			nums = append(nums,
				page-3,
				page-2,
				page-1,
				page, // 当前页
				page+1,
				page+2,
				page+3,
				0,     // 省略号
				pages, // 最后一页
			)
		} else {
			// 前面添加一个省略号（不带链接）

			// 当前页及其前面3页
			nums = append(nums,
				0,
				page-3,
				page-2,
				page-1,
				page, // 当前页
			)

			// 最后几页全部显示。用 i < pages 判断，避免 pages 为 int 最大值时
			// i++ 回绕导致死循环。
			for i := page; i < pages; {
				i++
				nums = append(nums, i)
			}
		}
	}

	p.PageNumbers = nums

	return
}

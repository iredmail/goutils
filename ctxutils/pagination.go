package ctxutils

import (
	"fmt"
	"math"
	"net/url"
)

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
func GenPagination(page int, URIPrefix string, total int64, limit int, ctxQueries map[string]string) (p Pagination) {
	p = Pagination{
		TotalItems:   total,
		TotalPages:   int(math.Ceil(float64(total) / float64(limit))),
		CurrentPage:  page,
		Limit:        limit,
		PageNumbers:  []int{},
		URIPrefix:    URIPrefix,
		PageBeginNum: (page-1)*limit + 1,
		PageLastNum:  page * limit,
	}

	uv := url.Values{}

	if len(ctxQueries) > 0 {
		for k, v := range ctxQueries {
			if uv.Has(k) {
				// 已经存在则替换旧值
				uv.Set(k, v)
			} else {
				// 不存在则添加
				uv.Add(k, v)
			}
		}
	}

	// 总是覆盖 `limit` 参数
	if uv.Has("limit") {
		uv.Set("limit", fmt.Sprintf("%d", limit))
	} else {
		uv.Add("limit", fmt.Sprintf("%d", limit))
	}

	p.URIPrefix += "?" + uv.Encode()

	if total == 0 {
		p.PageBeginNum = 0
		p.PageLastNum = 0

		return
	}

	pages := int(math.Ceil(float64(total) / float64(limit)))

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
		} else if page > 4 && page < pages-4 {
			// 添加当前页及其前后3页
			nums = append(nums,
				0, // 省略号
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

			// 最后几页全部显示
			for i := page + 1; i <= pages; i++ {
				nums = append(nums, i)
			}
		}
	}

	p.PageNumbers = nums

	return
}

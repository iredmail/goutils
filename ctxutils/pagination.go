package ctxutils

import (
	"math"
)

type Pagination struct {
	TotalItems  int64
	TotalPages  int
	CurrentPage int
	Limit       int   // Page size limit
	PageNumbers []int // 数字为 0 表示省略的范围，可以以省略号表示。
}

// GenPagination 根据当前页 `page`，总条目数 `total`，每页条目数 `limit` 生成分页链接。
func GenPagination(page int, total int64, limit int) (p Pagination) {
	p = Pagination{
		TotalItems:  total,
		TotalPages:  int(math.Ceil(float64(total) / float64(limit))),
		CurrentPage: page,
		Limit:       limit,
		PageNumbers: []int{},
	}

	var nums []int

	if total == 0 {
		return
	}

	pages := int(math.Ceil(float64(total) / float64(limit)))

	if pages <= 1 {
		// 不需要分页
		return
	}

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

package ctxutils

import (
	"math"
)

// PageLink 表示一个分页链接。
type PageLink struct {
	Page   int
	Active bool
}

// pageLink 将页数转化为结构体 PageLink。
func pageLink(currentPage, num int) PageLink {
	return PageLink{
		Page:   num,
		Active: currentPage == num,
	}
}

// GenPagination 根据当前页 `page`，总条目数 `total`，每页条目数 `limit` 生成分页链接。
func GenPagination(page int, total int64, limit int) (pl []PageLink) {
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
			pl = append(pl, pageLink(page, i))
		}
	} else {
		//
		// 第一页、当前页及其前后3页，最后一页
		//

		// 第一页
		pl = append(pl, pageLink(page, 1))

		if page <= 4 {
			// 前面几页全部显示
			for i := 2; i <= page; i++ {
				pl = append(pl, pageLink(page, i))
			}

			// 后面3页及最后一页
			pl = append(pl,
				pageLink(page, page+1),
				pageLink(page, page+2),
				pageLink(page, page+3),
				PageLink{Page: 0},     // 省略号
				pageLink(page, pages), // 最后一页
			)
		} else if page > 4 && page < pages-4 {
			// 添加当前页及其前后3页
			pl = append(pl,
				PageLink{Page: 0}, // 省略号
				pageLink(page, page-3),
				pageLink(page, page-2),
				pageLink(page, page-1),
				pageLink(page, page), // 当前页
				pageLink(page, page+1),
				pageLink(page, page+2),
				pageLink(page, page+3),
				PageLink{Page: 0},     // 省略号
				pageLink(page, pages), // 最后一页
			)
		} else {
			// 前面添加一个省略号（不带链接）

			// 当前页及其前面3页
			pl = append(pl,
				PageLink{Page: 0},
				pageLink(page, page-3),
				pageLink(page, page-2),
				pageLink(page, page-1),
				pageLink(page, page), // 当前页
			)

			// 最后几页全部显示
			for i := page + 1; i <= pages; i++ {
				pl = append(pl, pageLink(page, i))
			}
		}
	}

	return
}

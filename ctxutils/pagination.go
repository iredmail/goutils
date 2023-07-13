package ctxutils

import (
	"math"
)

type PageLink struct {
	Page   int
	Active bool
}
type pagination struct {
	Total int64
	Page  int
	Limit int
	Pages int
}

// PageLink 将页数转化为结构体 PageLink。
func (p pagination) pageLink(page int) PageLink {
	return PageLink{
		Page:   page,
		Active: p.Page == page,
	}
}

func (p pagination) pageLinks() (pl []PageLink) {
	if p.Pages <= 1 {
		return
	}

	if p.Pages <= 10 {
		// 10 页以内直接全部显示
		for i := 1; i <= p.Pages; i++ {
			pl = append(pl, p.pageLink(i))
		}
	} else {
		//
		// 第一页、当前页及其前后3页，最后一页
		//

		// 第一页
		pl = append(pl, p.pageLink(1))

		if p.Page <= 4 {
			// 前面几页全部显示
			for i := 2; i <= p.Page; i++ {
				pl = append(pl, p.pageLink(i))
			}

			// 后面3页及最后一页
			pl = append(pl,
				p.pageLink(p.Page+1),
				p.pageLink(p.Page+2),
				p.pageLink(p.Page+3),
				PageLink{Page: 0},   // 省略号
				p.pageLink(p.Pages), // 最后一页
			)
		} else if p.Page > 4 && p.Page < p.Pages-4 {
			// 添加当前页及其前后3页
			pl = append(pl,
				PageLink{Page: 0}, // 省略号
				p.pageLink(p.Page-3),
				p.pageLink(p.Page-2),
				p.pageLink(p.Page-1),
				p.pageLink(p.Page), // 当前页
				p.pageLink(p.Page+1),
				p.pageLink(p.Page+2),
				p.pageLink(p.Page+3),
				PageLink{Page: 0},   // 省略号
				p.pageLink(p.Pages), // 最后一页
			)
		} else {
			// 前面添加一个省略号（不带链接）

			// 当前页及其前面3页
			pl = append(pl,
				PageLink{Page: 0},
				p.pageLink(p.Page-3),
				p.pageLink(p.Page-2),
				p.pageLink(p.Page-1),
				p.pageLink(p.Page), // 当前页
			)

			// 最后几页全部显示
			for i := p.Page + 1; i <= p.Pages; i++ {
				pl = append(pl, p.pageLink(i))
			}
		}
	}

	return
}

// GenPagination 根据当前页 `page`，总条目数 `total`，每页条目数 `limit` 生成分页链接。
func GenPagination(page int, total int64, limit int) (pl []PageLink) {
	var p pagination

	if total == 0 {
		return
	}

	if total > 0 {
		p.Total = total
		p.Page = page
	}

	p.Pages = int(math.Ceil(float64(p.Total) / float64(p.Limit)))
	p.Limit = limit

	return p.pageLinks()
}

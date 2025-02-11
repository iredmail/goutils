package goutils

import (
	"strings"
)

type SortMethod struct {
	// sort by given name.
	Name string

	// Sort order. Asc (lowest to highest) by default.
	Desc bool
}

type SortMethods []SortMethod

func (sms SortMethods) HasName(name string) bool {
	for _, sm := range sms {
		if sm.Name == name {
			return true
		}
	}

	return false
}

// StrToSortMethods 将以冒号、逗号分隔定义的排序方法转换为 SortMethod 列表。
// 例如：`field1:asc`, `field1:asc,field2:desc`。
func StrToSortMethods(s string) (sms SortMethods) {
	for _, part := range strings.Split(s, ",") {
		name, order, found := strings.Cut(part, ":")
		name = strings.TrimSpace(name)
		order = strings.TrimSpace(order)

		// 无效的排序字段
		if len(name) == 0 {
			continue
		}

		sm := SortMethod{Name: name}

		if found {
			if order == "desc" {
				sm.Desc = true
			}
		}

		sms = append(sms, sm)
	}

	return
}

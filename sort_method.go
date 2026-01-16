package goutils

import (
	"slices"
	"strings"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
)

type SortMethod struct {
	// sort by given name.
	Name string

	// Sort order. Asc (lowest to highest) by default.
	Desc bool
}

func (sm SortMethod) IsValid() bool {
	return sm.Name != ""
}

// Order 确保 sm 里指定的排序字段是有效的，无效则用默认的 defaultSM 代替。
func (sm SortMethod) Order(validNames []string, defaultSM SortMethod) (od exp.OrderedExpression) {
	if !slices.Contains(validNames, sm.Name) {
		sm.Name = defaultSM.Name
		sm.Desc = defaultSM.Desc
	}

	if sm.Desc {
		od = goqu.I(sm.Name).Desc()
	} else {
		od = goqu.I(sm.Name).Asc()
	}

	return
}

type SortMethods []SortMethod

func (sms SortMethods) Has(name string) bool {
	for _, sm := range sms {
		if sm.Name == name {
			return true
		}
	}

	return false
}

func (sms SortMethods) Get(name string) (found bool, sm SortMethod) {
	for _, sm = range sms {
		if sm.Name == name {
			return true, sm
		}
	}

	return
}

// StrToSortMethod 将单个以冒号分隔定义的排序方法转换为 SortMethod。
// 例如：`field`, `field1:asc`, `field2:desc`。
func StrToSortMethod(s string) (sm SortMethod) {
	name, order, found := strings.Cut(s, ":")
	name = strings.TrimSpace(name)
	order = strings.TrimSpace(order)

	// 无效的排序字段
	if len(name) == 0 {
		return
	}

	sm.Name = name

	if found {
		if order == "desc" {
			sm.Desc = true
		}
	}

	return
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

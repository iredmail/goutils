package goutils

import (
	"strings"

	"github.com/doug-martin/goqu/v9"
)

// ParseDBSorts 解析字符串中的排序字段，例：field1:asc,field2:desc
func ParseDBSorts(sort string) *DBSorts {
	ds := &DBSorts{
		sorts: make(map[string]bool),
	}

	for _, s := range strings.Split(sort, ",") {
		sorts := strings.Split(strings.TrimSpace(s), ":")
		if len(sorts) != 2 {
			continue
		}

		ds.sorts[sorts[0]] = strings.ToLower(sorts[1]) == "asc"
	}

	return ds
}

type DBSorts struct {
	sorts map[string]bool
}

func (ds *DBSorts) Add(field string, asc bool) {
	if _, ok := ds.sorts[field]; !ok {
		ds.sorts[field] = asc
	}
}

func (ds *DBSorts) Has(field string) (found bool) {
	found, _ = ds.sorts[field]

	return
}

func (ds *DBSorts) Value(field string) (exist, value bool) {
	value, exist = ds.sorts[field]

	return
}

func (ds *DBSorts) Order(sd *goqu.SelectDataset) (order *goqu.SelectDataset) {
	for k, v := range ds.sorts {
		if v {
			order = sd.Order(goqu.I(k).Asc().NullsFirst())
		} else {
			order = sd.Order(goqu.I(k).Desc().NullsLast())
		}
	}

	return
}

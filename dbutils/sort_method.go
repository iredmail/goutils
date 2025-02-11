package dbutils

import (
	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/iredmail/goutils"
)

func SortMethodsToOrder(sms goutils.SortMethods) (exps []exp.OrderedExpression) {
	for _, sm := range sms {
		if sm.Desc {
			exps = append(exps, goqu.I(sm.Name).Desc().NullsLast())
		} else {
			exps = append(exps, goqu.I(sm.Name).Asc().NullsFirst())
		}
	}

	return
}

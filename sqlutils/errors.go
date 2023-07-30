package sqlutils

import (
	"modernc.org/sqlite"
)

// ErrIsDuplicate 检测 err 是否为插入数据重复
// document: https://www.sqlite.org/rescode.html
func ErrIsDuplicate(err error) bool {
	e, ok := err.(*sqlite.Error)
	if !ok {
		return ok
	}

	// (2067) SQLITE_CONSTRAINT_UNIQUE
	// (1555) SQLITE_CONSTRAINT_PRIMARYKEY
	return e.Code() == 2067 || e.Code() == 1555
}

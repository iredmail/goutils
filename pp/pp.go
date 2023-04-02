// Package pp 是 k0kubun/pp/v3 库的简单封装，方便在已导入 goutils 库的情况下直接调用，而不需要再额外处理 import 语句。
package pp

import "github.com/k0kubun/pp/v3"

func Print(v ...interface{}) {
	_, _ = pp.Print(v...)
}

func Println(v ...interface{}) {
	_, _ = pp.Println(v...)
}

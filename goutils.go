package goutils

import (
	"net"
	"os"
	"reflect"

	"github.com/google/uuid"
	"golang.org/x/exp/slices"
)

func IsUUID(u string) bool {
	if len(u) == 0 {
		return false
	}

	_, err := uuid.Parse(u)

	return err == nil
}

// IsEmpty
// supported: string, []any, map, ptr
func IsEmpty(v any) bool {
	tf := reflect.TypeOf(v)
	switch tf.Kind() {
	case reflect.String:
		return len(v.(string)) == 0
	case reflect.Slice:
		rv := reflect.ValueOf(v)
		return rv.Len() == 0
	case reflect.Map:
		rv := reflect.ValueOf(v)
		if rv.IsNil() {
			return true
		}

		return rv.Len() == 0
	case reflect.Pointer:
		rv := reflect.ValueOf(v)
		return rv.IsNil()
	}

	return false
}

// NotEmpty
// supported: string, []any, map, ptr
func NotEmpty(v any) bool {
	rt := reflect.TypeOf(v)
	switch rt.Kind() {
	case reflect.String:
		return len(v.(string)) > 0
	case reflect.Slice:
		rv := reflect.ValueOf(v)
		return rv.Len() > 0
	case reflect.Map:
		rv := reflect.ValueOf(v)
		if rv.IsNil() {
			return false
		}

		return rv.Len() > 0
	case reflect.Ptr:
		return v != nil
	}

	return false
}

func IsIPv4(address string) bool {
	ip := net.ParseIP(address)

	return ip.To4() != nil
}

func Intersect[T comparable](s1, s2 []T) []T {
	set := make([]T, 0)
	for _, v := range s1 {
		if slices.Contains(s2, v) {
			set = append(set, v)
		}
	}

	return set
}

// DestExists 检查目标对象（文件、目录、符号链接，等）是否存在。
func DestExists(pth string) bool {
	_, err := os.Stat(pth)
	if err != nil {
		return os.IsExist(err)
	}

	return true
}

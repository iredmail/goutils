package goutils

import (
	"net"
	"reflect"

	"golang.org/x/exp/slices"
	"github.com/google/uuid"
)

func IsUUID(u string) bool {
	if len(u) == 0 {
		return false
	}

	_, err := uuid.Parse(u)

	return err == nil
}

// IsEmpty
// supported: string, []any, ptr
func IsEmpty(v any) bool {
	tf := reflect.TypeOf(v)
	switch tf.Kind() {
	case reflect.String:
		return len(v.(string)) == 0
	case reflect.Slice:
		rv := reflect.ValueOf(v)
		return rv.Len() == 0
	case reflect.Pointer:
		rv := reflect.ValueOf(v)
		return rv.IsNil()
	}

	return false
}

// NotEmpty
// supported: string, []any, ptr
func NotEmpty(v any) bool {
	rt := reflect.TypeOf(v)
	switch rt.Kind() {
	case reflect.String:
		return len(v.(string)) > 0
	case reflect.Slice:
		rv := reflect.ValueOf(v)
		return rv.Len() > 0
	case reflect.Ptr:
		return v != nil
	}

	return true
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

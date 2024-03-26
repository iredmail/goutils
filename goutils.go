package goutils

import (
	"math"
	"net"
	"reflect"
	"slices"
	"strings"
)

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

func GetStructJSONTags(v any) (tags []string) {
	rv := reflect.ValueOf(v)
	for i := range rv.NumField() {
		field := rv.Field(i)
		if field.Kind() == reflect.Struct {
			tags = append(tags, GetStructJSONTags(field.Interface())...)
		}

		jsonTag := rv.Type().Field(i).Tag.Get("json")
		tag := strings.Split(jsonTag, ",")[0]
		if tag == "" || tag == "-" {
			continue
		}

		tags = append(tags, tag)
	}

	return
}

func CalculateTotalPages(total, pageSize float64) int {
	return int(math.Ceil(total / pageSize))
}

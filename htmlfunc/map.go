package htmlfunc

import "sort"

// GetMapStringInt64 从 key 为 string 类型，值为 int64 的 map 里取值。
func GetMapStringInt64(m map[string]int64, key string) int64 {
	if v, ok := m[key]; ok {
		return v
	}

	return 0
}

// GetMapInt64String 从 key 为 int64 类型，值为 string 的 map 里取值。
func GetMapInt64String(m map[int64]string, key int64) (s string) {
	if v, ok := m[key]; ok {
		return v
	}

	return
}

type KVStrInt64 struct {
	K string
	V int64
}

func SortMapStrInt64ByInt64Desc(m map[string]int64) (kvs []KVStrInt64) {
	for k, v := range m {
		kvs = append(kvs, KVStrInt64{
			K: k,
			V: v,
		})
	}

	sort.Slice(kvs, func(i, j int) bool {
		return kvs[i].V > kvs[j].V
	})

	return
}

func SortMapStrInt64ByStrDesc(m map[string]int64) (kvs []KVStrInt64) {
	for k, v := range m {
		kvs = append(kvs, KVStrInt64{
			K: k,
			V: v,
		})
	}

	sort.Slice(kvs, func(i, j int) bool {
		return kvs[i].K > kvs[j].K
	})

	return
}

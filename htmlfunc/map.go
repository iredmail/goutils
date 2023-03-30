package htmlfunc

func getMapIntValue[T int | int64](m map[string]T, key string) T {
	if v, ok := m[key]; ok {
		return v
	}

	return 0
}

func mapIntValue(m map[string]int, key string) int {
	return getMapIntValue[int](m, key)
}

func mapInt64Value(m map[string]int64, key string) int64 {
	return getMapIntValue[int64](m, key)
}

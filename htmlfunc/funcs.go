package htmlfunc

func FuncMap() map[string]interface{} {
	return map[string]interface{}{
		"map_int_value":   mapIntValue,
		"map_int64_value": mapInt64Value,
	}
}

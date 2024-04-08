package htmlfunc

func FuncMap() map[string]interface{} {
	return map[string]interface{}{
		"get_map_int64_string": GetMapInt64String,
		"get_map_string_int64": GetMapStringInt64,
	}
}

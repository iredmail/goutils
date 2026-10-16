package htmlfunc

func FuncMap() map[string]any {
	return map[string]any{
		"get_map_int64_string": GetMapInt64String,
		"get_map_string_int64": GetMapStringInt64,
	}
}

package sqlutils

import (
	"strings"
)

func GenSQLiteURIPragmas(pragmas map[string]string) string {
	if len(pragmas) == 0 {
		return ""
	}

	var params []string
	for k, v := range pragmas {
		params = append(params, "_pragma="+k+"%3d"+v) // 以 `%3d` 代替 `=`
	}

	return strings.Join(params, "&")
}

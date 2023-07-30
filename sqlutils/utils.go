package sqlutils

import (
	"strings"
)

func GenSQLiteURIPragmas(pragmas [][2]string) string {
	if len(pragmas) == 0 {
		return ""
	}

	var params []string
	for _, p := range pragmas {
		params = append(params, "_pragma="+p[0]+"%3d"+p[1]) // 以 `%3d` 代替 `=`
	}

	return strings.Join(params, "&")
}

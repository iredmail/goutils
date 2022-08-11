package sqlutils

// SQL 表 `system` 保存 key-value 格式的值，这里针对不同数据类型的 value 定义结构体，方便
// SQL 查询时利用 goqu 做自动转换。

type KVInt struct {
	K string
	V int
}

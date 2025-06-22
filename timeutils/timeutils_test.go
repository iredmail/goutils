package timeutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMonthStartEndEpochs(t *testing.T) {
	// 测试平年2月（28天）
	start, end := MonthStartEndEpochs(2023, 2)
	assert.Equal(t, int64(1675209600), start)
	assert.Equal(t, int64(1677628799), end)

	// 测试30天的月份（2023年4月）
	start, end = MonthStartEndEpochs(2023, 4)
	assert.Equal(t, int64(1680307200), start)
	assert.Equal(t, int64(1682899199), end)

	// 测试31天的月份（2025年1月）
	start, end = MonthStartEndEpochs(2025, 1)
	assert.Equal(t, int64(1735689600), start)
	assert.Equal(t, int64(1738367999), end)

	// 测试非法月份（应回退为1月）
	start, end = MonthStartEndEpochs(2025, 13)
	assert.Equal(t, int64(1735689600), start)
	assert.Equal(t, int64(1738367999), end)
}

func TestYearStartEndEpochs(t *testing.T) {
	// 测试常规年份（2023）
	start, end := YearStartEndEpochs(2023)
	assert.Equal(t, int64(1672531200), start) // 2023-01-01 00:00:00 UTC
	assert.Equal(t, int64(1704067199), end)   // 2023-12-31 23:59:59 UTC

	// 测试闰年（2020）
	start, end = YearStartEndEpochs(2020)
	assert.Equal(t, int64(1577836800), start) // 2020-01-01 00:00:00 UTC
	assert.Equal(t, int64(1609459199), end)   // 2020-12-31 23:59:59 UTC

	// 测试纪元元年（1970）
	start, end = YearStartEndEpochs(1970)
	assert.Equal(t, int64(0), start)      // 1970-01-01 00:00:00 UTC
	assert.Equal(t, int64(31535999), end) // 1970-12-31 23:59:59 UTC
}

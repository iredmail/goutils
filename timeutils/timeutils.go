package timeutils

import (
	"strconv"
	"strings"
	"time"
)

func EpochToDay(epoch int64) string {
	if epoch <= 0 {
		return ""
	}

	return time.Unix(epoch, 0).UTC().Format("2006-01-02")
}

func EpochToDatetime(epoch int64) string {
	if epoch <= 0 {
		return ""
	}

	return time.Unix(epoch, 0).UTC().Format("2006-01-02 15:04:05")
}

func TimeToDay(t time.Time) string {
	return EpochToDay(t.UTC().Unix())
}

func TimeToDatetime(t time.Time) string {
	return EpochToDatetime(t.UTC().Unix())
}

// YMDToday 以 int 类型返回（UTC 时间）当天的 `年月日`，例如 `20230130`。
func YMDToday() (ymd int) {
	now := time.Now().UTC()
	today := strings.ReplaceAll(now.Format(time.DateOnly), "-", "")
	ymd, _ = strconv.Atoi(today)

	return
}

// YMDYesterday 以 int 类型返回（UTC 时间）昨天的 `年月日`，例如 `20230129`。
func YMDYesterday() (ymd int) {
	now := time.Now().UTC()
	yt := strings.ReplaceAll(now.AddDate(0, 0, -1).Format(time.DateOnly), "-", "")
	ymd, _ = strconv.Atoi(yt)

	return
}

func DayStartEndEpochs(t time.Time) (start, end int64) {
	year, month, day := t.Date()
	start = time.Date(year, month, day, 0, 0, 0, 0, time.UTC).Unix()
	end = time.Date(year, month, day, 23, 59, 59, int(time.Second-time.Nanosecond), time.UTC).Unix()

	return
}

// EpochsExpiringMonth 返回当前时间（start）及未来一个月（31天）的时间。
func EpochsExpiringMonth() (start, end int64) {
	t := time.Now().UTC()

	start = t.Unix()
	end = t.AddDate(0, 0, 31).Unix() // 当前时间 + 31 天

	return
}

func MonthStartEndEpochs(year, month int) (start, end int64) {
	if year == 0 {
		now := time.Now().UTC()
		year = now.Year()
	}

	// 验证月份是否有效
	if month < 1 || month > 12 {
		month = 1
	}

	// 获取当月第一天00:00:00的时间
	firstDayOfMonth := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	start = firstDayOfMonth.Unix()

	// 获取下个月第一天00:00:00的时间，然后减去1秒得到当月最后一秒
	firstDayOfNextMonth := time.Date(year, time.Month(month)+1, 1, 0, 0, 0, 0, time.UTC)
	end = firstDayOfNextMonth.Add(-time.Second).Unix()

	return
}

func YearStartEndEpochs(year int) (start, end int64) {
	if year == 0 {
		now := time.Now().UTC()
		year = now.Year()
	}

	// 获取当年第一天00:00:00的时间
	firstDayOfMonth := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	start = firstDayOfMonth.Unix()

	// 获取第二年第一天00:00:00的时间
	lastDayOfMonth := time.Date(year+1, 1, 1, 0, 0, 0, 0, time.UTC)
	end = lastDayOfMonth.Unix() - 1

	return
}

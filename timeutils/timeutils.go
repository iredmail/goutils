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

// EpochsExpiringMonth 返回当前这个月的起始以及下个月最后一天的结尾。
func EpochsExpiringMonth() (startThisMonth, endNextMonth int64) {
	t := time.Now().UTC()
	startThisMonth = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC).Unix()
	endNextMonth = time.Date(t.Year(), t.Month()+2, 1, 0, 0, 0, 0, time.UTC).Unix()

	return
}

func MonthStartEndEpochs(ts ...time.Time) (start, end int64) {
	var t time.Time
	if len(ts) > 0 {
		t = ts[0]
	} else {
		t = time.Now()
	}

	start = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC).Unix()

	// 获取下个月的第一天
	nextMonth := time.Date(t.Year(), t.Month()+1, 1, 0, 0, 0, 0, time.UTC)

	// 获取当前月份的最后一天
	end = nextMonth.AddDate(0, 0, -1).UTC().Unix()

	return
}

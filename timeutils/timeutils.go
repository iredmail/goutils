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
	ymd, _ := strconv.Atoi(yt)

	return
}

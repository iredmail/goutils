package timeutils

import "time"

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

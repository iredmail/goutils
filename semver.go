package goutils

import (
	"strings"

	"github.com/Masterminds/semver/v3"
)

func IsValidSemVerion(s string) bool {
	_, err := semver.NewVersion(s)

	return err == nil
}

func HasNewVersion(old, latest string) bool {
	// 多数情况是最新版本。
	if old == latest {
		return false
	}

	if !strings.HasPrefix(old, "v") {
		old = "v" + old
	}

	if !strings.HasPrefix(latest, "v") {
		latest = "v" + latest
	}

	ov, err1 := semver.NewVersion(old)
	lv, err2 := semver.NewVersion(latest)
	if err1 != nil || err2 != nil {
		return false
	}

	return lv.GreaterThan(ov)
}

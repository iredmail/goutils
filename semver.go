package goutils

import (
	"github.com/Masterminds/semver/v3"
)

func IsValidSemVerion(s string) bool {
	_, err := semver.NewVersion(s)

	return err == nil
}

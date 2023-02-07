package goutils

import (
	"golang.org/x/mod/semver"
)

func IsValidSemVerion(s string) bool {
	return semver.IsValid(s)
}

func HasNewVersion(old, latest string) bool {
	// Result of semver.Compare(current, latest string):
	//	- 0 if current == latest
	//	-1 if current  < latest
	//	+1 if current  > latest
	return semver.Compare(old, latest) < 0
}

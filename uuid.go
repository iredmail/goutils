package goutils

import (
	"strings"

	"github.com/google/uuid"
)

func IsUUID(u string) bool {
	if len(u) == 0 {
		return false
	}

	_, err := uuid.Parse(u)

	return err == nil
}

func NewUUIDLicenseKey() string {
	return strings.ToUpper(uuid.NewString())
}

func IsUUIDLicenseKey(s string) bool {
	return IsUUID(s)
}

package dnsutils

import (
	"errors"
	"strings"
)

func ValidateDKIMSelector(s string) error {
	s = strings.TrimSpace(s)

	if len(s) > 63 {
		return errors.New("dkim selector too long")
	}
	if s[0] == '-' || s[len(s)-1] == '-' {
		return errors.New("dkim selector cannot start or end with '-'")
	}
	for _, ch := range s {
		if (ch >= 'a' && ch <= 'z') ||
			(ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') ||
			ch == '-' {
			continue
		}

		return errors.New("dkim selector contains invalid characters")
	}

	return nil
}

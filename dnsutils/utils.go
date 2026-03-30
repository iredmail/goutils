package dnsutils

import (
	"errors"
	"fmt"
	"strings"
)

const (
	dkimSelectorMaxLength = 20
)

// ValidateDKIMSelector 检查 DKIM selector 字符串。
// 仅允许最多20个字符，且只能包含小写字母、数字和短横线。
func ValidateDKIMSelector(s string) error {
	s = strings.ToLower(strings.TrimSpace(s))

	if s == "" {
		return errors.New("selector cannot be empty")
	}

	if len(s) > dkimSelectorMaxLength {
		return fmt.Errorf("selector too long (max %d characters)", dkimSelectorMaxLength)
	}

	for _, ch := range s {
		if (ch >= 'a' && ch <= 'z') ||
			(ch >= '0' && ch <= '9') ||
			ch == '-' {
			continue
		}

		return fmt.Errorf("selector contains invalid characters, only lowercase letters, digits and dashes are allowed")
	}

	return nil
}

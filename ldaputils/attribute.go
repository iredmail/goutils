package ldaputils

import (
	"strings"

	"github.com/go-ldap/ldap/v3"
)

func GetAttrValue(entry *ldap.Entry, attr string) string {
	return strings.TrimSpace(entry.GetAttributeValue(attr))
}

func GetAttrLowerValue(entry *ldap.Entry, attr string) string {
	return strings.ToLower(strings.TrimSpace(entry.GetAttributeValue(attr)))
}

func GetAttrValues(entry *ldap.Entry, attr string) (values []string) {
	values = entry.GetAttributeValues(attr)

	for i, v := range values {
		values[i] = strings.TrimSpace(v)
	}

	return
}

func GetAttrLowerValues(entry *ldap.Entry, attr string) (values []string) {
	values = entry.GetAttributeValues(attr)

	for i, v := range values {
		values[i] = strings.ToLower(strings.TrimSpace(v))
	}

	return
}

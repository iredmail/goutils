package ldaputils

import "strings"

// ExtractLDAPSuffixFromDN 从 LDAP dn 里提取出 ldap suffix。
// 示例：cn=abc,dc=example,dc=com -> dc=example,dc=com
func ExtractLDAPSuffixFromDN(basedn string) (suffix string) {
	basedn = strings.ToLower(strings.TrimSpace(basedn))

	if strings.HasPrefix(basedn, "dc=") {
		return basedn
	}

	idxFirstDC := strings.Index(basedn, ",dc=")
	if idxFirstDC > 0 {
		return basedn[idxFirstDC+1:]
	}

	return basedn
}

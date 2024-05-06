package dbutils

type LDAPConnConfig struct {
	LDAPURI                string
	LDAPSuffix             string // dc=xx
	LDAPBaseDN             string // o=domains,dc=xx
	LDAPDomainAdminsBaseDN string // o=domainAdmins,dc=xx,dc=xx
	LDAPBindDN             string
	LDAPBindPassword       string
	LDAPStartTLS           bool
}

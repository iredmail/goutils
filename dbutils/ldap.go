package dbutils

type LDAPConnConfig struct {
	URI                string
	BaseDN             string // o=domains,dc=xx
	DomainAdminsBaseDN string // o=domainAdmins,dc=xx,dc=xx
	BindDN             string
	BindPassword       string
	StartTLS           bool
}

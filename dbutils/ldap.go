package dbutils

type LDAPConnConfig struct {
	URI                string
	BaseDN             string // o=domains,dc=xx
	DomainAdminsBaseDN string // o=domainAdmins,dc=xx,dc=xx
	BindDN             string
	BindPassword       string
	StartTLS           bool

	// MaxConnections 定义 ldap 连接池维持的最大连接数。默认为 10。
	MaxConnections int
}

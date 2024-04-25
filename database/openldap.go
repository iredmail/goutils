package database

import (
	"crypto/tls"

	"github.com/iredmail/ldappool"
)

type LDAPConfig struct {
	URI                string
	Suffix             string // dc=xx
	BaseDN             string // o=domains,dc=xx
	DomainAdminsBaseDN string // o=domainAdmins,dc=xx,dc=xx
	BindDN             string
	BindPassword       string
	StartTLS           bool
}

type LDAP[C Config, T Conn] struct {
	LDAPConfig
}

func (l *LDAP[C, T]) GetDBType() DBType {
	return DBTypeLDAP
}

func (l *LDAP[C, T]) GetConfig() C {
	return any(l.LDAPConfig)
}

func (l *LDAP[C, T]) Connect() (conn T, err error) {
	opts := []ldappool.Option{
		ldappool.WithBindCredentials(l.BindDN, l.BindPassword),
	}

	if l.StartTLS {
		opts = append(opts, ldappool.WithTLSConfig(&tls.Config{InsecureSkipVerify: true}))
	}

	pool, err := ldappool.New(l.URI, opts...)
	if err != nil {
		return
	}

	conn = any(pool)

	return
}

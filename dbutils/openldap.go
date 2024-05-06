package dbutils

import (
	"crypto/tls"

	"github.com/go-ldap/ldap/v3"
	"github.com/iredmail/ldappool"
)

func NewLDAP(c LDAPConfig) (pool ldap.Client, err error) {
	opts := []ldappool.Option{
		ldappool.WithBindCredentials(c.LDAPBindDN, c.LDAPBindPassword),
	}

	if c.LDAPStartTLS {
		opts = append(opts, ldappool.WithTLSConfig(&tls.Config{InsecureSkipVerify: true}))
	}

	pool, err = ldappool.New(c.LDAPURI, opts...)

	return
}

type LDAPConfig struct {
	LDAPURI                string
	LDAPSuffix             string // dc=xx
	LDAPBaseDN             string // o=domains,dc=xx
	LDAPDomainAdminsBaseDN string // o=domainAdmins,dc=xx,dc=xx
	LDAPBindDN             string
	LDAPBindPassword       string
	LDAPStartTLS           bool
}

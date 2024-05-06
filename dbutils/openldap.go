package dbutils

import (
	"crypto/tls"

	"github.com/go-ldap/ldap/v3"
	"github.com/iredmail/ldappool"
)

func NewOpenLDAPConn(c LDAPConnConfig) (pool ldap.Client, err error) {
	opts := []ldappool.Option{
		ldappool.WithBindCredentials(c.LDAPBindDN, c.LDAPBindPassword),
	}

	if c.LDAPStartTLS {
		opts = append(opts, ldappool.WithTLSConfig(&tls.Config{InsecureSkipVerify: true}))
	}

	pool, err = ldappool.New(c.LDAPURI, opts...)

	return
}

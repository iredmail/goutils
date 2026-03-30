package dbutils

import (
	"crypto/tls"

	"github.com/iredmail/ldappool"
)

func NewOpenLDAPConn(c LDAPConnConfig) (pool *ldappool.Pool, err error) {
	opts := []ldappool.Option{
		ldappool.WithBindCredentials(c.BindDN, c.BindPassword),
	}

	if c.StartTLS {
		opts = append(opts, ldappool.WithTLSConfig(&tls.Config{InsecureSkipVerify: true}))
	}

	pool, err = ldappool.New(c.URI, opts...)

	return
}

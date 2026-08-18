package dbutils

import (
	"crypto/tls"
	"time"

	"github.com/iredmail/ldappool"
)

func NewOpenLDAPConn(c LDAPConnConfig) (pool *ldappool.Pool, err error) {
	opts := []ldappool.Option{
		ldappool.WithBindCredentials(c.BindDN, c.BindPassword),
	}

	if c.MaxConnections > 0 {
		opts = append(opts,
			ldappool.WithMaxConnections(c.MaxConnections),
		)
	}

	if c.ConnTimeout > 0 {
		opts = append(opts,
			ldappool.WithTimeout(time.Duration(c.ConnTimeout)*time.Second),
		)
	}

	if c.StartTLS {
		opts = append(opts, ldappool.WithTLSConfig(
			&tls.Config{
				InsecureSkipVerify: true,
			},
		))
	}

	pool, err = ldappool.New(c.URI, opts...)

	return
}

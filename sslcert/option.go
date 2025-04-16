package sslcert

import (
	"database/sql"
	"strings"

	"github.com/iredmail/goutils/emailutils"
)

type Option func(m *Manager)

func WithCertDomain(domains ...string) Option {
	return func(m *Manager) {
		for _, domain := range domains {
			if emailutils.IsDomain(domain) {
				m.certDomains = append(m.certDomains, strings.ToLower(domain))
			}
		}
	}
}

func WithDirCache(dir string) Option {
	return func(m *Manager) {
		m.cacheDir = dir
	}
}

func WithSQLiteCache(conn *sql.DB, tableName string) Option {
	cache, err := NewSQLiteCache(conn, tableName)
	if err != nil {
		panic(err)
	}

	return func(m *Manager) {
		m.autocertMgr.Cache = cache
	}
}

func WithSSLFile(certFile, keyFile string) Option {
	return func(m *Manager) {
		m.sslCertFile = certFile
		m.sslKeyFile = keyFile
	}
}

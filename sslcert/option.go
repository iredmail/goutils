package sslcert

import (
	"time"
)

type Option func(m *Manager)

func WithCertDomain(domains ...string) Option {
	return func(m *Manager) {
		m.certDomains = domains
	}
}

func WithAutoCertCacheDir(dir string) Option {
	return func(m *Manager) {
		m.autoCertCacheDir = dir
	}
}

func WithSSLFile(certFile, keyFile string) Option {
	return func(m *Manager) {
		m.sslCertFile = certFile
		m.sslKeyFile = keyFile
	}
}

func WithRenewBefore(d time.Duration) Option {
	return func(m *Manager) {
		m.autocertMgr.RenewBefore = d
	}
}

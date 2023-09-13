package sslcert

import (
	"github.com/iredmail/goutils/emailutils"
)

type Option func(m *Manager)

func WithCertDomain(domains ...string) Option {
	return func(m *Manager) {
		for _, domain := range domains {
			if emailutils.IsDomain(domain) {
				m.certDomains = append(m.certDomains, domain)
			}
		}
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

package sslcert

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

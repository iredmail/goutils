package sslcert

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net"
	"os"

	"github.com/iredmail/goutils"
	"golang.org/x/crypto/acme/autocert"
)

// New 初始化 ssl cert，一共分为两种模式：
// - 使用固定文件（cfg.SSLKeyFile, cfg.SSLCertFile）
// - autocert
func New(options ...Option) (m *Manager, err error) {
	m = &Manager{
		// 不管是否有 cert/key 文件，确保 `autocertMgr` 指针不为 nil。否则会触发 panic。
		autocertMgr: &autocert.Manager{
			// Client: &acme.Client{
			//	//使用 Let's Encrypt 的测试服务器
			//	DirectoryURL: "https://acme-staging-v02.api.letsencrypt.org/directory",
			// },
			// RenewBefore: cfg.AutocertRenewBefore,
			Prompt: autocert.AcceptTOS,
		},
	}

	for _, option := range options {
		option(m)
	}

	m.autocertMgr.Cache = autocert.DirCache(m.autoCertCacheDir)

	// 加载 /opt/spider/{cert,key}.pem
	if goutils.DestExists(m.sslCertFile) && goutils.DestExists(m.sslKeyFile) {
		cert, err := tls.LoadX509KeyPair(m.sslCertFile, m.sslKeyFile)
		if err != nil {
			err = fmt.Errorf("failed in initializing ssl certificate: %v", err)

			return
		}

		m.FixedCert = &cert

		// 注意 tls.LoadX509KeyPair 方法返回的 Leaf 对象为空，需要手动加载证书信息
		certRaw, err := os.ReadFile(m.sslCertFile)
		if err != nil {
			return
		}

		block, _ := pem.Decode(certRaw)
		certificate, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return
		}

		m.FixedCert.Leaf = certificate
		m.certDomains = cert.Leaf.DNSNames

		return
	}

	// 尝试创建 cache 目录
	if err = goutils.CreateDirIfNotExist(m.autoCertCacheDir, 0700); err != nil {
		err = fmt.Errorf("failed in creating autocert cache directory: %s, %v", m.autoCertCacheDir, err)

		return
	}

	// 不支持空域名使用 autocert
	if len(m.certDomains) == 0 {
		return
	}

	m.autocertMgr.HostPolicy = autocert.HostWhitelist(m.certDomains...)
	m.IsAutocert = true

	return
}

type Manager struct {
	// cert 是加载的 ssl cert/key 文件，或 autocert。
	//
	// 加载顺序和优先级：
	//
	// 1. /opt/spider/{cert,key}.pem
	// 2. autocert
	//

	IsAutocert bool // 是否使用 autocert 生成和管理证书

	FixedCert   *tls.Certificate
	autocertMgr *autocert.Manager

	autoCertCacheDir string
	certDomains      []string
	sslCertFile      string
	sslKeyFile       string
}

// Certificate 在存储证书的目录中查找包含主机名列表的证书
// key autocert.Cache 接口中以证书文件名作为 key 来获取证书
func (m *Manager) Certificate(key string) (*x509.Certificate, error) {
	// 使用固定的证书
	if m.FixedCert != nil {
		return m.FixedCert.Leaf, nil
	}

	// 如果 autocert 实例为空，返回空证书
	if !m.IsAutocert {
		return &x509.Certificate{}, nil
	}

	// 获取证书文件裸数据
	certRaw, err := m.autocertMgr.Cache.Get(context.Background(), key)

	// 未找到对应域名的缓存证书则返回空数据
	if err != nil {
		if errors.Is(err, autocert.ErrCacheMiss) {
			return &x509.Certificate{}, nil
		}

		return nil, err
	}

	_, pub := pem.Decode(certRaw)
	pubBlock, _ := pem.Decode(pub)

	return x509.ParseCertificate(pubBlock.Bytes)
}

// TLSConfig 根据当前 ssl cert 配置返回对应的 tls.Config
func (m *Manager) TLSConfig(minVersion uint16, cipherSuites []uint16) (tc *tls.Config) {
	if minVersion == 0 {
		minVersion = tls.VersionTLS12
	}

	tc = &tls.Config{
		MinVersion:     minVersion,
		CipherSuites:   cipherSuites,
		GetCertificate: m.GetCertificate,
	}

	return
}

func (m *Manager) GetCertificate(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	// postfix smtp 使用 tls 进行连接时 ServerName 会为空
	// 需要根据当前的证书的域名设置正确的 ServerName
	if hello.ServerName == "" && len(m.certDomains) > 0 {
		hello.ServerName = m.certDomains[0]
	}

	// 存在使用固定证书
	if m.FixedCert != nil {
		return m.FixedCert, nil
	}

	return m.autocertMgr.GetCertificate(hello)
}

func (m *Manager) Listener(addr string) (net.Listener, error) {
	var tc *tls.Config
	if m.FixedCert != nil {
		tc = &tls.Config{
			Certificates: []tls.Certificate{*m.FixedCert},
		}
	} else {
		tc = &tls.Config{
			GetCertificate: m.autocertMgr.GetCertificate,
			// By default NextProtos contains the "h2"
			// This has to be removed since Fasthttp does not support HTTP/2
			// Or it will cause a flood of PRI method logs
			// http://webconcepts.info/concepts/http-method/PRI
			NextProtos: []string{
				"http/1.1", "acme-tls/1",
			},
		}

		m.autocertMgr.TLSConfig()
	}

	return tls.Listen("tcp", addr, tc)
}

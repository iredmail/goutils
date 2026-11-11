package sslcert_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iredmail/goutils/sslcert"
)

func newAutoCertManager(t *testing.T) *sslcert.Manager {
	t.Helper()

	m, err := sslcert.New(sslcert.WithCertDomain("example.com"))
	require.NoError(t, err)

	return m
}

func TestHTTPHandlerRedirectsToHTTPS(t *testing.T) {
	handler := newAutoCertManager(t).HTTPHandler(nil)

	tests := []struct {
		name         string
		method       string
		host         string
		target       string
		wantCode     int
		wantLocation string
	}{
		{
			name:         "root",
			method:       http.MethodGet,
			target:       "http://example.com/",
			wantCode:     http.StatusFound,
			wantLocation: "https://example.com/",
		},
		{
			name:         "path and query",
			method:       http.MethodGet,
			target:       "http://example.com/a/b?c=d",
			wantCode:     http.StatusFound,
			wantLocation: "https://example.com/a/b?c=d",
		},
		{
			name:         "head",
			method:       http.MethodHead,
			target:       "http://example.com/a",
			wantCode:     http.StatusFound,
			wantLocation: "https://example.com/a",
		},
		{
			// autocert 内置逻辑不对 Host 做规范化，带端口时会显式补上 :443。
			name:         "host with port",
			method:       http.MethodGet,
			host:         "example.com:80",
			target:       "http://example.com/a",
			wantCode:     http.StatusFound,
			wantLocation: "https://example.com:443/a",
		},
		{
			// autocert 内置逻辑原样保留 Host 的大小写与结尾的点。
			name:         "uppercase host and trailing dot",
			method:       http.MethodGet,
			host:         "EXAMPLE.COM.",
			target:       "http://example.com/a",
			wantCode:     http.StatusFound,
			wantLocation: "https://EXAMPLE.COM./a",
		},
		{
			name:     "post",
			method:   http.MethodPost,
			target:   "http://example.com/a",
			wantCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.target, nil)
			if tt.host != "" {
				req.Host = tt.host
			}
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantCode, rec.Code)
			assert.Equal(t, tt.wantLocation, rec.Header().Get("Location"))
		})
	}
}

func TestHTTPHandlerAcmeChallenge(t *testing.T) {
	handler := newAutoCertManager(t).HTTPHandler(nil)

	tests := []struct {
		name     string
		host     string
		wantCode int
	}{
		{
			name:     "unknown token",
			host:     "example.com",
			wantCode: http.StatusNotFound,
		},
		{
			name:     "host not in whitelist",
			host:     "evil.com",
			wantCode: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "http://"+tt.host+"/.well-known/acme-challenge/unknown", nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantCode, rec.Code)
		})
	}
}

func TestHTTPHandlerFallback(t *testing.T) {
	handler := newAutoCertManager(t).HTTPHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}))

	req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusTeapot, rec.Code)
}

func TestHTTPHandlerWithoutAutoCert(t *testing.T) {
	m, err := sslcert.New()
	require.NoError(t, err)

	handler := m.HTTPHandler(nil)

	// 未配置自动证书时，非挑战请求仍由 autocert 内置逻辑重定向。
	req := httptest.NewRequest(http.MethodGet, "http://example.com/a", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusFound, rec.Code)
	assert.Equal(t, "https://example.com/a", rec.Header().Get("Location"))

	// 挑战路径仍由 autocert 处理，没有 token 时返回 404。
	req = httptest.NewRequest(http.MethodGet, "http://example.com/.well-known/acme-challenge/unknown", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestHTTPHandlerFixedCert(t *testing.T) {
	certFile, keyFile := writeTestCertFiles(t)

	m, err := sslcert.New(sslcert.WithSSLFile(certFile, keyFile))
	require.NoError(t, err)
	assert.False(t, m.IsAutocert)

	handler := m.HTTPHandler(nil)

	// 固定证书模式下，非挑战请求仍由 autocert 内置逻辑重定向。
	req := httptest.NewRequest(http.MethodGet, "http://example.com/a", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusFound, rec.Code)
	assert.Equal(t, "https://example.com/a", rec.Header().Get("Location"))

	// 挑战路径仍由 autocert 处理，没有 token 时返回 404。
	req = httptest.NewRequest(http.MethodGet, "http://example.com/.well-known/acme-challenge/token", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func writeTestCertFiles(t *testing.T) (certFile, keyFile string) {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "example.com"},
		DNSNames:     []string{"example.com"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	require.NoError(t, err)

	keyDER, err := x509.MarshalECPrivateKey(key)
	require.NoError(t, err)

	dir := t.TempDir()
	certFile = filepath.Join(dir, "cert.pem")
	keyFile = filepath.Join(dir, "key.pem")

	require.NoError(t, os.WriteFile(certFile,
		pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER}), 0600))
	require.NoError(t, os.WriteFile(keyFile,
		pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER}), 0600))

	return certFile, keyFile
}

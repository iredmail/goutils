package dnsutils

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testResolverFactory struct {
	name string
	new  func(addr string) Resolver
}

func TestLookupSPF(t *testing.T) {
	addr := startTestDNSServer(t, dnsTestResponder)

	for _, factory := range testResolverFactories(t) {
		t.Run(factory.name, func(t *testing.T) {
			notfound, records, errText := factory.new(addr).LookupSPF("gmail.test")
			assert.Empty(t, errText)
			assert.False(t, notfound)
			assert.Equal(t, []string{"v=spf1 redirect=_spf.google.test"}, records)
		})
	}
}

func TestLookupRecursiveSPF(t *testing.T) {
	addr := startTestDNSServer(t, dnsTestResponder)

	for _, factory := range testResolverFactories(t) {
		t.Run(factory.name, func(t *testing.T) {
			notfound, records, totalQueries, errText := factory.new(addr).LookupRecursiveSPF("gmail.test", 0)
			assert.Empty(t, errText)
			assert.False(t, notfound)
			assert.Equal(t, 2, totalQueries)
			assert.Equal(t, []string{"v=spf1 redirect=_spf.google.test"}, records)
		})
	}
}

func TestLookupTXTRecordFilteringAndNotfound(t *testing.T) {
	addr := startTestDNSServer(t, dnsTestResponder)

	for _, factory := range testResolverFactories(t) {
		t.Run(factory.name, func(t *testing.T) {
			notfound, records, errText := factory.new(addr).LookupDKIM("split.test", "selector")
			assert.Empty(t, errText)
			assert.False(t, notfound)
			assert.Equal(t, []string{"v=DKIM1; k=rsa; p=abcdef"}, records)

			notfound, records, errText = factory.new(addr).LookupDKIM("nomatch.test", "selector")
			assert.Empty(t, errText)
			assert.True(t, notfound)
			assert.Empty(t, records)

			notfound, records, errText = factory.new(addr).LookupDMARC("split.test")
			assert.Empty(t, errText)
			assert.False(t, notfound)
			assert.Equal(t, []string{"v=DMARC1; p=none; rua=mailto:d@example.com"}, records)

			notfound, records, errText = factory.new(addr).LookupDMARC("nomatch.test")
			assert.Empty(t, errText)
			assert.True(t, notfound)
			assert.Empty(t, records)

			notfound, records, errText = factory.new(addr).LookupSPF("spf-missing.test")
			assert.Empty(t, errText)
			assert.True(t, notfound)
			assert.Empty(t, records)
		})
	}
}

func TestCustomResolverExchangeHandlesRcodeAndTruncation(t *testing.T) {
	addr := startTestDNSServer(t, dnsTestResponder)
	r := NewResolver(2, addr)

	notfound, records, errText := r.LookupSPF("servfail.test")
	assert.False(t, notfound)
	assert.Empty(t, records)
	assert.Contains(t, errText, "SERVFAIL")

	notfound, records, errText = r.LookupDMARC("nodata.test")
	assert.Empty(t, errText)
	assert.True(t, notfound)
	assert.Empty(t, records)

	notfound, records, errText = r.LookupSPF("truncated.test")
	assert.Empty(t, errText)
	assert.False(t, notfound)
	assert.Equal(t, []string{"v=spf1 include:mail.truncated.test ~all"}, records)
}

func TestLookupRecursiveSPFCountsMXQueriesAndLimit(t *testing.T) {
	addr := startTestDNSServer(t, dnsTestResponder)

	for _, factory := range testResolverFactories(t) {
		t.Run(factory.name, func(t *testing.T) {
			notfound, records, totalQueries, errText := factory.new(addr).LookupRecursiveSPF("mx2.test", 0)
			assert.Empty(t, errText)
			assert.False(t, notfound)
			assert.Equal(t, []string{"v=spf1 mx -all"}, records)
			assert.Equal(t, 4, totalQueries)

			notfound, records, totalQueries, errText = factory.new(addr).LookupRecursiveSPF("mx10.test", 0)
			assert.Empty(t, errText)
			assert.False(t, notfound)
			assert.Equal(t, []string{"v=spf1 mx -all"}, records)
			assert.Equal(t, 10, totalQueries)
		})
	}
}

func TestLookupRecursiveSPFCountsExistsQueriesAndLimit(t *testing.T) {
	addr := startTestDNSServer(t, dnsTestResponder)

	for _, factory := range testResolverFactories(t) {
		t.Run(factory.name, func(t *testing.T) {
			notfound, records, totalQueries, errText := factory.new(addr).LookupRecursiveSPF("exists2.test", 0)
			assert.Empty(t, errText)
			assert.False(t, notfound)
			assert.Equal(t, []string{"v=spf1 exists:one.exists.test exists:two.exists.test -all"}, records)
			assert.Equal(t, 3, totalQueries)

			notfound, records, totalQueries, errText = factory.new(addr).LookupRecursiveSPF("exists10.test", 0)
			assert.Empty(t, errText)
			assert.False(t, notfound)
			assert.Equal(t, []string{"v=spf1 exists:e01.exists.test exists:e02.exists.test exists:e03.exists.test exists:e04.exists.test exists:e05.exists.test exists:e06.exists.test exists:e07.exists.test exists:e08.exists.test exists:e09.exists.test exists:e10.exists.test -all"}, records)
			assert.Equal(t, 10, totalQueries)
		})
	}
}

func testResolverFactories(t *testing.T) []testResolverFactory {
	t.Helper()
	t.Setenv("GODEBUG", "netdns=go")

	return []testResolverFactory{
		{
			name: "defaultResolver",
			new: func(addr string) Resolver {
				return &defaultResolver{
					resolver: &net.Resolver{
						PreferGo: true,
						Dial: func(ctx context.Context, network, _ string) (net.Conn, error) {
							var d net.Dialer

							return d.DialContext(ctx, network, addr)
						},
					},
					timeout: 2 * time.Second,
				}
			},
		},
		{
			name: "customResolver",
			new: func(addr string) Resolver {
				return NewResolver(2, addr)
			},
		},
	}
}

func startTestDNSServer(t *testing.T, responder func(m *dns.Msg, q dns.Question, network string)) string {
	t.Helper()

	udpConn, err := net.ListenPacket("udp", "127.0.0.1:0")
	require.NoError(t, err)

	host, port, err := net.SplitHostPort(udpConn.LocalAddr().String())
	require.NoError(t, err)

	tcpListener, err := net.Listen("tcp", net.JoinHostPort(host, port))
	require.NoError(t, err)

	handler := func(network string) dns.HandlerFunc {
		return func(w dns.ResponseWriter, req *dns.Msg) {
			m := new(dns.Msg)
			m.SetReply(req)
			m.Authoritative = true
			m.RecursionAvailable = true
			responder(m, req.Question[0], network)

			_ = w.WriteMsg(m)
		}
	}

	udpServer := &dns.Server{PacketConn: udpConn, Handler: handler("udp")}
	tcpServer := &dns.Server{Listener: tcpListener, Handler: handler("tcp")}

	go func() { _ = udpServer.ActivateAndServe() }()
	go func() { _ = tcpServer.ActivateAndServe() }()

	t.Cleanup(func() {
		_ = udpServer.Shutdown()
		_ = tcpServer.Shutdown()
	})

	return udpConn.LocalAddr().String()
}

func dnsTestResponder(m *dns.Msg, q dns.Question, network string) {
	switch {
	case q.Name == "gmail.test." && q.Qtype == dns.TypeTXT:
		m.Answer = []dns.RR{&dns.TXT{
			Hdr: dns.RR_Header{Name: q.Name, Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 60},
			Txt: []string{"v=spf1 redirect=_spf.", "google.test"},
		}}
	case q.Name == "mx2.test." && q.Qtype == dns.TypeTXT:
		m.Answer = []dns.RR{&dns.TXT{
			Hdr: dns.RR_Header{Name: q.Name, Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 60},
			Txt: []string{"v=spf1 mx -all"},
		}}
	case q.Name == "mx10.test." && q.Qtype == dns.TypeTXT:
		m.Answer = []dns.RR{&dns.TXT{
			Hdr: dns.RR_Header{Name: q.Name, Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 60},
			Txt: []string{"v=spf1 mx -all"},
		}}
	case q.Name == "exists2.test." && q.Qtype == dns.TypeTXT:
		m.Answer = []dns.RR{&dns.TXT{
			Hdr: dns.RR_Header{Name: q.Name, Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 60},
			Txt: []string{"v=spf1 exists:one.exists.test exists:two.exists.test -all"},
		}}
	case q.Name == "exists10.test." && q.Qtype == dns.TypeTXT:
		m.Answer = []dns.RR{&dns.TXT{
			Hdr: dns.RR_Header{Name: q.Name, Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 60},
			Txt: []string{"v=spf1 exists:e01.exists.test exists:e02.exists.test exists:e03.exists.test exists:e04.exists.test exists:e05.exists.test exists:e06.exists.test exists:e07.exists.test exists:e08.exists.test exists:e09.exists.test exists:e10.exists.test -all"},
		}}
	case q.Name == "_spf.google.test." && q.Qtype == dns.TypeTXT:
		m.Answer = []dns.RR{&dns.TXT{
			Hdr: dns.RR_Header{Name: q.Name, Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 60},
			Txt: []string{"v=spf1 ip4:1.2.3.4 -all"},
		}}
	case q.Name == "mx2.test." && q.Qtype == dns.TypeMX:
		m.Answer = []dns.RR{
			&dns.MX{Hdr: dns.RR_Header{Name: q.Name, Rrtype: dns.TypeMX, Class: dns.ClassINET, Ttl: 60}, Preference: 10, Mx: "mx1.mx2.test."},
			&dns.MX{Hdr: dns.RR_Header{Name: q.Name, Rrtype: dns.TypeMX, Class: dns.ClassINET, Ttl: 60}, Preference: 20, Mx: "mx2.mx2.test."},
		}
	case q.Name == "mx10.test." && q.Qtype == dns.TypeMX:
		m.Answer = []dns.RR{
			&dns.MX{Hdr: dns.RR_Header{Name: q.Name, Rrtype: dns.TypeMX, Class: dns.ClassINET, Ttl: 60}, Preference: 10, Mx: "mx01.mx10.test."},
			&dns.MX{Hdr: dns.RR_Header{Name: q.Name, Rrtype: dns.TypeMX, Class: dns.ClassINET, Ttl: 60}, Preference: 11, Mx: "mx02.mx10.test."},
			&dns.MX{Hdr: dns.RR_Header{Name: q.Name, Rrtype: dns.TypeMX, Class: dns.ClassINET, Ttl: 60}, Preference: 12, Mx: "mx03.mx10.test."},
			&dns.MX{Hdr: dns.RR_Header{Name: q.Name, Rrtype: dns.TypeMX, Class: dns.ClassINET, Ttl: 60}, Preference: 13, Mx: "mx04.mx10.test."},
			&dns.MX{Hdr: dns.RR_Header{Name: q.Name, Rrtype: dns.TypeMX, Class: dns.ClassINET, Ttl: 60}, Preference: 14, Mx: "mx05.mx10.test."},
			&dns.MX{Hdr: dns.RR_Header{Name: q.Name, Rrtype: dns.TypeMX, Class: dns.ClassINET, Ttl: 60}, Preference: 15, Mx: "mx06.mx10.test."},
			&dns.MX{Hdr: dns.RR_Header{Name: q.Name, Rrtype: dns.TypeMX, Class: dns.ClassINET, Ttl: 60}, Preference: 16, Mx: "mx07.mx10.test."},
			&dns.MX{Hdr: dns.RR_Header{Name: q.Name, Rrtype: dns.TypeMX, Class: dns.ClassINET, Ttl: 60}, Preference: 17, Mx: "mx08.mx10.test."},
			&dns.MX{Hdr: dns.RR_Header{Name: q.Name, Rrtype: dns.TypeMX, Class: dns.ClassINET, Ttl: 60}, Preference: 18, Mx: "mx09.mx10.test."},
			&dns.MX{Hdr: dns.RR_Header{Name: q.Name, Rrtype: dns.TypeMX, Class: dns.ClassINET, Ttl: 60}, Preference: 19, Mx: "mx10.mx10.test."},
		}
	case q.Name == "_dmarc.split.test." && q.Qtype == dns.TypeTXT:
		m.Answer = []dns.RR{&dns.TXT{
			Hdr: dns.RR_Header{Name: q.Name, Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 60},
			Txt: []string{"v=DMARC1; p=none; rua=mailto:d@", "example.com"},
		}}
	case q.Name == "selector._domainkey.split.test." && q.Qtype == dns.TypeTXT:
		m.Answer = []dns.RR{&dns.TXT{
			Hdr: dns.RR_Header{Name: q.Name, Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 60},
			Txt: []string{"v=DKIM1; k=rsa; p=abc", "def"},
		}}
	case q.Name == "selector._domainkey.nomatch.test." && q.Qtype == dns.TypeTXT:
		m.Answer = []dns.RR{&dns.TXT{
			Hdr: dns.RR_Header{Name: q.Name, Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 60},
			Txt: []string{"google-site-verification=abc"},
		}}
	case q.Name == "_dmarc.nodata.test." && q.Qtype == dns.TypeTXT:
		m.Ns = []dns.RR{testSOA(q.Name)}
	case q.Name == "_dmarc.nomatch.test." && q.Qtype == dns.TypeTXT:
		m.Answer = []dns.RR{&dns.TXT{
			Hdr: dns.RR_Header{Name: q.Name, Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 60},
			Txt: []string{"google-site-verification=abc"},
		}}
	case q.Name == "spf-missing.test." && q.Qtype == dns.TypeTXT:
		m.Answer = []dns.RR{&dns.TXT{
			Hdr: dns.RR_Header{Name: q.Name, Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 60},
			Txt: []string{"google-site-verification=abc"},
		}}
	case q.Name == "servfail.test." && q.Qtype == dns.TypeTXT:
		m.Rcode = dns.RcodeServerFailure
	case q.Name == "truncated.test." && q.Qtype == dns.TypeTXT:
		if network == "udp" {
			m.Truncated = true

			return
		}

		m.Answer = []dns.RR{&dns.TXT{
			Hdr: dns.RR_Header{Name: q.Name, Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 60},
			Txt: []string{"v=spf1 include:mail.", "truncated.test ~all"},
		}}
	default:
		m.Rcode = dns.RcodeNameError
		m.Ns = []dns.RR{testSOA(q.Name)}
	}
}

func testSOA(name string) dns.RR {
	return &dns.SOA{
		Hdr:     dns.RR_Header{Name: name, Rrtype: dns.TypeSOA, Class: dns.ClassINET, Ttl: 60},
		Ns:      fmt.Sprintf("ns1.%s", name),
		Mbox:    fmt.Sprintf("hostmaster.%s", name),
		Serial:  1,
		Refresh: 3600,
		Retry:   600,
		Expire:  86400,
		Minttl:  60,
	}
}

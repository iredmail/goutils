package dnsutils

import (
	"net"
	"regexp"
	"time"

	"github.com/miekg/dns"
)

var (
	defaultDNSQueryTimeout = 10 * time.Second

	// 正则表达式，用于匹配 SPF、DKIM、DMARC 记录。不区分大小写。
	regxSPF   = regexp.MustCompile(`(?i)^v=spf1`)
	regxDKIM  = regexp.MustCompile(`(?i)^v=DKIM1;`)
	regxDMARC = regexp.MustCompile(`(?i)^v=DMARC1;`)
)

const (
	spfDNSQueryTypeA   uint16 = 1  // RFC 1035: A
	spfDNSQueryTypeMX  uint16 = 15 // RFC 1035: MX
	spfDNSQueryTypePTR uint16 = 12 // RFC 1035: PTR
	// exists 不是实际 RR type，只用于 LookupRecursiveSPF 内部标记“会触发一次 DNS 查询”。
	spfDNSQueryTypeExists uint16 = 0
)

func NewResolver(queryTimeoutSeconds int, dnsAddr ...string) Resolver {
	timeout := defaultDNSQueryTimeout
	if queryTimeoutSeconds > 0 {
		timeout = time.Duration(queryTimeoutSeconds) * time.Second
	}

	if len(dnsAddr) > 0 && dnsAddr[0] != "" {
		return &customResolver{
			client:  &dns.Client{Timeout: timeout},
			dnsAddr: dnsAddr[0],
		}
	}

	return &defaultResolver{resolver: net.DefaultResolver, timeout: timeout}
}

type Resolver interface {
	LookupHost(domain string) (notfound bool, ip4s, ip6s []string, errText string)
	LookupA(domain string) (notfound bool, ip4s []string, errText string)
	LookupAAAA(domain string) (notfound bool, ip6s []string, errText string)
	LookupMX(domain string) (notfound bool, records []MXRecord, errText string)
	LookupDKIM(domain, selector string) (notfound bool, records []string, errText string)
	LookupDMARC(domain string) (notfound bool, records []string, errText string)
	LookupSPF(domain string) (notfound bool, records []string, errText string)
	LookupRecursiveSPF(domain string, _totalQueries int, dnsType ...uint16) (notfound bool, spf []string, totalQueries int, errText string)
	LookupSRV(domain, dnsTypeStr string) (notfound bool, records []SRVRecord, errText string)
	LookupPtr(ip string) (notfound bool, records []string, errText string)
}

type MXRecord struct {
	MX       string `json:"mx"`
	Priority uint16 `json:"priority"`
}

type SRVRecord struct {
	Priority uint16 `json:"priority,omitempty"`
	Port     uint16 `json:"port,omitempty"`
	Weight   uint16 `json:"weight"`
	Target   string `json:"target"`
}

type ResponseDNSRecords[T any] struct {
	Domain       string `json:"domain"`
	Notfound     bool   `json:"notfound"`
	TotalQueries int    `json:"total_queries,omitempty"`
	Records      []T    `json:"records"`
	Error        string `json:"error"`
}

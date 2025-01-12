package dnsutils

import (
	"errors"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/miekg/dns"
	"golang.org/x/exp/rand"
)

var (
	defaultTimeout = 10 * time.Second
	defaultServers = []string{
		"8.8.8.8:53",
		"8.8.4.4:53",
		"1.1.1.1:53",
	}

	// 将 SPF 记录里的所有域名解析成 IP 地址的查询次数。
	// RFC 文档规定不能超过 10 次，这里略微调大一些，查询 30 次，以避免无休止地递归查询。
	defaultMaxSPFQueries = 30

	// 正则表达式，用于匹配 SPF、DKIM、DMARC 记录。不区分大小写。
	regxSPF   = regexp.MustCompile(`(?i)^v=spf1`)
	regxDKIM  = regexp.MustCompile(`(?i)^v=DKIM1;`)
	regxDMARC = regexp.MustCompile(`(?i)^v=DMARC1;`)
)

func getRandomServer() string {
	rand.Seed(uint64(time.Now().UnixNano()))
	idx := rand.Intn(len(defaultServers))

	return defaultServers[idx]
}

func newClientAndMsg() (*dns.Client, *dns.Msg) {
	client := new(dns.Client)
	client.Timeout = 10 * time.Second

	msg := new(dns.Msg)

	return client, msg
}

func queryDomain(domain string, dnsType uint16, dnsServers ...string) (found bool, answers []dns.RR, rtt time.Duration, err error) {
	domain = strings.ToLower(domain)

	client, msg := newClientAndMsg()
	msg.SetQuestion(dns.Fqdn(domain), dnsType)
	msg.RecursionDesired = true

	var srv string
	if len(dnsServers) > 0 {
		srv = dnsServers[0]
	} else {
		srv = getRandomServer()
	}

	r, _, err := client.Exchange(msg, srv)
	if err != nil {
		return
	}

	if r.Rcode != dns.RcodeSuccess {
		return
	}

	found = true
	answers = r.Answer

	return
}

func QueryA(domain string, dnsServers ...string) (found bool, result ResultA, err error) {
	found, answers, rtt, err := queryDomain(domain, dns.TypeA, dnsServers...)
	if err != nil || !found {
		return
	}

	result.Domain = domain
	result.RTT = rtt

	for _, ans := range answers {
		if a, ok := ans.(*dns.A); ok {
			result.IPs = append(result.IPs, a.A.String())
		}
	}

	sort.Strings(result.IPs)

	return
}

func QueryAAAA(domain string, dnsServers ...string) (found bool, result ResultAAAA, err error) {
	found, answers, rtt, err := queryDomain(domain, dns.TypeAAAA, dnsServers...)
	if err != nil || !found {
		return
	}

	result.Domain = domain
	result.RTT = rtt

	for _, ans := range answers {
		if a, ok := ans.(*dns.AAAA); ok {
			result.IPs = append(result.IPs, a.AAAA.String())
		}
	}
	sort.Strings(result.IPs)

	return
}

func QueryMX(domain string) (found bool, result ResultMX, err error) {
	found, answers, rtt, err := queryDomain(domain, dns.TypeMX)
	if err != nil || !found {
		return
	}

	result.Domain = domain
	result.RTT = rtt

	var hosts []HostMX
	for _, ans := range answers {
		if mx, ok := ans.(*dns.MX); ok {
			// Remove trailing dot.
			hostname := strings.TrimRight(mx.Mx, ".")

			hosts = append(hosts, HostMX{
				Hostname: hostname,
				TTL:      mx.Hdr.Ttl,
				Priority: mx.Preference,
			})

			result.Hostnames = append(result.Hostnames, hostname)
		}
	}

	sort.Strings(result.Hostnames)

	// 保存排序后的 hosts
	for _, hostname := range result.Hostnames {
		for _, host := range hosts {
			if host.Hostname == hostname {
				result.Hosts = append(result.Hosts, host)
			}
		}
	}

	return
}

func queryTXT(domain string, dnsServers ...string) (found bool, answers []dns.RR, rtt time.Duration, err error) {
	return queryDomain(domain, dns.TypeTXT, dnsServers...)
}

func QueryDKIM(domain, selector string) (result ResultDKIM, err error) {
	if selector == "" {
		err = errors.New("selector is missing")

		return
	}

	found, answers, rtt, err := queryTXT(domain)
	if err != nil || !found {
		return
	}

	result.Domain = domain
	result.RTT = rtt

	// 一个域名一般只有一个 DKIM 记录，这里只取第一个。
	for _, ans := range answers {
		if txt, ok := ans.(*dns.TXT); ok {
			for _, txtStr := range txt.Txt {
				if regxDMARC.MatchString(txtStr) {
					result.DKIM = txtStr
					result.TTL = txt.Hdr.Ttl

					break
				}
			}
		}
	}

	return
}

// func QueryDMARC(domain string) (result ResultDMARC, err error) {}
// func QueryMTASTS(domain string) (result ResultMTASTS, err error) {}
// func QueryTLSRPT(domain string) (result ResultTLSRPT, err error) {}
// func QueryDMARC(domain string) (result ResultDMARC, err error) {}
// func QueryMTASTS(domain string) (result ResultMTASTS, err error) {}
// func QueryTLSRPT(domain string) (result ResultTLSRPT, err error) {}

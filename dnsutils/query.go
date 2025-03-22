package dnsutils

import (
	"regexp"
	"sort"
	"strings"
	"sync"
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

func queryDomain(domain string, dnsType uint16, dnsServers ...string) (found bool, answers []dns.RR, duration time.Duration, err error) {
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

	start := time.Now()
	r, _, err := client.Exchange(msg, srv)
	if err != nil {
		return
	}
	duration = time.Since(start)

	if r.Rcode != dns.RcodeSuccess {
		return
	}

	found = true
	answers = r.Answer

	return
}

func queryTXT(domain string, dnsServers ...string) (found bool, answers []dns.RR, duration time.Duration, err error) {
	return queryDomain(domain, dns.TypeTXT, dnsServers...)
}

// QueryAll 查询指定邮件域名的所有相关 DNS 记录。注意：只返回原始记录，不进行任何处理。
func QueryAll(domain string, dnsServers ...string) (result ResultAll) {
	result.Domain = domain

	var wg sync.WaitGroup
	wg.Add(6) // A, AAAA, MX, SPF, DKIM, DMARC

	go func() {
		defer wg.Done()

		found, resultA := QueryA(domain, dnsServers...)
		if resultA.Error != nil {
			return
		}

		if found {
			result.mu.Lock()
			defer result.mu.Unlock()

			result.ResultA = resultA
			result.Duration += resultA.Duration
		}
	}()

	go func() {
		defer wg.Done()

		found, resultAAAA := QueryAAAA(domain, dnsServers...)
		if resultAAAA.Error != nil {
			return
		}

		if found {
			result.mu.Lock()
			defer result.mu.Unlock()

			result.ResultAAAA = resultAAAA
			result.Duration += resultAAAA.Duration
		}
	}()

	go func() {
		defer wg.Done()

		found, resultMX := QueryMX(domain)
		if resultMX.Error != nil {
			return
		}

		if found {
			result.mu.Lock()
			defer result.mu.Unlock()

			result.ResultMX = resultMX
			result.Duration += resultMX.Duration
		}
	}()

	go func() {
		defer wg.Done()

		found, resultSPF, err := QuerySPF(domain)
		if err != nil {
			return
		}
		if found {
			result.mu.Lock()
			defer result.mu.Unlock()

			result.ResultSPF = resultSPF
			result.Duration += resultSPF.Duration
		}
	}()

	go func() {
		defer wg.Done()

		found, resultDKIM := QueryDKIM(domain, "dkim")
		if resultDKIM.Error != nil {
			return
		}

		if found {
			result.mu.Lock()
			defer result.mu.Unlock()

			result.ResultDKIM = resultDKIM
			result.Duration += resultDKIM.Duration
		}
	}()

	go func() {
		defer wg.Done()

		found, resultDMARC := QueryDMARC(domain)
		if resultDMARC.Error != nil {
			return
		}

		if found {
			result.mu.Lock()
			defer result.mu.Unlock()

			result.ResultDMARC = resultDMARC
			result.Duration += resultDMARC.Duration
		}
	}()

	// TODO MTASTS

	wg.Wait()

	return
}

func QueryA(domain string, dnsServers ...string) (found bool, result ResultA) {
	found, answers, rtt, err := queryDomain(domain, dns.TypeA, dnsServers...)
	if err != nil {
		result.Error = err

		return
	}

	if !found {
		return
	}

	result.Domain = domain
	result.Duration = rtt

	for _, ans := range answers {
		if a, ok := ans.(*dns.A); ok {
			result.IPs = append(result.IPs, a.A.String())
		}
	}

	sort.Strings(result.IPs)

	return
}

func QueryAAAA(domain string, dnsServers ...string) (found bool, result ResultAAAA) {
	found, answers, rtt, err := queryDomain(domain, dns.TypeAAAA, dnsServers...)
	if err != nil {
		result.Error = err

		return
	}

	if !found {
		return
	}

	result.Domain = domain
	result.Duration = rtt

	for _, ans := range answers {
		if a, ok := ans.(*dns.AAAA); ok {
			result.IPs = append(result.IPs, a.AAAA.String())
		}
	}
	sort.Strings(result.IPs)

	return
}

func QueryMX(domain string) (found bool, result ResultMX) {
	found, answers, duration, err := queryDomain(domain, dns.TypeMX)
	if err != nil {
		result.Error = err

		return
	}

	if !found {
		return
	}

	result.Domain = domain
	result.Duration = duration

	var hosts []HostMX
	for _, ans := range answers {
		if mx, ok := ans.(*dns.MX); ok {
			// Remove trailing dot.
			hostname := strings.TrimRight(mx.Mx, ".")

			hostMX := HostMX{
				Hostname: hostname,
				TTL:      mx.Hdr.Ttl,
				Priority: mx.Preference,
			}

			// Resolve to IP addresses.
			foundA, resultA := QueryA(hostname)
			result.Duration += resultA.Duration
			if foundA {
				hostMX.IP4 = resultA.IPs
			}

			foundAAAA, resultAAAA := QueryAAAA(hostname)
			result.Duration += resultAAAA.Duration
			if foundAAAA {
				hostMX.IP6 = resultAAAA.IPs
			}

			hosts = append(hosts, hostMX)
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

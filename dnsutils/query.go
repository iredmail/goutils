package dnsutils

import (
	"strings"
	"time"

	"github.com/miekg/dns"
	"golang.org/x/exp/rand"
)

var defaultDNSServers = []string{
	"8.8.8.8:53",
	"8.8.4.4:53",
	"1.1.1.1:53",
}

func getRandomServer() string {
	rand.Seed(uint64(time.Now().UnixNano()))
	idx := rand.Intn(len(defaultDNSServers))

	return defaultDNSServers[idx]
}

func newClientAndMsg() (*dns.Client, *dns.Msg) {
	client := new(dns.Client)
	msg := new(dns.Msg)

	return client, msg
}

func queryDomain(domain string, dnsType uint16, dnsServers ...string) (found bool, answer []dns.RR, err error) {
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
	answer = r.Answer

	return
}

func QueryA(domain string, dnsServers ...string) (found bool, result ResultA, err error) {
	found, answer, err := queryDomain(domain, dns.TypeA, dnsServers...)
	if err != nil || !found {
		return
	}

	result.Domain = domain

	for _, ans := range answer {
		if a, ok := ans.(*dns.A); ok {
			result.IPs = append(result.IPs, a.A.String())
		}
	}

	return
}

func QueryAAAA(domain string, dnsServers ...string) (found bool, result ResultAAAA, err error) {
	found, answer, err := queryDomain(domain, dns.TypeAAAA, dnsServers...)
	if err != nil || !found {
		return
	}

	result.Domain = domain

	for _, ans := range answer {
		if a, ok := ans.(*dns.AAAA); ok {
			result.IPs = append(result.IPs, a.AAAA.String())
		}
	}

	return
}

// func QueryMX(domain string) (result ResultMX, err error) {}
// func QuerySPF(domain string) (result ResultSPF, err error) {}
// func QueryDKIM(domain string) (result ResultDKIM, err error) {}
// func QueryDMARC(domain string) (result ResultDMARC, err error) {}
// func QueryMTASTS(domain string) (result ResultMTASTS, err error) {}
// func QueryTLSRPT(domain string) (result ResultTLSRPT, err error) {}
// func QueryDMARC(domain string) (result ResultDMARC, err error) {}
// func QueryMTASTS(domain string) (result ResultMTASTS, err error) {}
// func QueryTLSRPT(domain string) (result ResultTLSRPT, err error) {}

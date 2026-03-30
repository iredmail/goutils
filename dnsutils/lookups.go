package dnsutils

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/miekg/dns"

	"github.com/iredmail/goutils/emailutils"
)

const defaultSelector = "dkim"

var (
	defaultDNSQueryTimeout = 10 * time.Second

	// defaultDNSServers 包含一些公共 DNS 服务器的地址，格式为 "IP:Port"。
	// FIXME 用标准库的 `net.LookupNetIP()` 代替 `miekg/dns`。
	defaultDNSServers = []string{
		"8.8.8.8:53",
		"8.8.4.4:53",
		"1.1.1.1:53",
	}

	// 正则表达式，用于匹配 SPF、DKIM、DMARC 记录。不区分大小写。
	regxSPF   = regexp.MustCompile(`(?i)^v=spf1`)
	regxDKIM  = regexp.MustCompile(`(?i)^v=DKIM1;`)
	regxDMARC = regexp.MustCompile(`(?i)^v=DMARC1;`)
)

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

func getRandomDNSServer() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	idx := r.Intn(len(defaultDNSServers))

	return defaultDNSServers[idx]
}

func IsDNSErrorNoSuchHost(err error) (v bool, e string) {
	if err == nil {
		return false, ""
	}

	if _err, ok := errors.AsType[*net.DNSError](err); ok {
		v = _err.Err == "no such host"
		if !v {
			e = err.Error()
		}
	}

	return
}

func LookupA(domain string) (ip4s []string, err error) {
	client := new(dns.Client)
	client.Timeout = defaultDNSQueryTimeout
	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(domain), dns.TypeA)
	msg.RecursionDesired = true

	r, _, err := client.Exchange(msg, getRandomDNSServer())
	if err != nil {
		return
	}

	for _, a := range r.Answer {
		_a, ok := a.(*dns.A)
		if ok {
			ip4s = append(ip4s, _a.A.String())
		}
	}

	return
}

func LookupAAAA(domain string) (ip6s []string, err error) {
	client := new(dns.Client)
	client.Timeout = defaultDNSQueryTimeout
	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(domain), dns.TypeAAAA)
	msg.RecursionDesired = true

	r, _, err := client.Exchange(msg, getRandomDNSServer())
	if err != nil {

		return
	}

	for _, a := range r.Answer {
		_a, ok := a.(*dns.AAAA)
		if ok {
			ip6s = append(ip6s, _a.AAAA.String())
		}
	}

	return
}

func LookupMX(domain string) (notfound bool, records []MXRecord, errStr string) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultDNSQueryTimeout)
	defer cancel()

	var mxs []*net.MX
	mxs, err := net.DefaultResolver.LookupMX(ctx, domain)
	notfound, errStr = IsDNSErrorNoSuchHost(err)
	if notfound || err != nil {
		return
	}

	for _, mx := range mxs {
		records = append(records, MXRecord{
			MX:       strings.TrimSuffix(mx.Host, "."),
			Priority: mx.Pref,
		})
	}

	// Sort by mx priority
	slices.SortFunc(records, func(a, b MXRecord) int {
		return cmp.Compare(a.Priority, b.Priority)
	})

	return
}

func LookupSPF(domain string) (records []string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultDNSQueryTimeout)
	defer cancel()

	var txts []string
	txts, err = net.DefaultResolver.LookupTXT(ctx, domain)
	for _, txt := range txts {
		if regxSPF.MatchString(txt) {
			records = append(records, txt)

			break
		}
	}

	return
}

func LookupDKIM(domain, selector string) (notfound bool, records []string, errStr string) {
	if selector == "" {
		selector = defaultSelector
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultDNSQueryTimeout)
	defer cancel()

	txts, err := net.DefaultResolver.LookupTXT(ctx, fmt.Sprintf("dkim._domainkey.%s", domain))
	notfound, errStr = IsDNSErrorNoSuchHost(err)
	if notfound || err != nil {
		return
	}

	for _, txt := range txts {
		if regxDKIM.MatchString(txt) {
			records = append(records, txt)

			break
		}
	}

	return
}

func LookupPtr(ip string) (notfound bool, records []string, errStr string) {
	client := new(dns.Client)
	client.Timeout = defaultDNSQueryTimeout
	msg := new(dns.Msg)
	var arpa string
	arpa, err := dns.ReverseAddr(ip)
	if err != nil {
		errStr = err.Error()

		return
	}

	msg.SetQuestion(arpa, dns.TypePTR)
	msg.RecursionDesired = true

	r, _, err := client.Exchange(msg, getRandomDNSServer())
	if err != nil {
		errStr = err.Error()

		return
	}

	for _, a := range r.Answer {
		_a, ok := a.(*dns.PTR)
		if ok {
			records = append(records, strings.TrimSuffix(_a.Ptr, "."))
		}
	}

	notfound = len(records) == 0

	return
}

func LookupDMARC(domain string) (notfound bool, records []string, errStr string) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultDNSQueryTimeout)
	defer cancel()

	txts, err := net.DefaultResolver.LookupTXT(ctx, fmt.Sprintf("_dmarc.%s", domain))
	notfound, errStr = IsDNSErrorNoSuchHost(err)
	if notfound || err != nil {
		return
	}

	for _, txt := range txts {
		if regxDMARC.MatchString(txt) {
			records = append(records, txt)

			break
		}
	}

	return
}

func LookupSRV(domain, dnsTypeStr string) (notfound bool, records []SRVRecord, errStr string) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultDNSQueryTimeout)
	defer cancel()

	_, srvs, err := net.DefaultResolver.LookupSRV(ctx, dnsTypeStr, "tcp", domain)
	notfound, errStr = IsDNSErrorNoSuchHost(err)
	if notfound || err != nil {
		return
	}

	for _, srv := range srvs {
		records = append(records, SRVRecord{
			Priority: srv.Priority,
			Port:     srv.Port,
			Weight:   srv.Weight,
			Target:   strings.TrimSuffix(srv.Target, "."),
		})
	}

	return
}

func LookupRecursiveSPF(domain string, _totalQueries int, dnsType ...uint16) (spf []string, totalQueries int, err error) {
	// FYI http://www.open-spf.org/SPF_Record_Syntax/
	if _totalQueries > 10 {
		return
	}

	if len(dnsType) > 0 {
		switch dnsType[0] {
		case dns.TypeA:
			totalQueries = _totalQueries + 1

			return
		case dns.TypeMX:
			_, mx, _ := LookupMX(domain)
			for _, r := range mx {
				totalQueries = _totalQueries + 1
				_, totalQueries, _ = LookupRecursiveSPF(r.MX, totalQueries, dns.TypeA)
			}

			return
		case dns.TypePTR:
			_, ptr, _ := LookupPtr(domain)
			for _, p := range ptr {
				totalQueries = _totalQueries + 1
				_, totalQueries, _ = LookupRecursiveSPF(p, totalQueries, dns.TypeA)
			}

			return
		}
	}

	_spf, _err := LookupSPF(domain)
	if _totalQueries == 0 {
		spf = _spf
		totalQueries = 1
		err = _err
	} else {
		totalQueries = _totalQueries + 1
	}

	if len(_spf) == 0 {
		return
	}

	mechs := strings.Fields(_spf[0])
	for _, mech := range mechs {
		if strings.HasPrefix(mech, "+") || strings.HasPrefix(mech, "-") ||
			strings.HasPrefix(mech, "~") || strings.HasPrefix(mech, "?") {
			mech = mech[1:]
		}

		if mech == "a" {
			_, totalQueries, _ = LookupRecursiveSPF(domain, totalQueries, dns.TypeA)
		} else if mech == "mx" {
			_, totalQueries, _ = LookupRecursiveSPF(domain, totalQueries, dns.TypeMX)
		} else if mech == "ptr" {
			_, totalQueries, _ = LookupRecursiveSPF(domain, totalQueries, dns.TypePTR)
		} else if strings.HasPrefix(mech, "a:") {
			// a:<domain>
			// a:<domain>/<prefix-length>
			a := strings.TrimPrefix(mech, "a:")
			split := strings.Split(a, "/")
			if len(split) > 1 {
				a = split[0]
			}

			if !emailutils.IsDomain(a) {
				return
			}

			_, totalQueries, _ = LookupRecursiveSPF(a, totalQueries, dns.TypeA)
		} else if strings.HasPrefix(mech, "mx:") {
			// mx:<domain>
			// mx:<domain>/<prefix-length>
			mx := strings.TrimPrefix(mech, "mx:")
			split := strings.Split(mx, "/")
			if len(split) > 1 {
				mx = split[0]
			}

			if !emailutils.IsDomain(mx) {
				return
			}

			_, totalQueries, _ = LookupRecursiveSPF(mx, totalQueries, dns.TypeMX)
		} else if strings.HasPrefix(mech, "ptr:") {
			_, totalQueries, _ = LookupRecursiveSPF(strings.TrimPrefix(mech, "ptr:"), totalQueries, dns.TypePTR)
		} else if strings.HasPrefix(mech, "include:") {
			_, totalQueries, _ = LookupRecursiveSPF(strings.TrimPrefix(mech, "include:"), totalQueries)
		} else if strings.HasPrefix(mech, "redirect=") {
			_, totalQueries, _ = LookupRecursiveSPF(strings.TrimPrefix(mech, "redirect="), totalQueries)
		}
	}

	return
}

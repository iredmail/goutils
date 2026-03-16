package dnsutils

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"net"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/iredmail/goutils/emailutils"
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

// LookupHost 查询域名的 A 和 AAAA 记录，并分别返回 IPv4 和 IPv6 地址列表。
func LookupHost(domain string) (ip4s, ip6s []string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultDNSQueryTimeout)
	defer cancel()

	ips, err := net.DefaultResolver.LookupNetIP(ctx, "ip", domain)
	if err != nil {
		return
	}

	for _, ip := range ips {
		if ip.Is4() {
			ip4s = append(ip4s, ip.String())
		} else if ip.Is6() {
			ip6s = append(ip6s, ip.String())
		}
	}

	return
}

// LookupA 查询域名的 A 记录，并返回 IPv4 地址列表。
func LookupA(domain string) (ip4s []string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultDNSQueryTimeout)
	defer cancel()

	ips, err := net.DefaultResolver.LookupNetIP(ctx, "ip4", domain)
	if err != nil {
		return
	}

	for _, ip := range ips {
		ip4s = append(ip4s, ip.String())
	}

	return
}

// LookupAAAA 查询域名的 AAAA 记录，并返回 IPv6 地址列表。
func LookupAAAA(domain string) (ip6s []string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultDNSQueryTimeout)
	defer cancel()

	ips, err := net.DefaultResolver.LookupNetIP(ctx, "ip6", domain)
	if err != nil {
		return
	}

	for _, ip := range ips {
		ip6s = append(ip6s, ip.String())
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

func LookupDKIM(domain string) (notfound bool, records []string, errStr string) {
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
	ctx, cancel := context.WithTimeout(context.Background(), defaultDNSQueryTimeout)
	defer cancel()

	hosts, err := net.DefaultResolver.LookupAddr(ctx, ip)
	notfound, errStr = IsDNSErrorNoSuchHost(err)
	if err != nil {
		return
	}

	for _, host := range hosts {
		records = append(records, strings.TrimSuffix(host, "."))
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

func LookupRecursiveSPF(domain string, _totalQueries int, dnsType ...uint16) (spf []string, totalQueries int, err error) {
	// FYI http://www.open-spf.org/SPF_Record_Syntax/
	if _totalQueries > 10 {
		return
	}

	if len(dnsType) > 0 {
		switch dnsType[0] {
		case spfDNSQueryTypeA:
			totalQueries = _totalQueries + 1

			return
		case spfDNSQueryTypeMX:
			_, mx, _ := LookupMX(domain)
			for _, r := range mx {
				totalQueries = _totalQueries + 1
				_, totalQueries, _ = LookupRecursiveSPF(r.MX, totalQueries, spfDNSQueryTypeA)
			}

			return
		case spfDNSQueryTypePTR:
			_, ptr, _ := LookupPtr(domain)
			for _, p := range ptr {
				totalQueries = _totalQueries + 1
				_, totalQueries, _ = LookupRecursiveSPF(p, totalQueries, spfDNSQueryTypeA)
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
			_, totalQueries, _ = LookupRecursiveSPF(domain, totalQueries, spfDNSQueryTypeA)
		} else if mech == "mx" {
			_, totalQueries, _ = LookupRecursiveSPF(domain, totalQueries, spfDNSQueryTypeMX)
		} else if mech == "ptr" {
			_, totalQueries, _ = LookupRecursiveSPF(domain, totalQueries, spfDNSQueryTypePTR)
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

			_, totalQueries, _ = LookupRecursiveSPF(a, totalQueries, spfDNSQueryTypeA)
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

			_, totalQueries, _ = LookupRecursiveSPF(mx, totalQueries, spfDNSQueryTypeMX)
		} else if strings.HasPrefix(mech, "ptr:") {
			_, totalQueries, _ = LookupRecursiveSPF(strings.TrimPrefix(mech, "ptr:"), totalQueries, spfDNSQueryTypePTR)
		} else if strings.HasPrefix(mech, "include:") {
			_, totalQueries, _ = LookupRecursiveSPF(strings.TrimPrefix(mech, "include:"), totalQueries)
		} else if strings.HasPrefix(mech, "redirect=") {
			_, totalQueries, _ = LookupRecursiveSPF(strings.TrimPrefix(mech, "redirect="), totalQueries)
		}
	}

	return
}

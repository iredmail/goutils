package dnsutils

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"net"
	"slices"
	"strings"
	"time"
)

var _ Resolver = (*defaultResolver)(nil)

type defaultResolver struct {
	resolver *net.Resolver
	timeout  time.Duration
}

// LookupHost 查询域名的 A 和 AAAA 记录，并分别返回 IPv4 和 IPv6 地址列表。
func (dr *defaultResolver) LookupHost(domain string) (notfound bool, ip4s, ip6s []string, errText string) {
	ctx, cancel := context.WithTimeout(context.Background(), dr.timeout)
	defer cancel()

	ips, err := dr.resolver.LookupNetIP(ctx, "ip", domain)
	notfound, errText = dr.isDNSErrorNoSuchHost(err)
	if notfound || err != nil {
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
func (dr *defaultResolver) LookupA(domain string) (notfound bool, ip4s []string, errText string) {
	ctx, cancel := context.WithTimeout(context.Background(), dr.timeout)
	defer cancel()

	ips, err := dr.resolver.LookupNetIP(ctx, "ip4", domain)
	notfound, errText = dr.isDNSErrorNoSuchHost(err)
	if notfound || err != nil {
		return
	}

	for _, ip := range ips {
		ip4s = append(ip4s, ip.String())
	}

	return
}

// LookupAAAA 查询域名的 AAAA 记录，并返回 IPv6 地址列表。
func (dr *defaultResolver) LookupAAAA(domain string) (notfound bool, ip6s []string, errText string) {
	ctx, cancel := context.WithTimeout(context.Background(), dr.timeout)
	defer cancel()

	ips, err := dr.resolver.LookupNetIP(ctx, "ip6", domain)
	notfound, errText = dr.isDNSErrorNoSuchHost(err)
	if notfound || err != nil {
		return
	}

	for _, ip := range ips {
		ip6s = append(ip6s, ip.String())
	}

	return
}

func (dr *defaultResolver) LookupMX(domain string) (notfound bool, records []MXRecord, errText string) {
	ctx, cancel := context.WithTimeout(context.Background(), dr.timeout)
	defer cancel()

	var mxs []*net.MX
	mxs, err := dr.resolver.LookupMX(ctx, domain)
	notfound, errText = dr.isDNSErrorNoSuchHost(err)
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

	notfound = len(records) == 0

	return
}

func (dr *defaultResolver) LookupDKIM(domain, selector string) (notfound bool, records []string, errText string) {
	ctx, cancel := context.WithTimeout(context.Background(), dr.timeout)
	defer cancel()

	txts, err := dr.resolver.LookupTXT(ctx, fmt.Sprintf("%s._domainkey.%s", selector, domain))
	notfound, errText = dr.isDNSErrorNoSuchHost(err)
	if notfound || err != nil {
		return
	}

	for _, txt := range txts {
		if regxDKIM.MatchString(txt) {
			// net.Resolver 已经返回了拼接后的完整 TXT 字符串，
			// 这里保留完整命中的 DKIM 记录，供调用方直接展示或诊断。
			records = append(records, txt)
		}
	}

	// 域名存在但没有匹配到 DKIM 记录时，也应该视为“未找到目标记录”。
	notfound = len(records) == 0

	return
}

func (dr *defaultResolver) LookupPtr(ip string) (notfound bool, records []string, errText string) {
	ctx, cancel := context.WithTimeout(context.Background(), dr.timeout)
	defer cancel()

	hosts, err := dr.resolver.LookupAddr(ctx, ip)
	notfound, errText = dr.isDNSErrorNoSuchHost(err)
	if err != nil {
		return
	}

	for _, host := range hosts {
		records = append(records, strings.TrimSuffix(host, "."))
	}

	notfound = len(records) == 0

	return
}

func (dr *defaultResolver) LookupDMARC(domain string) (notfound bool, records []string, errText string) {
	ctx, cancel := context.WithTimeout(context.Background(), dr.timeout)
	defer cancel()

	txts, err := dr.resolver.LookupTXT(ctx, fmt.Sprintf("_dmarc.%s", domain))
	notfound, errText = dr.isDNSErrorNoSuchHost(err)
	if notfound || err != nil {
		return
	}

	for _, txt := range txts {
		if regxDMARC.MatchString(txt) {
			// LookupTXT 返回的是完整 TXT 记录；这里直接保存完整命中值，
			// 避免调用方只能看到被截断的 DMARC 内容。
			records = append(records, txt)
		}
	}

	// 查询成功但没有匹配到 DMARC 记录时，应明确返回 notfound=true。
	notfound = len(records) == 0

	return
}

func (dr *defaultResolver) LookupSRV(domain, dnsTypeStr string) (notfound bool, records []SRVRecord, errText string) {
	ctx, cancel := context.WithTimeout(context.Background(), dr.timeout)
	defer cancel()

	_, srvs, err := dr.resolver.LookupSRV(ctx, dnsTypeStr, "tcp", domain)
	notfound, errText = dr.isDNSErrorNoSuchHost(err)
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

	notfound = len(records) == 0

	return
}

func (dr *defaultResolver) LookupSPF(domain string) (notfound bool, records []string, errText string) {
	ctx, cancel := context.WithTimeout(context.Background(), dr.timeout)
	defer cancel()

	var txts []string
	txts, err := dr.resolver.LookupTXT(ctx, domain)
	notfound, errText = dr.isDNSErrorNoSuchHost(err)
	if notfound || err != nil {
		return
	}

	for _, txt := range txts {
		if regxSPF.MatchString(txt) {
			// SPF 的后续递归解析依赖完整原文，因此这里保留完整命中的记录，
			// 而不是在首条命中后提前退出。
			records = append(records, txt)
		}
	}

	// 域名存在但未发布 SPF 记录时，返回 notfound=true 更符合语义。
	notfound = len(records) == 0

	return
}

func (dr *defaultResolver) LookupRecursiveSPF(domain string, _totalQueries int, dnsType ...uint16) (notfound bool, spf []string, totalQueries int, errText string) {
	return lookupRecursiveSPF(dr, domain, _totalQueries, dnsType...)
}

func (dr *defaultResolver) isDNSErrorNoSuchHost(err error) (v bool, e string) {
	if err == nil {
		return false, ""
	}

	e = err.Error()

	if _err, ok := errors.AsType[*net.DNSError](err); ok {
		v = _err.Err == "no such host"
		if v {
			e = ""
		}
	}

	return
}

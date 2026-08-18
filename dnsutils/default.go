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

	"github.com/iredmail/goutils/emailutils"
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
	// FYI http://www.open-spf.org/SPF_Record_Syntax/
	// RFC 7208 要求会触发 DNS 查询的 SPF 机制/修饰符总数最多为 10。
	if _totalQueries >= 10 {
		totalQueries = _totalQueries

		return
	}

	if len(dnsType) > 0 {
		switch dnsType[0] {
		case spfDNSQueryTypeA:
			totalQueries = _totalQueries + 1

			return
		case spfDNSQueryTypeMX:
			// mx 机制本身就会触发一次 MX 查询，这一步即使拿不到任何 MX 主机，
			// 也应该计入 SPF 的总查询次数。
			totalQueries = _totalQueries + 1

			_, mx, _ := dr.LookupMX(domain)
			for _, _r := range mx {
				if totalQueries >= 10 {
					return
				}

				_, _, totalQueries, _ = dr.LookupRecursiveSPF(_r.MX, totalQueries, spfDNSQueryTypeA)
			}

			return
		case spfDNSQueryTypePTR:
			// ptr 机制至少会进行一次 PTR 查询；先把这一步计入总数。
			// 完整 PTR 语义仍依赖连接 IP，当前 API 只能做近似统计。
			totalQueries = _totalQueries + 1

			_, ptr, _ := dr.LookupPtr(domain)
			for _, p := range ptr {
				if totalQueries >= 10 {
					return
				}

				_, _, totalQueries, _ = dr.LookupRecursiveSPF(p, totalQueries, spfDNSQueryTypeA)
			}

			return
		}
	}

	var _spf []string
	notfound, _spf, errText = dr.LookupSPF(domain)
	if notfound {
		return
	}

	if _totalQueries == 0 {
		spf = _spf
		totalQueries = 1
	} else {
		totalQueries = _totalQueries + 1
	}

	if len(_spf) == 0 {
		return
	}

	var after string
	var ok bool
	for mech := range strings.FieldsSeq(_spf[0]) {
		if strings.HasPrefix(mech, "+") || strings.HasPrefix(mech, "-") ||
			strings.HasPrefix(mech, "~") || strings.HasPrefix(mech, "?") {
			mech = mech[1:]
		}

		if mech == "a" {
			_, _, totalQueries, _ = dr.LookupRecursiveSPF(domain, totalQueries, spfDNSQueryTypeA)
		} else if mech == "mx" {
			_, _, totalQueries, _ = dr.LookupRecursiveSPF(domain, totalQueries, spfDNSQueryTypeMX)
		} else if mech == "ptr" {
			_, _, totalQueries, _ = dr.LookupRecursiveSPF(domain, totalQueries, spfDNSQueryTypePTR)
		} else if after, ok = strings.CutPrefix(mech, "a:"); ok {
			// a:<domain>
			// a:<domain>/<prefix-length>
			a := after
			split := strings.Split(a, "/")
			if len(split) > 1 {
				a = split[0]
			}

			if !emailutils.IsDomain(a) {
				return
			}

			_, _, totalQueries, _ = dr.LookupRecursiveSPF(a, totalQueries, spfDNSQueryTypeA)
		} else if after, ok = strings.CutPrefix(mech, "mx:"); ok {
			// mx:<domain>
			// mx:<domain>/<prefix-length>
			mx := after
			split := strings.Split(mx, "/")
			if len(split) > 1 {
				mx = split[0]
			}

			if !emailutils.IsDomain(mx) {
				return
			}

			_, _, totalQueries, _ = dr.LookupRecursiveSPF(mx, totalQueries, spfDNSQueryTypeMX)
		} else if after, ok = strings.CutPrefix(mech, "ptr:"); ok {
			_, _, totalQueries, _ = dr.LookupRecursiveSPF(after, totalQueries, spfDNSQueryTypePTR)
		} else if after, ok = strings.CutPrefix(mech, "include:"); ok {
			_, _, totalQueries, _ = dr.LookupRecursiveSPF(after, totalQueries)
		} else if after, ok = strings.CutPrefix(mech, "redirect="); ok {
			_, _, totalQueries, _ = dr.LookupRecursiveSPF(after, totalQueries)
		}
	}

	return
}

func (dr *defaultResolver) isDNSErrorNoSuchHost(err error) (v bool, e string) {
	if err == nil {
		return false, ""
	}

	e = err.Error()

	if _err, ok := errors.AsType[*net.DNSError](err); ok {
		v = _err.Err == "no such host"
	}

	return
}

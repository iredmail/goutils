package dnsutils

import (
	"net"
	"regexp"
	"strings"
	"time"

	"github.com/miekg/dns"

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
	// exists 不是实际 RR type，只用于 LookupRecursiveSPF 内部标记“会触发一次 DNS 查询”。
	spfDNSQueryTypeExists uint16 = 0
)

func NewResolver(server string, timeout int) Resolver {
	timeoutDuration := defaultDNSQueryTimeout

	if timeout > 0 {
		timeoutDuration = time.Duration(timeout) * time.Second
	}

	if server != "" {
		return &customResolver{
			client:  &dns.Client{Timeout: timeoutDuration},
			dnsAddr: server,
		}
	}

	return &defaultResolver{
		resolver: net.DefaultResolver,
		timeout:  timeoutDuration,
	}
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

func lookupRecursiveSPF(r Resolver, domain string, _totalQueries int, dnsType ...uint16) (notfound bool, spf []string, totalQueries int, errText string) {
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
		case spfDNSQueryTypeExists:
			// exists 机制会额外触发一次 DNS 查询；这里只做计数，不继续做完整 SPF 求值。
			totalQueries = _totalQueries + 1

			return
		case spfDNSQueryTypeMX:
			// mx 机制本身会产生一次 MX 查询；即使后面没有任何 MX 主机，
			// 这次查询也应该计入总次数。
			totalQueries = _totalQueries + 1

			_, mx, _ := r.LookupMX(domain)
			for _, _r := range mx {
				if totalQueries >= 10 {
					return
				}

				_, _, totalQueries, _ = lookupRecursiveSPF(r, _r.MX, totalQueries, spfDNSQueryTypeA)
			}

			return
		case spfDNSQueryTypePTR:
			// ptr 机制至少会触发一次 PTR 查询；这里先把这一步记入计数。
			// 注意：完整的 PTR SPF 语义仍然依赖连接 IP，当前 API 只能做近似统计。
			totalQueries = _totalQueries + 1

			/*
				_, ptr, _ := r.LookupPtr(domain)
				for _, p := range ptr {
					if totalQueries >= 10 {
						return
					}

					_, _, totalQueries, _ = lookupRecursiveSPF(r, p, totalQueries, spfDNSQueryTypeA)
				}
			*/

			return
		}
	}

	var _spf []string
	notfound, _spf, errText = r.LookupSPF(domain)
	if _totalQueries == 0 {
		spf = _spf
		totalQueries = 1
	} else {
		totalQueries = _totalQueries + 1
	}

	if notfound || len(_spf) == 0 {
		return
	}

	var after string
	var ok bool
	for mech := range strings.FieldsSeq(_spf[0]) {
		mech = strings.ToLower(mech)

		if strings.HasPrefix(mech, "+") || strings.HasPrefix(mech, "-") ||
			strings.HasPrefix(mech, "~") || strings.HasPrefix(mech, "?") {
			mech = mech[1:]
		}

		if mech == "a" {
			_, _, totalQueries, _ = lookupRecursiveSPF(r, domain, totalQueries, spfDNSQueryTypeA)
		} else if mech == "mx" {
			_, _, totalQueries, _ = lookupRecursiveSPF(r, domain, totalQueries, spfDNSQueryTypeMX)
		} else if mech == "ptr" {
			_, _, totalQueries, _ = lookupRecursiveSPF(r, domain, totalQueries, spfDNSQueryTypePTR)
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

			_, _, totalQueries, _ = lookupRecursiveSPF(r, a, totalQueries, spfDNSQueryTypeA)
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

			_, _, totalQueries, _ = lookupRecursiveSPF(r, mx, totalQueries, spfDNSQueryTypeMX)
		} else if after, ok = strings.CutPrefix(mech, "ptr:"); ok {
			_, _, totalQueries, _ = lookupRecursiveSPF(r, after, totalQueries, spfDNSQueryTypePTR)
		} else if after, ok = strings.CutPrefix(mech, "exists:"); ok {
			// exists:<domain-spec> 是 RFC 7208 里的 SPF 机制之一：
			// 1) 先对 domain-spec 做 macro 展开；
			// 2) 再查询展开后的域名是否“存在”可解析的 A 记录；
			// 3) 这个动作本身会消耗一次 DNS 查询配额，必须计入 10 次上限。
			//
			// 但当前函数只是“递归查询次数估算器”，并没有 client IP / macro 上下文，
			// 因此这里只做计数，不尝试做完整的 exists 匹配求值。
			_, _, totalQueries, _ = lookupRecursiveSPF(r, after, totalQueries, spfDNSQueryTypeExists)
		} else if after, ok = strings.CutPrefix(mech, "include:"); ok {
			_, _, totalQueries, _ = lookupRecursiveSPF(r, after, totalQueries)
		} else if after, ok = strings.CutPrefix(mech, "redirect="); ok {
			_, _, totalQueries, _ = lookupRecursiveSPF(r, after, totalQueries)
		}
	}

	return
}

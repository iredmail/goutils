package dnsutils

import (
	"cmp"
	"fmt"
	"slices"
	"strings"

	"github.com/miekg/dns"

	"github.com/iredmail/goutils/emailutils"
)

var _ Resolver = (*customResolver)(nil)

type customResolver struct {
	client  *dns.Client
	dnsAddr string
}

// LookupHost 查询域名的 A 和 AAAA 记录，并分别返回 IPv4 和 IPv6 地址列表。
func (cr *customResolver) LookupHost(domain string) (notfound bool, ip4s, ip6s []string, errText string) {
	var errTextIP4, errTextIP46 string

	notfound, ip4s, errTextIP4 = cr.LookupA(domain)

	var _notfound bool
	_notfound, ip6s, errTextIP46 = cr.LookupAAAA(domain)

	notfound = notfound && _notfound

	var errs []string
	if errTextIP4 != "" {
		errs = append(errs, errTextIP4)
	}

	if errTextIP46 != "" {
		errs = append(errs, errTextIP46)
	}

	errText = strings.Join(errs, "; ")

	return
}

// LookupA 查询域名的 A 记录，并返回 IPv4 地址列表。
func (cr *customResolver) LookupA(domain string) (notfound bool, ip4s []string, errText string) {
	var answers []dns.RR
	notfound, answers, errText = cr.exchange(domain, dns.TypeA)
	if notfound || errText != "" {
		return
	}

	for _, a := range answers {
		_a, ok := a.(*dns.A)
		if ok {
			ip4s = append(ip4s, _a.A.String())
		}
	}

	return
}

// LookupAAAA 查询域名的 AAAA 记录，并返回 IPv6 地址列表。
func (cr *customResolver) LookupAAAA(domain string) (notfound bool, ip6s []string, errText string) {
	var answers []dns.RR
	notfound, answers, errText = cr.exchange(domain, dns.TypeAAAA)
	if notfound || errText != "" {
		return
	}

	for _, a := range answers {
		_a, ok := a.(*dns.AAAA)
		if ok {
			ip6s = append(ip6s, _a.AAAA.String())
		}
	}

	return
}

func (cr *customResolver) LookupMX(domain string) (notfound bool, records []MXRecord, errText string) {
	var answers []dns.RR
	notfound, answers, errText = cr.exchange(domain, dns.TypeMX)
	if notfound || errText != "" {
		return
	}

	for _, a := range answers {
		mx, ok := a.(*dns.MX)
		if ok {
			records = append(records, MXRecord{
				MX:       strings.TrimSuffix(mx.Mx, "."),
				Priority: mx.Preference,
			})
		}
	}

	// Sort by mx priority
	slices.SortFunc(records, func(a, b MXRecord) int {
		return cmp.Compare(a.Priority, b.Priority)
	})

	return
}

func (cr *customResolver) LookupDKIM(domain, selector string) (notfound bool, records []string, errText string) {
	var answers []dns.RR
	notfound, answers, errText = cr.exchange(fmt.Sprintf("%s._domainkey.%s", selector, domain), dns.TypeTXT)
	if notfound || errText != "" {
		return
	}

	for _, a := range answers {
		txts, ok := a.(*dns.TXT)
		if !ok {
			continue
		}

		// 一条 TXT RR 在 DNS 报文里可能被拆成多个字符串片段，
		// 必须先拼回完整内容再匹配并返回，否则长 DKIM 记录会被截断。
		txt := strings.Join(txts.Txt, "")
		if regxDKIM.MatchString(txt) {
			// 返回完整命中的记录，供调用方做后续展示或诊断。
			records = append(records, txt)
		}
	}

	// 域名存在但没有匹配到 DKIM 记录时，也应该视为“未找到目标记录”。
	notfound = len(records) == 0

	return
}

func (cr *customResolver) LookupPtr(ip string) (notfound bool, records []string, errText string) {
	arpa, err := dns.ReverseAddr(ip)
	if err != nil {
		errText = err.Error()

		return
	}

	var answers []dns.RR
	notfound, answers, errText = cr.exchange(arpa, dns.TypePTR)
	if notfound || errText != "" {
		return
	}

	for _, a := range answers {
		ptr, ok := a.(*dns.PTR)
		if ok {
			records = append(records, strings.TrimSuffix(ptr.Ptr, "."))
		}
	}

	return
}

func (cr *customResolver) LookupDMARC(domain string) (notfound bool, records []string, errText string) {
	var answers []dns.RR
	notfound, answers, errText = cr.exchange(fmt.Sprintf("_dmarc.%s", domain), dns.TypeTXT)
	if notfound || errText != "" {
		return
	}

	for _, a := range answers {
		txts, ok := a.(*dns.TXT)
		if !ok {
			continue
		}

		// DMARC TXT 也可能被拆分成多个片段，先拼接后再做正则匹配，
		// 并把完整记录返回给调用方，避免只拿到前半段内容。
		txt := strings.Join(txts.Txt, "")
		if regxDMARC.MatchString(txt) {
			// 保留完整命中的 DMARC 记录，便于上层直接展示和排障。
			records = append(records, txt)
		}
	}

	// 查询成功但没有匹配到 DMARC 记录时，应明确返回 notfound=true。
	notfound = len(records) == 0

	return
}

func (cr *customResolver) LookupSRV(domain, dnsTypeStr string) (notfound bool, records []SRVRecord, errText string) {
	// 例如: _sip._tcp.example.com.
	target := fmt.Sprintf("_%s._tcp.%s", dnsTypeStr, domain)

	var answers []dns.RR
	notfound, answers, errText = cr.exchange(target, dns.TypeSRV)
	if notfound || errText != "" {
		return
	}

	for _, a := range answers {
		srv, ok := a.(*dns.SRV)
		if !ok {
			continue
		}

		records = append(records, SRVRecord{
			Priority: srv.Priority,
			Port:     srv.Port,
			Weight:   srv.Weight,
			Target:   strings.TrimSuffix(srv.Target, "."),
		})
	}

	return
}

func (cr *customResolver) LookupSPF(domain string) (notfound bool, records []string, errText string) {
	var answers []dns.RR
	notfound, answers, errText = cr.exchange(domain, dns.TypeTXT)
	if notfound || errText != "" {
		return
	}

	for _, a := range answers {
		txts, ok := a.(*dns.TXT)
		if !ok {
			continue
		}

		// SPF 记录经常因为过长被拆成多个 TXT 片段，必须先拼接成完整字符串，
		// 否则只 append 某一个片段会导致 include/redirect 等机制被截断。
		txt := strings.Join(txts.Txt, "")
		if regxSPF.MatchString(txt) {
			// 保留完整命中的 SPF 记录，后续递归解析依赖完整原文。
			records = append(records, txt)
		}
	}

	// 域名存在但未发布 SPF 记录时，返回 notfound=true 更符合调用方预期。
	notfound = len(records) == 0

	return
}

func (cr *customResolver) LookupRecursiveSPF(domain string, _totalQueries int, dnsType ...uint16) (notfound bool, spf []string, totalQueries int, errText string) {
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

			_, mx, _ := cr.LookupMX(domain)
			for _, _r := range mx {
				if totalQueries >= 10 {
					return
				}

				_, _, totalQueries, _ = cr.LookupRecursiveSPF(_r.MX, totalQueries, spfDNSQueryTypeA)
			}

			return
		case spfDNSQueryTypePTR:
			// ptr 机制至少会触发一次 PTR 查询；这里先把这一步记入计数。
			// 注意：完整的 PTR SPF 语义仍然依赖连接 IP，当前 API 只能做近似统计。
			totalQueries = _totalQueries + 1

			_, ptr, _ := cr.LookupPtr(domain)
			for _, p := range ptr {
				if totalQueries >= 10 {
					return
				}

				_, _, totalQueries, _ = cr.LookupRecursiveSPF(p, totalQueries, spfDNSQueryTypeA)
			}

			return
		}
	}

	var _spf []string
	notfound, _spf, errText = cr.LookupSPF(domain)
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
			_, _, totalQueries, _ = cr.LookupRecursiveSPF(domain, totalQueries, spfDNSQueryTypeA)
		} else if mech == "mx" {
			_, _, totalQueries, _ = cr.LookupRecursiveSPF(domain, totalQueries, spfDNSQueryTypeMX)
		} else if mech == "ptr" {
			_, _, totalQueries, _ = cr.LookupRecursiveSPF(domain, totalQueries, spfDNSQueryTypePTR)
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

			_, _, totalQueries, _ = cr.LookupRecursiveSPF(a, totalQueries, spfDNSQueryTypeA)
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

			_, _, totalQueries, _ = cr.LookupRecursiveSPF(mx, totalQueries, spfDNSQueryTypeMX)
		} else if after, ok = strings.CutPrefix(mech, "ptr:"); ok {
			_, _, totalQueries, _ = cr.LookupRecursiveSPF(after, totalQueries, spfDNSQueryTypePTR)
		} else if after, ok = strings.CutPrefix(mech, "exists:"); ok {
			// exists:<domain-spec> 是 RFC 7208 里的 SPF 机制之一：
			// 1) 先对 domain-spec 做 macro 展开；
			// 2) 再查询展开后的域名是否“存在”可解析的 A 记录；
			// 3) 这个动作本身会消耗一次 DNS 查询配额，必须计入 10 次上限。
			//
			// 但当前函数只是“递归查询次数估算器”，并没有 client IP / macro 上下文，
			// 因此这里只做计数，不尝试做完整的 exists 匹配求值。
			_, _, totalQueries, _ = cr.LookupRecursiveSPF(after, totalQueries, spfDNSQueryTypeExists)
		} else if after, ok = strings.CutPrefix(mech, "include:"); ok {
			_, _, totalQueries, _ = cr.LookupRecursiveSPF(after, totalQueries)
		} else if after, ok = strings.CutPrefix(mech, "redirect="); ok {
			_, _, totalQueries, _ = cr.LookupRecursiveSPF(after, totalQueries)
		}
	}

	return
}

func (cr *customResolver) exchange(domain string, dnsType uint16) (notfound bool, answers []dns.RR, errText string) {
	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(domain), dnsType)
	// 显式带上 EDNS0，尽量减少长 TXT 记录因 UDP 报文过小而被截断的概率。
	msg.SetEdns0(1232, false)

	// 复制一份 client，避免在 TCP 回退时直接修改共享的 resolver 配置。
	client := *cr.client

	r, _, err := client.Exchange(msg, cr.dnsAddr)
	if err != nil {
		errText = err.Error()

		return
	}

	if r == nil {
		errText = "empty dns response"

		return
	}

	if r.Truncated && client.Net != "tcp" {
		// UDP 响应被截断时，按 DNS 常见处理方式回退到 TCP 重试，
		// 这样可以拿到完整的 TXT/SPF/DMARC/DKIM 记录。
		tcpClient := client
		tcpClient.Net = "tcp"

		r, _, err = tcpClient.Exchange(msg, cr.dnsAddr)
		if err != nil {
			errText = err.Error()

			return
		}

		if r == nil {
			errText = "empty dns response"

			return
		}
	}

	switch r.Rcode {
	case dns.RcodeSuccess:
		answers = r.Answer
	case dns.RcodeNameError:
		// 只有 NXDOMAIN 才表示名字不存在，统一映射为 notfound。
		notfound = true
	default:
		// 其他 RCODE（如 SERVFAIL/REFUSED）不能静默吞掉，
		// 否则上层会误以为“查询成功但没有记录”。
		rcode := dns.RcodeToString[r.Rcode]
		if rcode == "" {
			rcode = fmt.Sprintf("RCODE(%d)", r.Rcode)
		}

		errText = fmt.Sprintf("dns query failed with rcode %s", rcode)
	}

	if notfound || errText != "" {
		return
	}

	return
}

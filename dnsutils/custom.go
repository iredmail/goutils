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

		for _, txt := range txts.Txt {
			if regxDKIM.MatchString(txt) {
				records = append(records, txt)

				break
			}
		}
	}

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

		for _, txt := range txts.Txt {
			if regxDMARC.MatchString(txt) {
				records = append(records, txt)

				break
			}
		}
	}

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

		for _, txt := range txts.Txt {
			if regxSPF.MatchString(txt) {
				records = append(records, txt)

				break
			}
		}
	}

	return
}

func (cr *customResolver) LookupRecursiveSPF(domain string, _totalQueries int, dnsType ...uint16) (notfound bool, spf []string, totalQueries int, errText string) {
	// FYI http://www.open-spf.org/SPF_Record_Syntax/
	if _totalQueries > 10 {
		totalQueries = _totalQueries

		return
	}

	if len(dnsType) > 0 {
		switch dnsType[0] {
		case spfDNSQueryTypeA:
			totalQueries = _totalQueries + 1

			return
		case spfDNSQueryTypeMX:
			_, mx, _ := cr.LookupMX(domain)
			for _, _r := range mx {
				totalQueries = _totalQueries + 1
				_, _, totalQueries, _ = cr.LookupRecursiveSPF(_r.MX, totalQueries, spfDNSQueryTypeA)
			}

			return
		case spfDNSQueryTypePTR:
			_, ptr, _ := cr.LookupPtr(domain)
			for _, p := range ptr {
				totalQueries = _totalQueries + 1
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

	r, _, err := cr.client.Exchange(msg, cr.dnsAddr)
	if err != nil {
		errText = err.Error()

		return
	}

	notfound = r.Rcode == dns.RcodeNameError
	if notfound {
		return
	}

	answers = r.Answer

	return
}

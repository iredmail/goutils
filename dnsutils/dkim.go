package dnsutils

import (
	"errors"

	"github.com/miekg/dns"
)

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

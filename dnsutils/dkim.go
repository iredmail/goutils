package dnsutils

import (
	"errors"

	"github.com/miekg/dns"
)

// TODO QueryDKIM
// QueryDKIM 查询域名的 DKIM 记录。格式为：`<selector>._domainkey.<domain>`。
func QueryDKIM(domain, selector string) (found bool, result ResultDKIM) {
	if selector == "" {
		result.Error = errors.New("selector is missing")

		return
	}

	found, answers, duration, err := queryTXT(domain)
	if err != nil {
		result.Error = err

		return
	}

	if !found {
		return
	}

	result.Domain = domain
	result.Duration = duration

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

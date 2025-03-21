package dnsutils

import (
	"strings"
	"time"

	"github.com/miekg/dns"
)

type ResultDMARC struct {
	Domain   string
	Txt      string
	Duration time.Duration
	TTL      uint32
	Params   map[string]string
}

func QueryDMARC(domain string) (found bool, result ResultDMARC, err error) {
	foundTxt, answers, duration, err := queryTXT("_dmarc." + domain)
	if err != nil || !foundTxt {
		return
	}

	result.Domain = domain
	result.Duration = duration

	for _, ans := range answers {
		if txt, ok := ans.(*dns.TXT); ok {
			for _, txtStr := range txt.Txt {
				if regxDMARC.MatchString(txtStr) {
					found = true
					result.Txt = txtStr
					result.TTL = txt.Hdr.Ttl
					result.Params = parseDMARCParams(txtStr)

					return
				}
			}
		}
	}

	return
}

func parseDMARCParams(txt string) map[string]string {
	params := make(map[string]string)
	parts := strings.Split(txt, ";")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if k, v, found := strings.Cut(part, "="); found {
			params[strings.ToLower(k)] = strings.ToLower(v)
		}
	}

	return params
}

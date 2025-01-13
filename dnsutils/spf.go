package dnsutils

import (
	"strings"
	"time"

	"github.com/miekg/dns"
)

// SPFMech 定义原始 SPF 记录中的一个 mech 及其对应的 prefixed qualifier。
type SPFMech struct {
	Qualifier string
	Mech      string
	Value     string
}

// ResultSPF 定义 SPF 记录的查询结果。
type ResultSPF struct {
	Domain string
	RTT    time.Duration
	TTL    uint32
	Txt    string // 原始的 SPF 记录

	Mechs []SPFMech

	IP4     []string // `ip4:`
	IP6     []string // `ip6:`
	A       []string // `a:`
	MX      []string // `mx:`
	Include []string // `include:`
	Exists  []string // `exists:`

	// 经过完整解析后得到的所有 IP 地址及其对应的 action。
	// 例如：
	// {
	// 	"172.105.68.48": "+",
	// 	"2a01:7e01::f03c:93ff:fe25:7e10": "+",
	// 	"2a01:7e01::f03c:91ff:fe74:9543": "+",
	// 	"172.104.245.227": "-",
	// }
	IPActions map[string]string
}

// QuerySPF 查询 SPF 记录。
// FYI http://www.open-spf.org/SPF_Record_Syntax/
func QuerySPF(domain string) (found bool, result ResultSPF, err error) {
	found, answers, rtt, err := queryTXT(domain)
	if err != nil || !found {
		return
	}

	result.Domain = domain
	result.RTT = rtt

	// 一个域名一般只有一个 SPF 记录，这里只取第一个。
	for _, ans := range answers {
		if txt, ok := ans.(*dns.TXT); ok {
			for _, txtStr := range txt.Txt {
				if regxSPF.MatchString(txtStr) {
					result.Txt = txtStr
					result.TTL = txt.Hdr.Ttl

					break
				}
			}
		}
	}

	//
	// 处理 SPF 记录的各个 mechanisms
	//
	// Samples:
	// "v=spf1 mx:iredmail.org ip4:172.105.68.48 ip6:2a01:7e01::f03c:93ff:fe25:7e10 ip4:172.104.245.227 -all"

	// 记录总的查询次数。
	// var totalQueries int

	var mqs []SPFMech

	// 将 SPF 记录 split 后根据 tag 的不同做分类
	mechs := strings.Fields(result.Txt)
	for _, mech := range mechs {
		mq := SPFMech{}

		if strings.HasPrefix(mech, "+") || strings.HasPrefix(mech, "-") || strings.HasPrefix(mech, "~") || strings.HasPrefix(mech, "?") {
			mq.Qualifier = mech[0:1]
			mech = mech[1:]
		} else {
			mq.Qualifier = "+" // defaults to "+" if missing.
		}

		mechInLower := strings.ToLower(mech)

		if mechInLower == "a" {
			mq.Mech = "a"
			mq.Value = domain
		} else if mechInLower == "mx" {
			mq.Mech = "mx"
			mq.Value = domain
		} else if mechInLower == "ptr" {
			// TODO handle `ptr`
			continue
		} else if mechInLower == "all" {
			mq.Mech = "all"
		} else {
			// handle `redirect=<domain>`
			if strings.HasPrefix(mechInLower, "redirect=") {
				_, value, _ := strings.Cut(mechInLower, "=")
				mq.Mech = "redirect"
				mq.Value = value
			} else {
				// handle `mech:value`
				mechName, _, found := strings.Cut(mechInLower, ":")
				if !found {
					continue
				}

				switch mechName {
				case "ip4", "a", "mx", "include", "exists":
					_, value, _ := strings.Cut(mechInLower, ":")
					mq.Mech = mechName
					mq.Value = value
				case "ip6":
					_, value, _ := strings.Cut(mech, ":") // 使用未经转换为小写的值
					mq.Mech = mechName
					mq.Value = value
				}
			}
		}

		mqs = append(mqs, mq)
	}

	// TODO 解析完了所有的 mechanisms，启动并发查询。

	return
}

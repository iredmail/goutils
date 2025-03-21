package dnsutils

import (
	"slices"
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
	Domain       string
	Txt          string        // 原始的 SPF 记录
	Duration     time.Duration // 总耗时
	TTL          uint32        // Time-To-Live
	TotalQueries int           // 总共执行了多少次 DNS 查询

	mechs []SPFMech

	// 已经查询过的 mech，避免重复查询。
	finished []SPFMech

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

func newResultSPF() ResultSPF {
	return ResultSPF{
		IPActions: make(map[string]string),
	}
}

// QuerySPF 查询 SPF 记录。
// FYI http://www.open-spf.org/SPF_Record_Syntax/
func QuerySPF(domain string) (foundRecord bool, result ResultSPF, err error) {
	foundTxt, answers, duration, err := queryTXT(domain)
	if err != nil || !foundTxt {
		return
	}

	result = newResultSPF()
	result.Domain = domain
	result.Duration = duration

	// 一个域名一般只有一个 SPF 记录，这里只取第一个。
	for _, ans := range answers {
		if foundRecord {
			break
		}

		if txt, ok := ans.(*dns.TXT); ok {
			for _, txtStr := range txt.Txt {
				if regxSPF.MatchString(txtStr) {
					foundRecord = true

					result.Txt = txtStr
					result.TTL = txt.Hdr.Ttl

					break
				}
			}
		}
	}

	if !foundRecord {
		return
	}

	//
	// 处理 SPF 记录的各个 mechanisms
	//
	// Samples:
	// "v=spf1 mx:iredmail.org ip4:172.105.68.48 ip6:2a01:7e01::f03c:93ff:fe25:7e10 ip4:172.104.245.227 -all"
	// gmail.com: "v=spf1 redirect=_spf.google.com"

	// 记录总的查询次数。
	result.TotalQueries = 1

	// 将 SPF 记录 split 后根据 tag 的不同做分类
	mechs := strings.Fields(result.Txt)
	for _, mech := range mechs {
		mq := SPFMech{}

		if strings.HasPrefix(mech, "+") || strings.HasPrefix(mech, "-") || strings.HasPrefix(mech, "~") || strings.HasPrefix(mech, "?") {
			mq.Qualifier = mech[0:1]
			mech = mech[1:]
		}

		mechInLower := strings.ToLower(mech)

		if mechInLower == "a" {
			mq.Mech = "a"
			mq.Value = domain
		} else if mechInLower == "mx" {
			mq.Mech = "mx"
			mq.Value = domain
			// result.mxs = append(result.mxs, domain)
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
				case "ip4", "a", "mx", "include", "?include", "exists":
					// FIXME 实际处理 `exists`
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

		result.mechs = append(result.mechs, mq)
	}

	// FIXME 并发查询
	for _, mq := range result.mechs {
		if slices.Contains(result.finished, mq) {
			continue
		}

		result.finished = append(result.finished, mq)

		switch mq.Mech {
		case "ip4", "ip6":
			result.IPActions[mq.Value] = mq.Qualifier
		case "a":
			result.TotalQueries++
			result.finished = append(result.finished, mq)

			found, ra, err := QueryA(mq.Value)
			result.Duration += ra.RTT

			if err != nil || !found {
				continue
			}

			for _, ip := range ra.IPs {
				result.IPActions[ip] = mq.Qualifier
			}
		case "mx":
			result.TotalQueries++

			found, rmx, err := QueryMX(mq.Value)
			result.Duration += rmx.RTT
			if err != nil || !found {
				continue
			}

			// 递归解析
			for _, hostMX := range rmx.Hosts {
				result.TotalQueries++

				found, ra, err := QueryA(hostMX.Hostname)
				result.Duration += ra.RTT

				if err != nil || !found {
					continue
				}

				for _, ip := range ra.IPs {
					result.IPActions[ip] = mq.Qualifier
				}
			}
		case "include", "?include", "redirect":
			result.TotalQueries++

			found, rspf, err := QuerySPF(mq.Value)
			result.Duration += rspf.Duration
			if err != nil || !found {
				continue
			}

			result.TotalQueries += rspf.TotalQueries
			for k, v := range rspf.IPActions {
				result.IPActions[k] = v
			}
		}
	}

	return
}

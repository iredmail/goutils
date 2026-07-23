package dnsutils

import (
	"net"
	"strconv"
	"strings"
)

const (
	// defaultSPFMaxQueryDepth 是 SPF 记录查询的最大递归深度，超过此值将停止递归查询。
	// 10 是 SPF 规范中推荐的最大查询次数，超过此值可能会导致 DNS 查询失败或被拒绝。
	defaultSPFMaxQueryDepth = 10
)

// GetDomainSPFNetworks Retrieves the SPF networks for a given domain.
//
//	Default maxDepth is 10, but can be overridden by passing a positive integer as the second argument.
func GetDomainSPFNetworks(domain string, maxDepth ...int) (networks []string, err error) {
	_maxDepth := defaultSPFMaxQueryDepth
	if len(maxDepth) > 0 && maxDepth[0] > 0 {
		_maxDepth = maxDepth[0]
	}

	return recursiveGetDomainSPFNetworks(domain, _maxDepth, 0)
}

func recursiveGetDomainSPFNetworks(domain string, maxDepth, curDepth int) (networks []string, err error) {
	if curDepth > maxDepth {
		return
	}

	records, err := LookupSPF(domain)
	if err != nil || len(records) == 0 {
		return
	}

	for mech := range strings.FieldsSeq(records[0]) {
		if strings.EqualFold(mech, "v=spf1") {
			continue
		}

		// Skip mechanisms with SPF qualifiers that indicate non-allowance.
		if strings.HasPrefix(mech, "-") ||
			strings.HasPrefix(mech, "~") {
			continue
		}

		if strings.HasPrefix(mech, "+") ||
			strings.HasPrefix(mech, "?") {
			mech = mech[1:]

			if mech == "" {
				continue
			}
		}

		_networks, err := getSPFMechanismNetworks(mech, domain, maxDepth, curDepth)
		if err != nil {
			return nil, err
		}

		networks = append(networks, _networks...)
	}

	return
}

// IsAllowedIPInSPF Checks if a given IP address is allowed by the SPF records of a specified domain.
//
//	Default maxDepth is 10, but can be overridden by passing a positive integer as the third argument.
func IsAllowedIPInSPF(domain string, ip net.IP, maxDepth ...int) (matched bool, err error) {
	_maxDepth := defaultSPFMaxQueryDepth
	if len(maxDepth) > 0 && maxDepth[0] > 0 {
		_maxDepth = maxDepth[0]
	}

	return recursiveIsAllowedInSPF(domain, ip, _maxDepth, 0)
}

func recursiveIsAllowedInSPF(domain string, ip net.IP, maxDepth, curDepth int) (matched bool, err error) {
	if curDepth > maxDepth {
		return
	}

	records, err := LookupSPF(domain)
	if err != nil || len(records) == 0 {
		return
	}

	fields := strings.Fields(records[0])
	for _, mech := range fields {
		mech = strings.TrimSpace(mech)
		if mech == "" || strings.EqualFold(mech, "v=spf1") {
			continue
		}

		// Strip SPF qualifier.
		if strings.HasPrefix(mech, "-") ||
			strings.HasPrefix(mech, "~") {
			continue
		}

		if strings.HasPrefix(mech, "+") ||
			strings.HasPrefix(mech, "?") {
			mech = mech[1:]
			// Allow all mechanism.
			if mech == "all" {
				return true, nil
			}
		}

		_networks, err := getSPFMechanismNetworks(mech, domain, maxDepth, curDepth)
		if err != nil {
			return false, err
		}

		for _, network := range _networks {
			if ipInCIDROrSingle(network, ip) {
				return true, nil
			}
		}
	}

	return false, nil
}

// getSPFMechanismNetworks retrieves the networks associated with a specific SPF mechanism for a given domain.
func getSPFMechanismNetworks(mech, domain string, maxDepth, curDepth int) (networks []string, err error) {
	var after string
	var ok bool
	var _networks []string

	if after, ok = strings.CutPrefix(mech, "ip4:"); ok {
		networks = append(networks, after)
	} else if after, ok = strings.CutPrefix(mech, "ip6:"); ok {
		networks = append(networks, after)
	} else if mech == "a" {
		_networks, err = getHostNetworks(domain)
		if err != nil {
			return nil, err
		}

		networks = append(networks, _networks...)
	} else if after, ok = strings.CutPrefix(mech, "a/"); ok {
		_networks, _ = getHostNetworks(domain, after)
		networks = append(networks, _networks...)
	} else if after, ok = strings.CutPrefix(mech, "a:"); ok {
		// 处理 `a:<domain>/24`
		_domain, _prefix, _ := strings.CutLast(after, "/")
		_networks, _ = getHostNetworks(_domain, _prefix)
		networks = append(networks, _networks...)
	} else if mech == "mx" || strings.HasPrefix(mech, "mx:") {
		_domain, prefix := parseSPFDomainAndPrefix(mech, "mx", domain)
		mxs, err := net.LookupMX(_domain)
		if err != nil {
			return nil, err
		}

		for _, mx := range mxs {
			host := strings.TrimSuffix(mx.Host, ".")
			if host == "" {
				continue
			}

			_networks, err = getHostNetworks(host, prefix)
			if err != nil {
				return nil, err
			}

			networks = append(networks, _networks...)
		}
	} else if after, ok = strings.CutPrefix(mech, "include:"); ok {
		includeDomain := strings.TrimSpace(after)

		return recursiveGetDomainSPFNetworks(includeDomain, maxDepth, curDepth+1)
	} else if after, ok = strings.CutPrefix(mech, "redirect="); ok {
		redirectDomain := strings.TrimSpace(after)

		return recursiveGetDomainSPFNetworks(redirectDomain, maxDepth, curDepth+1)
	}

	return
}

// parseSPFDomainAndPrefix extracts the domain and CIDR prefix from an SPF mechanism string.
func parseSPFDomainAndPrefix(mech, tag, fallbackDomain string) (domain, prefix string) {
	mech = strings.TrimPrefix(mech, tag) // mx:example.com -> :example.com
	mech = strings.TrimPrefix(mech, ":") // :example.com -> example.com
	mech = strings.TrimSpace(mech)
	if mech == "" {
		return fallbackDomain, ""
	}

	// Handle "a/24" / "mx/24" (no explicit domain, only CIDR prefix).
	if strings.HasPrefix(mech, "/") {
		return fallbackDomain, strings.TrimLeft(mech, "/")
	}

	domain, prefix, _ = strings.Cut(mech, "/")
	domain = strings.TrimSpace(domain)
	prefix = strings.TrimLeft(strings.TrimSpace(prefix), "/")
	if domain == "" {
		domain = fallbackDomain
	}

	return
}

// getClassCNetwork 根据 IP 和 prefix 长度获取对应的 C 类网段。
// 例如：ip 为 192.168.0.1, prefix 为 24，则返回 192.168.0.0/24。
// FIXME 添加测试用例。
func getClassCNetwork(ip net.IP, prefix int) (ipNet *net.IPNet, ipNetStr string) {
	// 统一为 4 字节或 16 字节表示
	if ip4 := ip.To4(); ip4 != nil {
		ip = ip4
	} else if ip16 := ip.To16(); ip16 != nil {
		ip = ip16
	} else {
		return nil, ""
	}

	// 确定总位数
	var bits int
	if len(ip) == net.IPv4len {
		bits = 32
	} else {
		bits = 128
	}

	// 验证 prefix 范围
	if prefix < 0 || prefix > bits {
		return nil, ""
	}

	mask := net.CIDRMask(prefix, bits)
	network := ip.Mask(mask)

	ipNet = &net.IPNet{
		IP:   network,
		Mask: mask,
	}

	ipNetStr = ipNet.String()

	return
}

// getHostNetworks 将 domain 解析为 IPv4/IPv6 地址，并根据 prefix 生成 CIDR 网络表示。
func getHostNetworks(domain string, prefix ...string) (networks []string, err error) {
	var prefixInt int
	if len(prefix) > 0 && prefix[0] != "" {
		prefixInt, err = strconv.Atoi(prefix[0])
		if err != nil {
			return nil, err
		}
	}

	ips, err := net.LookupIP(domain)
	for _, ip := range ips {
		if prefixInt != 0 {
			_, cidr := getClassCNetwork(ip, prefixInt)
			if cidr != "" {
				networks = append(networks, cidr)
			}
		} else {
			networks = append(networks, ip.String())
		}
	}

	return
}

// ipInCIDROrSingle checks if the given IP address is within the specified CIDR range or matches a single IP address.
func ipInCIDROrSingle(value string, clientIP net.IP) bool {
	value = strings.TrimSpace(value)
	if value == "" {
		return false
	}

	if strings.Contains(value, "/") {
		_, cidr, err := net.ParseCIDR(value)
		if err != nil {
			return false
		}

		return cidr.Contains(clientIP)
	}

	parsedIP := net.ParseIP(value)
	if parsedIP == nil {
		return false
	}

	return parsedIP.Equal(clientIP)
}

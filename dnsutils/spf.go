package dnsutils

import (
	"net"
	"strings"
)

const maxDomainSPFDepth = 10

func IsAllowedIPInSPF(domain string, ip net.IP, depth int) (matched bool, err error) {
	if domain == "" || ip == nil || depth > maxDomainSPFDepth {
		return
	}

	records, err := LookupSPF(domain)
	if err != nil || len(records) == 0 {
		return
	}

	fields := strings.Fields(records[0])
	for _, mech := range fields {
		mech = strings.TrimSpace(mech)
		if mech == "" || mech == "all" || strings.EqualFold(mech, "v=spf1") {
			continue
		}

		// Strip SPF qualifier.
		if strings.HasPrefix(mech, "+") ||
			strings.HasPrefix(mech, "-") ||
			strings.HasPrefix(mech, "~") ||
			strings.HasPrefix(mech, "?") {
			mech = mech[1:]
		}

		matched, err = matchSPFMechanism(mech, domain, ip, depth)
		if err != nil || matched {
			return
		}
	}

	return
}

func matchSPFMechanism(mech, domain string, clientIP net.IP, depth int) (bool, error) {
	switch {
	case strings.HasPrefix(mech, "ip4:"):
		return ipInCIDROrSingle(strings.TrimPrefix(mech, "ip4:"), clientIP), nil
	case strings.HasPrefix(mech, "ip6:"):
		return ipInCIDROrSingle(strings.TrimPrefix(mech, "ip6:"), clientIP), nil
	case mech == "a", strings.HasPrefix(mech, "a:"):
		_domain, prefix := parseSPFDomainAndPrefix(mech, "a", domain)

		return matchHostByA(_domain, prefix, clientIP)
	case mech == "mx", strings.HasPrefix(mech, "mx:"):
		_domain, prefix := parseSPFDomainAndPrefix(mech, "mx", domain)
		mxs, err := net.LookupMX(_domain)
		for _, mx := range mxs {
			host := strings.TrimSuffix(mx.Host, ".")
			if host == "" {
				continue
			}

			matched, err := matchHostByA(host, prefix, clientIP)
			if err != nil || matched {
				return matched, err
			}
		}

		return false, err
	case strings.HasPrefix(mech, "include:"):
		includeDomain := strings.TrimSpace(strings.TrimPrefix(mech, "include:"))

		return IsAllowedIPInSPF(includeDomain, clientIP, depth+1)
	case strings.HasPrefix(mech, "redirect="):
		redirectDomain := strings.TrimSpace(strings.TrimPrefix(mech, "redirect="))

		return IsAllowedIPInSPF(redirectDomain, clientIP, depth+1)
	}

	return false, nil
}

func parseSPFDomainAndPrefix(mech, tag, fallbackDomain string) (domain, prefix string) {
	mech = strings.TrimPrefix(mech, tag)
	mech = strings.TrimPrefix(mech, ":")
	if mech == "" {
		return fallbackDomain, ""
	}

	domain = mech
	var found bool
	domain, prefix, found = strings.Cut(mech, "/")
	if found {
		domain = strings.TrimSpace(domain)
	}

	return
}

func matchHostByA(domain, prefix string, clientIP net.IP) (bool, error) {
	ips, err := net.LookupIP(domain)
	for _, ip := range ips {
		_ip := ip.String()
		if prefix != "" {
			_ip += "/" + prefix
		}

		if ipInCIDROrSingle(ip.String(), clientIP) {
			return true, nil
		}
	}

	return false, err
}

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

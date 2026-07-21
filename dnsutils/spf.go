package dnsutils

import (
	"net"
	"strings"
)

const maxDomainSPFDepth = 10

func GetDomainSPFNetworks(domain string, depth ...int) (networks []string, err error) {
	if domain == "" || (len(depth) > 0 && depth[0] > 0 && depth[0] > maxDomainSPFDepth) {
		return
	}

	_depth := 0
	if len(depth) > 0 {
		_depth = depth[0]
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

		_networks, err := getSPFMechanismNetworks(mech, domain, _depth)
		if err != nil {
			return nil, err
		}

		networks = append(networks, _networks...)
	}

	return
}

func IsAllowedIPInSPF(domain string, ip net.IP, depth ...int) (matched bool, err error) {
	if domain == "" || ip == nil || (len(depth) > 0 && depth[0] > 0 && depth[0] > maxDomainSPFDepth) {
		return
	}

	_depth := 0
	if len(depth) > 0 {
		_depth = depth[0]
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

		_networks, err := getSPFMechanismNetworks(mech, domain, _depth)
		if err != nil {
			return false, err
		}

		for _, network := range _networks {
			if ipInCIDROrSingle(network, ip) {
				return true, nil
			}
		}
	}

	return
}

func getSPFMechanismNetworks(mech, domain string, depth int) (networks []string, err error) {
	switch {
	case strings.HasPrefix(mech, "ip4:"):
		networks = append(networks, strings.TrimPrefix(mech, "ip4:"))
	case strings.HasPrefix(mech, "ip6:"):
		networks = append(networks, strings.TrimPrefix(mech, "ip6:"))
	case mech == "a", strings.HasPrefix(mech, "a:"):
		_domain, prefix := parseSPFDomainAndPrefix(mech, "a", domain)
		_networks, err := getHostNetworks(_domain, prefix)
		if err != nil {
			return nil, err
		}

		networks = append(networks, _networks...)
	case mech == "mx", strings.HasPrefix(mech, "mx:"):
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

			_networks, err := getHostNetworks(host, prefix)
			if err != nil {
				return nil, err
			}

			networks = append(networks, _networks...)
		}
	case strings.HasPrefix(mech, "include:"):
		includeDomain := strings.TrimSpace(strings.TrimPrefix(mech, "include:"))

		return GetDomainSPFNetworks(includeDomain, depth+1)
	case strings.HasPrefix(mech, "redirect="):
		redirectDomain := strings.TrimSpace(strings.TrimPrefix(mech, "redirect="))

		return GetDomainSPFNetworks(redirectDomain, depth+1)
	}

	return
}

func parseSPFDomainAndPrefix(mech, tag, fallbackDomain string) (domain, prefix string) {
	mech = strings.TrimPrefix(mech, tag)
	mech = strings.TrimPrefix(mech, ":")
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

func getHostNetworks(domain, prefix string) (networks []string, err error) {
	ips, err := net.LookupIP(domain)
	for _, ip := range ips {
		if prefix != "" {
			networks = append(networks, ip.String()+"/"+prefix)
		} else {
			networks = append(networks, ip.String())
		}
	}

	return
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

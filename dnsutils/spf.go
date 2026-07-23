package dnsutils

import (
	"net"
	"strings"
)

const defaultMaxDomainSPFDepth = 10

// GetDomainSPFNetworks Retrieves the SPF networks for a given domain.
//
//	Default maxDepth is 10, but can be overridden by passing a positive integer as the second argument.
func GetDomainSPFNetworks(domain string, maxDepth ...int) (networks []string, err error) {
	_maxDepth := defaultMaxDomainSPFDepth
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

	fields := strings.Fields(records[0])
	for _, mech := range fields {
		mech = strings.TrimSpace(mech)
		if mech == "" || mech == "all" || strings.EqualFold(mech, "v=spf1") {
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
			// Ignore the "all" mechanism as it does not provide specific network information.
			if mech == "all" {
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
	_maxDepth := defaultMaxDomainSPFDepth
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
	switch {
	case strings.HasPrefix(mech, "ip4:"):
		networks = append(networks, strings.TrimPrefix(mech, "ip4:"))
	case strings.HasPrefix(mech, "ip6:"):
		networks = append(networks, strings.TrimPrefix(mech, "ip6:"))
	case mech == "a", strings.HasPrefix(mech, "a/"), strings.HasPrefix(mech, "a:"):
		_domain, prefix := parseSPFDomainAndPrefix(mech, "a", domain)
		_networks, err := getHostNetworks(_domain, prefix)
		if err != nil {
			return nil, err
		}

		networks = append(networks, _networks...)
	case mech == "mx", strings.HasPrefix(mech, "mx/"), strings.HasPrefix(mech, "mx:"):
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

		return recursiveGetDomainSPFNetworks(includeDomain, maxDepth, curDepth+1)
	case strings.HasPrefix(mech, "redirect="):
		redirectDomain := strings.TrimSpace(strings.TrimPrefix(mech, "redirect="))

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

// getHostNetworks retrieves the IP addresses for a given domain and appends the specified CIDR prefix if provided.
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

package dnsutils

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseSPFDomainAndPrefix(t *testing.T) {
	domain, prefix := parseSPFDomainAndPrefix("a:a.io/24", "a", "a.io")
	assert.Equal(t, "a.io", domain)
	assert.Equal(t, "24", prefix)

	domain, prefix = parseSPFDomainAndPrefix("a:a.io", "a", "a.io")
	assert.Equal(t, "a.io", domain)
	assert.Equal(t, "", prefix)

	domain, prefix = parseSPFDomainAndPrefix("a/24", "a", "a.io")
	assert.Equal(t, "a.io", domain)
	assert.Equal(t, "24", prefix)

	domain, prefix = parseSPFDomainAndPrefix("a:", "a", "a.io")
	assert.Equal(t, "a.io", domain)
	assert.Equal(t, "", prefix)

	domain, prefix = parseSPFDomainAndPrefix("mx:example.com/24", "mx", "example.org")
	assert.Equal(t, "example.com", domain)
	assert.Equal(t, "24", prefix)

	domain, prefix = parseSPFDomainAndPrefix("mx:example.com", "mx", "example.org")
	assert.Equal(t, "example.com", domain)
	assert.Equal(t, "", prefix)

	domain, prefix = parseSPFDomainAndPrefix("mx/24", "mx", "example.org")
	assert.Equal(t, "example.org", domain)
	assert.Equal(t, "24", prefix)

	domain, prefix = parseSPFDomainAndPrefix("mx:", "mx", "example.org")
	assert.Equal(t, "example.org", domain)
	assert.Equal(t, "", prefix)
}

func TestIsAllowedServerInSPFDepthLimit(t *testing.T) {
	// dig +short -t txt example.com
	// "_k2n1y4vw3qtb4skdx9e7dxt97qrmmq9"
	// "v=spf1 -all"
	allowed, err := IsAllowedIPInSPF("example.com", net.ParseIP("203.0.113.10"))
	assert.NoError(t, err)
	assert.False(t, allowed)

	allowed, err = IsAllowedIPInSPF("iredmail.org", net.ParseIP("172.105.68.48"))
	assert.NoError(t, err)
	assert.True(t, allowed)

	allowed, err = IsAllowedIPInSPF("iredmail.org", net.ParseIP("2a01:7e01::f03c:91ff:fe74:9543"))
	assert.NoError(t, err)
	assert.True(t, allowed)

	allowed, err = IsAllowedIPInSPF("gmail.com", net.ParseIP("209.85.128.3"))
	assert.NoError(t, err)
	assert.True(t, allowed)
}

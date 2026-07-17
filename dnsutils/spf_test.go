package dnsutils

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsAllowedServerInSPFDepthLimit(t *testing.T) {
	// dig +short -t txt example.com
	// "_k2n1y4vw3qtb4skdx9e7dxt97qrmmq9"
	// "v=spf1 -all"
	allowed, err := IsAllowedIPInSPF("example.com", net.ParseIP("203.0.113.10"), 0)
	assert.NoError(t, err)
	assert.False(t, allowed)

	allowed, err = IsAllowedIPInSPF("iredmail.org", net.ParseIP("172.105.68.48"), 0)
	assert.NoError(t, err)
	assert.True(t, allowed)

	allowed, err = IsAllowedIPInSPF("iredmail.org", net.ParseIP("2a01:7e01::f03c:91ff:fe74:9543"), 0)
	assert.NoError(t, err)
	assert.True(t, allowed)

	allowed, err = IsAllowedIPInSPF("gmail.com", net.ParseIP("209.85.128.3"), 0)
	assert.NoError(t, err)
	assert.True(t, allowed)
}

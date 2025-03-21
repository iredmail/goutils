package dnsutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQuerySPF(t *testing.T) {
	// domain := "iredmail.org"
	domain := "gmail.com"
	found, result, err := QuerySPF(domain)
	assert.Nil(t, err)
	assert.True(t, found)
	assert.Equal(t, domain, result.Domain)
	assert.Equal(t, "v=spf1 redirect=_spf.google.com", result.Txt)

	domain = "iredmail.org"
	found, result, err = QuerySPF(domain)
	assert.Nil(t, err)
	assert.True(t, found)
	assert.Equal(t, domain, result.Domain)
	assert.Equal(t, "v=spf1 mx:iredmail.org ip4:172.105.68.48 ip6:2a01:7e01::f03c:93ff:fe25:7e10 ip6:2a01:7e01::f03c:91ff:fe74:9543 ip4:172.104.245.227 -all", result.Txt)
}

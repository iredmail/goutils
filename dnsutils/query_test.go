package dnsutils

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueryA(t *testing.T) {
	domain := "mail.iredmail.org"
	found, result, err := QueryA(domain)
	assert.Nil(t, err)
	assert.True(t, found)
	assert.Equal(t, 1, len(result.IPs))
	assert.True(t, slices.Contains(result.IPs, "172.105.68.48"))

	found, _, _ = QueryA("not-exist-abcdefgehize.com")
	assert.False(t, found)
}

func TestQueryAAAA(t *testing.T) {
	domain := "mail.iredmail.org"
	found, result, err := QueryAAAA(domain)
	assert.Nil(t, err)
	assert.True(t, found)
	assert.Equal(t, 1, len(result.IPs))
	assert.True(t, slices.Contains(result.IPs, "2a01:7e01:e001:36f::2"))

	found, _, _ = QueryAAAA("not-exist-abcdefgehize.com")
	assert.False(t, found)
}

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

func TestQueryMX(t *testing.T) {
	domain := "gmail.com"
	found, result, err := QueryMX(domain)
	assert.Nil(t, err)
	assert.True(t, found)
	assert.Equal(t, 5, len(result.Hostnames))
	assert.Equal(t, "alt1.gmail-smtp-in.l.google.com", result.Hostnames[0])
	assert.Equal(t, "alt2.gmail-smtp-in.l.google.com", result.Hostnames[1])
	assert.Equal(t, "alt3.gmail-smtp-in.l.google.com", result.Hostnames[2])
	assert.Equal(t, "alt4.gmail-smtp-in.l.google.com", result.Hostnames[3])
	assert.Equal(t, "gmail-smtp-in.l.google.com", result.Hostnames[4])

	assert.Equal(t, "alt1.gmail-smtp-in.l.google.com", result.Hosts[0].Hostname)
	assert.Equal(t, "alt2.gmail-smtp-in.l.google.com", result.Hosts[1].Hostname)
	assert.Equal(t, "alt3.gmail-smtp-in.l.google.com", result.Hosts[2].Hostname)
	assert.Equal(t, "alt4.gmail-smtp-in.l.google.com", result.Hosts[3].Hostname)
	assert.Equal(t, "gmail-smtp-in.l.google.com", result.Hosts[4].Hostname)
}

func TestQuerySPF(t *testing.T) {
	domain := "iredmail.org"
	found, result, err := QuerySPF(domain)
	assert.Nil(t, err)
	assert.True(t, found)
	assert.Equal(t, domain, result.Domain)
}

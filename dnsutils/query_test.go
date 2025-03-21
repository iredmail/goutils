package dnsutils

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueryA(t *testing.T) {
	domain := "mail.iredmail.org"
	found, result := QueryA(domain)
	assert.True(t, found)
	assert.Nil(t, result.Error)
	assert.Equal(t, 1, len(result.IPs))
	assert.True(t, slices.Contains(result.IPs, "172.105.68.48"))

	found, result = QueryA("not-exist-abcdefgehize.com")
	assert.False(t, found)
	assert.Nil(t, result.Error)
}

func TestQueryAAAA(t *testing.T) {
	domain := "mail.iredmail.org"
	found, result := QueryAAAA(domain)
	assert.Nil(t, result.Error)
	assert.True(t, found)
	assert.Equal(t, 1, len(result.IPs))
	assert.True(t, slices.Contains(result.IPs, "2a01:7e01:e001:36f::2"))

	found, result = QueryAAAA("not-exist-abcdefgehize.com")
	assert.False(t, found)
	assert.Nil(t, result.Error)
}

func TestQueryMX(t *testing.T) {
	domain := "gmail.com"

	found, result := QueryMX(domain)
	assert.Nil(t, result.Error)
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

func TestQueryAll(t *testing.T) {
	domain := "iredmail.org"

	result := QueryAll(domain)
	assert.Nil(t, result.ResultA.Error)
	assert.Nil(t, result.ResultAAAA.Error)
	assert.Nil(t, result.ResultMX.Error)
	assert.Nil(t, result.ResultSPF.Error)
	assert.Nil(t, result.ResultDKIM.Error)
	assert.Nil(t, result.ResultDMARC.Error)
}

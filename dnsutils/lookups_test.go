package dnsutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLookupSPF(t *testing.T) {
	r, _ := NewResolver()
	domain := "gmail.com"
	records, err := r.LookupSPF(domain)
	assert.Nil(t, err)
	assert.Equal(t, records, []string{"v=spf1 redirect=_spf.google.com"})
}

func TestLookupRecursiveSPF(t *testing.T) {
	r, _ := NewResolver(WithDNSServer("8.8.8.8:53"))

	domain := "gmail.com"
	records, totalQueries, err := r.LookupRecursiveSPF(domain, 0)
	assert.Nil(t, err)
	assert.Equal(t, totalQueries, 2)
	assert.Equal(t, records, []string{"v=spf1 redirect=_spf.google.com"})
}

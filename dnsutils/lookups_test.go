package dnsutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLookupSPF(t *testing.T) {
	r := NewResolver(10)
	domain := "gmail.com"
	notfound, records, err := r.LookupSPF(domain)
	assert.True(t, err == "")
	assert.False(t, notfound)
	assert.Equal(t, records, []string{"v=spf1 redirect=_spf.google.com"})
}

func TestLookupRecursiveSPF(t *testing.T) {
	r := NewResolver(10, "8.8.8.8:53")

	domain := "gmail.com"
	notfound, records, totalQueries, err := r.LookupRecursiveSPF(domain, 0)
	assert.True(t, err == "")
	assert.False(t, notfound)
	assert.Equal(t, totalQueries, 2)
	assert.Equal(t, records, []string{"v=spf1 redirect=_spf.google.com"})
}

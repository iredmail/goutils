package dnsutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLookupSPF(t *testing.T) {
	domain := "gmail.com"
	records, err := LookupSPF(domain)
	assert.Nil(t, err)
	assert.Equal(t, records, []string{"v=spf1 redirect=_spf.google.com"})
}

func TestLookupRecursiveSPF(t *testing.T) {
	domain := "gmail.com"
	records, totalQueries, err := LookupRecursiveSPF(domain, 0)
	assert.Nil(t, err)
	assert.Equal(t, totalQueries, 2)
	assert.Equal(t, records, []string{"v=spf1 redirect=_spf.google.com"})
}

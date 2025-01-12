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
}

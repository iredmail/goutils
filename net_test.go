package goutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNet(t *testing.T) {
	assert.Equal(t, true, IsIP("192.168.0.1"))
	assert.Equal(t, true, IsIP("192.168.0.0"))
	assert.Equal(t, false, IsIP("192.168.0.0/24"))

	assert.Equal(t, true, IsCIDR("192.168.0.0/24"))
	assert.Equal(t, false, IsCIDR("192.168.0.0"))
}

func TestIsWildcardAddr(t *testing.T) {
	assert.Equal(t, true, IsWildcardAddr("user@*"))
	assert.Equal(t, true, IsWildcardAddr("-@*"))
	assert.Equal(t, true, IsWildcardAddr("user-@*"))
	assert.Equal(t, false, IsWildcardAddr("user@abc"))
	assert.Equal(t, false, IsWildcardAddr("*@*"))
}

func TestIsWildcardIPv4(t *testing.T) {
	// Test cases with one wildcard
	assert.Equal(t, true, IsWildcardIPv4("192.168.0.*"))
	assert.Equal(t, true, IsWildcardIPv4("192.168.*.1"))
	assert.Equal(t, true, IsWildcardIPv4("192.*.0.1"))
	assert.Equal(t, true, IsWildcardIPv4("*.168.0.1"))

	// Test cases with two wildcards
	assert.Equal(t, true, IsWildcardIPv4("192.168.*.*"))
	assert.Equal(t, true, IsWildcardIPv4("192.*.*.1"))
	assert.Equal(t, true, IsWildcardIPv4("*.*.0.1"))

	// Test cases with three wildcards
	assert.Equal(t, true, IsWildcardIPv4("192.*.*.*"))
	assert.Equal(t, true, IsWildcardIPv4("*.168.*.*"))
	assert.Equal(t, true, IsWildcardIPv4("*.*.0.*"))

	// Test cases with four wildcards
	assert.Equal(t, true, IsWildcardIPv4("*.*.*.*"))

	// Test cases without wildcards
	assert.Equal(t, false, IsWildcardIPv4("192.168.0.1"))
	assert.Equal(t, false, IsWildcardIPv4("192.168.0.0/24"))
	assert.Equal(t, false, IsWildcardIPv4("192.168.0.0"))

	// Test cases with wildcards but not valid IP addresses
	assert.Equal(t, false, IsWildcardIPv4("192.168.*.0/24"))
	assert.Equal(t, false, IsWildcardIPv4("192.168.*.256"))
	assert.Equal(t, false, IsWildcardIPv4("192.168.*.*.1"))
	assert.Equal(t, false, IsWildcardIPv4("192.168.*.*.*"))
	assert.Equal(t, false, IsWildcardIPv4("192.168.*.1.1"))

	// Test cases with invalid inputs
	assert.Equal(t, false, IsWildcardIPv4("192.168.*"))
	assert.Equal(t, false, IsWildcardIPv4("192.168.*."))
	assert.Equal(t, false, IsWildcardIPv4("192.168..*"))
	assert.Equal(t, false, IsWildcardIPv4("192.168.*.."))
	assert.Equal(t, false, IsWildcardIPv4("192.168.*.*."))
}

package emailutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmail(t *testing.T) {
	// IsEmail
	assert.False(t, IsEmail("abc"))
	assert.False(t, IsEmail("abc.com"))
	assert.True(t, IsEmail("user@abc.com"))

	// IsDomain
	assert.True(t, IsDomain("abc.com"))
	assert.True(t, IsDomain("0.io"))
	assert.True(t, IsDomain("x.io"))
	assert.True(t, IsDomain("0000.io"))
	assert.True(t, IsDomain("u22.x.io"))
	assert.False(t, IsDomain("com"))
	assert.False(t, IsDomain("abcdefg"))
	assert.False(t, IsDomain("1234"))

	assert.Equal(t, ExtractDomain("user@A.io"), "a.io")
	assert.Equal(t, ExtractDomain("user@[192.168.1.1]"), "192.168.1.1")

	assert.Equal(t, ExtractUsername("user"), "user")          // invalid email
	assert.Equal(t, ExtractUsername("user@A.io"), "user")     // valid
	assert.Equal(t, ExtractUsername("user+ext@A.io"), "user") // valid with extension

	// Username address extension
	assert.Equal(t, StripExtension("User@A.Io"), "user@a.io")
	assert.Equal(t, StripExtension("User+ext-123=456@a.iO"), "user@a.io")
	assert.Equal(t, StripExtension("User-123=456@A.iO"), "user-123=456@a.io")

	// Parse email addresses.
	expected := `"Name" <u@d.io>`
	addrs := []string{
		`Name <u@d.io>`,     // 正常
		`Name <'u@d.io''>`,  // email 地址带单引号
		`'Name' <'u@d.io'>`, // 名字和 email 都带单引号
	}

	for _, v := range addrs {
		addr, err := ParseAddress(v)
		assert.Nil(t, err)
		assert.Equal(t, expected, addr.String())
	}

	// 使用 IP 地址作为域名。
	/*
		expected = `"Name" <u@[172.16.1.1]>`
		addrs = []string{
			`Name <u@[172.16.1.1]>`,     // 正常
			`Name <'u@[172.16.1.1]'>`,   // email 地址带单引号
			`'Name' <'u@[172.16.1.1]'>`, // 名字和 email 都带单引号
		}

		for _, v := range addrs {
			addr, err := ParseAddress(v)
			assert.Nil(t, err)
			assert.Equal(t, expected, addr.String())
		}
	*/
}

func TestNetwork(t *testing.T) {
	assert.True(t, IsStrictIP("192.168.2.113"))
	assert.False(t, IsStrictIP("192.168.2.1132"))
	assert.False(t, IsStrictIP("192.2.1132"))

	assert.True(t, IsCIDRNetwork("192.168.2.1/24"))
	assert.False(t, IsCIDRNetwork("192.168.2.1"))

	assert.True(t, IsWildcardAddr("172.13.1.*"))
	assert.True(t, IsWildcardAddr("172.13.*.1"))
	assert.True(t, IsWildcardAddr("172.*.1.1"))
	assert.False(t, IsWildcardAddr("172.2.*"))
	assert.False(t, IsWildcardAddr("172.256.*.1"))

	assert.True(t, IsWildcardIPv4("172.31.1.*"))
	assert.False(t, IsWildcardIPv4("172.256.1.*"))
	assert.False(t, IsWildcardIPv4("172.1.*"))
}

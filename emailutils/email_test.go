package email

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
	assert.True(t, IsDomain("0000.io"))
	assert.False(t, IsDomain("com"))
	assert.False(t, IsDomain("abcdefg"))
	assert.False(t, IsDomain("1234"))

	assert.Equal(t, ExtractDomainFromEmail("user@A.io"), "a.io")
	assert.Equal(t, ExtractDomainFromEmail("user@[192.168.1.1]"), "192.168.1.1")

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

package emailutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsEmail(t *testing.T) {
	// Invalid
	assert.False(t, IsEmail("abc"))
	assert.False(t, IsEmail("abc.com"))
	assert.False(t, IsEmail("user.123@abc@abc.com"))
	assert.False(t, IsEmail("user@domain"))
	assert.False(t, IsEmail("win-q1hgaa9hc1g$@b.io")) // $

	// Valid
	assert.True(t, IsEmail("user@abc.com"))
	assert.True(t, IsEmail("user+abc@abc.com"))
	assert.True(t, IsEmail("user.abc@abc.com"))
	assert.True(t, IsEmail("user.123@abc.com"))
	assert.True(t, IsEmail("user@sub3.sub2.sub1.com"))
	assert.True(t, IsEmail("lcastaã±eda@ruska.com.pe"))
	assert.True(t, IsEmail("lcastaeda@ruska.com.pe"))
}

func TestIsDomain(t *testing.T) {
	// IsDomain
	assert.True(t, IsDomain("abc.com"))
	assert.True(t, IsDomain("0.io"))
	assert.True(t, IsDomain("x.io"))
	assert.True(t, IsDomain("0000.io"))
	assert.True(t, IsDomain("u22.x.io"))
	assert.False(t, IsDomain("com"))
	assert.False(t, IsDomain("abcdefg"))
	assert.False(t, IsDomain("114.114.114.114"))
	assert.False(t, IsDomain("1234"))
	assert.False(t, IsDomain(".abc.com"))
}

func TestIsFQDN(t *testing.T) {
	// IsFQDN
	assert.True(t, IsFQDN("mail.demo.io"))
	assert.False(t, IsFQDN("demo"))
}

func TestExtractDomain(t *testing.T) {
	assert.Equal(t, ExtractDomain("user@A.io"), "a.io")
	assert.Equal(t, ExtractDomain("user@[192.168.1.1]"), "192.168.1.1")
}

func TestExtractDomains(t *testing.T) {
	assert.Equal(t, 0, len(ExtractDomains([]string{})))
	assert.Equal(t, []string{"a.io"}, ExtractDomains([]string{"user@A.io"}))
	assert.Equal(t, []string{"a.io", "b.io"}, ExtractDomains([]string{"user@A.io", "user2@a.IO", "user@B.io"}))
	assert.Equal(t, []string{"a.io", "b.io", "c.io"}, ExtractDomains([]string{"1@c.io", "1@a.io", "2@b.IO", "1@b.io", "2@a.IO", "3@c.io"}))
}

func TestExtractUsername(t *testing.T) {
	assert.Equal(t, ExtractUsername("user"), "user")          // invalid email
	assert.Equal(t, ExtractUsername("user@A.io"), "user")     // valid
	assert.Equal(t, ExtractUsername("user+ext@A.io"), "user") // valid with extension
}

func TestStripExtension(t *testing.T) {
	// Username address extension
	assert.Equal(t, StripExtension("User@A.Io"), "user@a.io")
	assert.Equal(t, StripExtension("User+ext-123=456@a.iO"), "user@a.io")
	assert.Equal(t, StripExtension("User-123=456@A.iO"), "user-123=456@a.io")
}

func TestParseAddress(t *testing.T) {
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

	// iso-8859-9
	addr, err := ParseAddress("=?iso-8859-9?Q?Javuz_Ma=FElak?= <user@domain.tr>")
	assert.Nil(t, err)
	assert.Equal(t, "Javuz Maşlak", addr.Name)
	assert.Equal(t, "user@domain.tr", addr.Address)

	// Name 和 Address 之间有换行。Microsoft 发出的邮件常有这样的格式。
	s := `Microsoft account team
	<account-security-noreply@accountprotection.microsoft.com>`
	addr, err = ParseAddress(s)
	assert.Nil(t, err)
	assert.Equal(t, "Microsoft account team", addr.Name)
	assert.Equal(t, "account-security-noreply@accountprotection.microsoft.com", addr.Address)

	s = `"=?utf-8?Q?=E5=BC=80=E6=99=AE=E7=A5=A8=E5=8A=A0=E5=BE=AE=E4=BF=A1:17376744614
 =E6=99=95=E6=9B=B4=E5=AE=B9=E7=A0=B8=E5=85=85=E7=BA=B9=E5=8D=89=E7=BB=85=E6=84=8Finfo@iredmail.orginfo@iredmail.org?=" <cwlqlwqgq@ohwtf.com>`
	addr, err = ParseAddress(s)
	assert.Nil(t, err)
	assert.Equal(t, "cwlqlwqgq@ohwtf.com", addr.Address)

	s = `GameStop & Comenity Capital Bank 
 <GameStop-apply@e.breadfinancial.com>`
	addr, err = ParseAddress(s)
	assert.Nil(t, err)
	assert.Equal(t, "GameStop & Comenity Capital Bank", addr.Name)
	assert.Equal(t, "GameStop-apply@e.breadfinancial.com", addr.Address)

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

	// gb2312，且 display name 里带逗号
	addr, err = ParseAddress("=?GB2312?B?zuTD97+tIFd1LCBNaW5na2Fp?= <user@domain.com>")
	assert.Nil(t, err)
	assert.Equal(t, "武明凯 Wu, Mingkai", addr.Name)
	assert.Equal(t, "user@domain.com", addr.Address)

	// Name 里有多个 `<`
	addr, err = ParseAddress("Display < Name <user@domain.com>")
	assert.Nil(t, err)
	assert.Equal(t, "Display < Name", addr.Name)
	assert.Equal(t, "user@domain.com", addr.Address)

	// Name 里有逗号
	addr, err = ParseAddress("Display, Name <user@domain.com>")
	assert.Nil(t, err)
	assert.Equal(t, "Display, Name", addr.Name)
	assert.Equal(t, "user@domain.com", addr.Address)
}
func TestFilterValidEmails(t *testing.T) {
	emails := []string{"a", "b.io", "user@c.io", "d@", "e@f.com", "g+ext@h.com"}
	valid, invalid := FilterValidEmails(emails)
	assert.Equal(t, []string{"user@c.io", "e@f.com", "g+ext@h.com"}, valid)
	assert.Equal(t, []string{"a", "b.io", "d@"}, invalid)
}

func TestFilterValidDomains(t *testing.T) {
	domains := []string{"a", "b.io", "test.com", "b"}
	valid, invalid := FilterValidDomains(domains)
	assert.Equal(t, []string{"b.io", "test.com"}, valid)
	assert.Equal(t, []string{"a", "b"}, invalid)
}

func TestNetwork(t *testing.T) {
	assert.True(t, IsWildcardAddr("172.13.1.*"))
	assert.True(t, IsWildcardAddr("172.13.*.1"))
	assert.True(t, IsWildcardAddr("172.*.1.1"))
	assert.False(t, IsWildcardAddr("172.2.*"))
	assert.False(t, IsWildcardAddr("172.256.*.1"))

	assert.True(t, IsWildcardIPv4("172.31.1.*"))
	assert.False(t, IsWildcardIPv4("172.256.1.*"))
	assert.False(t, IsWildcardIPv4("172.1.*"))
}

func TestToLower(t *testing.T) {
	// Test with a valid email address
	email := "UsEr+LoG@ExAmPlE.CoM"
	expected := "user+LoG@example.com"
	result := ToLowerWithExt(email)
	assert.Equal(t, expected, result)

	// Test with an email address with no extension
	email = "UsEr@ExAmple.CoM"
	expected = "user@example.com"
	result = ToLowerWithExt(email)
	assert.Equal(t, expected, result)

	// Test with an invalid email address
	email = "invalid email address"
	expected = "invalid email address"
	result = ToLowerWithExt(email)
	assert.Equal(t, expected, result)

	// Test with an email address with no username
	email = "example.com"
	expected = "example.com"
	result = ToLowerWithExt(email)
	assert.Equal(t, expected, result)

	// Test with an email address with no domain
	email = "user@example"
	expected = "user@example"
	result = ToLowerWithExt(email)
	assert.Equal(t, expected, result)
}

func TestIsTLDDomain(t *testing.T) {
	// Valid TLD domains
	assert.True(t, IsTLDDomain("aB"))  // Minimum length (2 characters)
	assert.True(t, IsTLDDomain("a-B")) // Hyphen with mixed case
	assert.True(t, IsTLDDomain("A-b")) // Hyphen with mixed case
	assert.True(t, IsTLDDomain("com"))
	assert.True(t, IsTLDDomain("org"))
	assert.True(t, IsTLDDomain("COM"))
	assert.True(t, IsTLDDomain("ORG"))
	assert.True(t, IsTLDDomain("Co"))
	assert.True(t, IsTLDDomain("iO"))
	assert.True(t, IsTLDDomain("APP"))
	assert.True(t, IsTLDDomain("Example"))
	assert.True(t, IsTLDDomain("a-B-c"))
	assert.True(t, IsTLDDomain("X-1-2-3"))
	assert.True(t, IsTLDDomain("abcDEFghiJKLmnoPQRstuVWXY")) // 25 characters
	assert.True(t, IsTLDDomain("ABCDEFGHIJKLMNOPQRSTUVWXY")) // 25 uppercase characters

	// Invalid TLD domains
	assert.False(t, IsTLDDomain(""))
	assert.False(t, IsTLDDomain("a"))                          // Too short
	assert.False(t, IsTLDDomain("abcdefghijklmnopqrstuvwxyz")) // Too long (26 characters)
	assert.False(t, IsTLDDomain("com."))
	assert.False(t, IsTLDDomain(".com"))
	assert.False(t, IsTLDDomain("example.com"))
	assert.False(t, IsTLDDomain("c_o_m"))       // Underscore
	assert.False(t, IsTLDDomain("домен"))       // Non-ASCII characters
	assert.False(t, IsTLDDomain("com!"))        // Special character
	assert.False(t, IsTLDDomain("-com"))        // Starts with hyphen
	assert.False(t, IsTLDDomain("com-"))        // Ends with hyphen
	assert.False(t, IsTLDDomain("192.168.1.1")) // IP address
	assert.False(t, IsTLDDomain("local host"))  // Space not allowed
	assert.False(t, IsTLDDomain("A"))           // Single uppercase character (too short)
}

func TestExtractEmailsInCommaString(t *testing.T) {
	type D struct {
		input  string
		output []string
	}

	data := []D{
		// One email
		{"u1@a.io", []string{"u1@a.io"}},

		// Multiple emails
		{"u1@a.io, u2@a.io, u3@a.io", []string{"u1@a.io", "u2@a.io", "u3@a.io"}},

		// Invalid emails will be discarded
		{"u1@a.io, u2@a.io, invalid-email", []string{"u1@a.io", "u2@a.io"}},

		// Email with address extension
		{"u1@a.io, u2+Ext@a.io", []string{"u1@a.io", "u2+Ext@a.io"}},

		// Email with upper cases in username and domain parts.
		{"UsEr1@A.iO, uSeR2+ExT@a.io", []string{"user1@a.io", "user2+ExT@a.io"}},
	}

	for _, d := range data {
		emails := ExtractEmailsInCommaString(d.input)
		assert.Equal(t, d.output, emails)
	}
}

package emailutils

import (
	"errors"
	"fmt"
	"net"
	"net/mail"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/exp/slices"
)

var (
	regexEmail     = regexp.MustCompile(`^[a-zA-Z0-9.!#$%&'*+\/=?^_` + "`" + `{|}~-]+@[a-zA-Z0-9\.\-]+\.[a-zA-Z0-9]{2,25}$`)
	regexDomain    = regexp.MustCompile(`^[a-zA-Z0-9\.\-]+\.[a-zA-Z0-9]{2,25}$`)
	regexTLDDomain = regexp.MustCompile(`[a-z0-9\-]{2,25}`)
	regexFQDN      = regexp.MustCompile(`^([a-zA-Z0-9]{1}[a-zA-Z0-9-]{0,62})(\.[a-zA-Z0-9]{1}[a-zA-Z0-9-]{0,62})*?(\.[a-zA-Z]{1}[a-zA-Z0-9]{0,62})\.?$`)
)

// IsEmail 校验给定字符串是否为格式正确的邮件地址。
func IsEmail(s string) bool {
	if len(s) < 3 || len(s) > 254 {
		return false
	}

	return regexEmail.MatchString(s)
}

func IsFQDN(s string) bool {
	return regexFQDN.MatchString(s)
}

// IsDomain 校验给定字符串是否为格式正确的邮件域名。
func IsDomain(s string) bool {
	if len(s) < 4 || len(s) > 254 {
		return false
	}

	return regexDomain.MatchString(s)
}

func IsTLDDomain(d string) bool {
	if !IsDomain(d) {
		return false
	}

	return regexTLDDomain.MatchString(d)
}

func IsWildcardAddr(s string) bool {
	s = strings.ReplaceAll(s, "*", "1")

	return net.ParseIP(s) != nil
}
func IsStrictIP(s string) bool {
	return net.ParseIP(s) != nil
}
func IsCIDRNetwork(s string) bool {
	_, _, err := net.ParseCIDR(s)

	return err == nil
}
func IsWildcardIPv4(s string) bool {
	s = strings.ReplaceAll(s, "*", "1")
	ip := net.ParseIP(s)
	if ip == nil {
		return false
	}

	return ip.To4() != nil
}

// ExtractUsername returns username (without extension) of email address.
// If s is not a valid email address, s is returned.
func ExtractUsername(s string) string {
	if !IsEmail(s) {
		return s
	}

	return strings.Split(StripExtension(s), "@")[0]
}

// ExtractDomain 返回邮件地址里的（转换为小写字母的）域名部分。
// 如果域名是 IP 地址（如：`[192.168.1.1]`），则返回（不含中括号的）IP 地址。
func ExtractDomain(e string) string {
	parts := strings.Split(e, "@")
	domain := strings.ToLower(parts[len(parts)-1])

	if strings.HasPrefix(domain, "[") {
		// IP address.
		d1 := strings.TrimPrefix(domain, "[")
		d2 := strings.TrimSuffix(d1, "]")

		return d2
	}

	return domain
}

// ExtractEmailLocalPart 返回邮件地址里的 local part 部分。
func ExtractEmailLocalPart(e string) (string, error) {
	parts := strings.Split(e, "@")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid archiving address: %s", e)
	}

	return parts[0], nil
}

// ExtractDomainFromEmail 返回邮件地址里的（转换为小写字母的）域名部分。
// 如果域名是 IP 地址（如：`[192.168.1.1]`），则返回（不含中括号的）IP 地址。
func ExtractDomainFromEmail(e string) string {
	parts := strings.Split(e, "@")
	domain := parts[len(parts)-1]

	if strings.HasPrefix(domain, "[") {
		// IP address.
		d1 := strings.TrimPrefix(domain, "[")
		d2 := strings.TrimSuffix(d1, "]")

		return d2
	}

	return strings.ToLower(domain)
}

// ExtractDomains 从多个邮件地址里提取邮件域名并转换为小写。
func ExtractDomains(emails []string) (domains []string) {
	for _, addr := range emails {
		d := ExtractDomain(addr)

		if !slices.Contains(domains, d) {
			domains = append(domains, strings.ToLower(d))
		}
	}

	return domains
}

// StripExtension 移除邮件地址里的 `+extension` 扩展，并将邮件地址转换为小写。
// 如果 `email` 不是有效的邮件地址格式，则原样返回。
func StripExtension(email string) string {
	if !IsEmail(email) {
		return email
	}

	username, domain, found := strings.Cut(email, "@")
	if !found {
		return email
	}

	username, _, found = strings.Cut(username, "+")
	if !found {
		return email
	}

	return strings.ToLower(username + "@" + domain)
}

// ParseAddress 是 `mail.ParseAddress()` 的简单封装：
// - 去除首尾的引号
// - 将邮件地址转换为小写
// FIXME Go 官方的 `mail.ParseAddress()` 不支持一些不规范的地址，如 `Name <user@[172.16.1.1]>`。
//
//	考虑用第三方库代替，否则配置参数里的 archiving_domain 归档邮件域名不能用内部 IP 地址。
func ParseAddress(address string) (*mail.Address, error) {
	addr, err := mail.ParseAddress(address)
	if err != nil {
		// 移除错误信息前面的 `mail: ` 字符
		return nil, errors.New(strings.TrimPrefix(err.Error(), "mail: "))
	}

	// 去掉首尾的引号。部分 Microsoft Outlook 客户端会带上引号。
	newName := strings.Trim(addr.Name, `'"`)
	newAddr := strings.Trim(addr.Address, `'"`)

	return &mail.Address{
		Name:    newName,
		Address: strings.ToLower(newAddr),
	}, nil
}

// ExtractEmailsFromAddressList 从 `To:`, `Cc:` 等含有多个邮件地址的字符串里提取（不包含地址扩展的）完整邮件地址。
// - 去除首位的引号
// - 将邮件地址转换为小写
func ExtractEmailsFromAddressList(s string) (emails []string, err error) {
	addrs, err := mail.ParseAddressList(s)
	if err != nil {
		return nil, err
	}

	for _, addr := range addrs {
		// 去掉首尾的引号。部分 Microsoft Outlook 客户端会带上引号。
		i := strings.Trim(addr.Address, `'"`)

		// 去掉地址扩展并转换为小写
		i = strings.ToLower(StripExtension(i))

		emails = append(emails, i)
	}

	return
}

// IsValidASCIIHeaderName 判断给定的邮件头名是否仅含有 ASCII 字符。
//
// RFC822 里邮件头（field）的规范是：
//
//	field =  field-name ":" [ field-body ] CRLF
//
// 邮件头名称（field-name）的规范是：
//
//	field-name =  1*<any CHAR, excluding CTLs, SPACE, and ":">
//
// CHAR 的定义是：
//
//	CHAR =  <any ASCII character>        ; (  0-177,  0.-127.)
func IsValidASCIIHeaderName(name string) bool {
	for i := 0; i < len(name); i++ {
		if name[i] > unicode.MaxASCII {
			return false
		}
	}

	return true
}

func FilterValidEmails(addrs []string) (valid []string, invalid []string) {
	for _, addr := range addrs {
		if IsEmail(addr) {
			valid = append(valid, strings.ToLower(addr))
		} else {
			invalid = append(invalid, strings.ToLower(addr))
		}
	}

	return
}

func FilterValidDomains(domains []string) (valid []string, invalid []string) {
	for _, d := range domains {
		if IsDomain(d) {
			valid = append(valid, strings.ToLower(d))
		} else {
			invalid = append(invalid, strings.ToLower(d))
		}
	}

	return
}

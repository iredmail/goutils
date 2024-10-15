package emailutils

import (
	"errors"
	"fmt"
	"net"
	"net/mail"
	"regexp"
	"slices"
	"strings"
	"unicode"

	"github.com/iredmail/goutils"
)

var (
	regexDomain = regexp.MustCompile(`^[a-zA-Z0-9\.\-]+\.[a-zA-Z0-9]{2,25}$`)

	// - 以字母或数字开头，长度为 2-25 个字符
	// - 不能以 `-` 结尾
	regexTLDDomain = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9\-]{0,23}[a-zA-Z0-9]$`)

	// FQDN 域名的首字母
	regexValidDomainFirstChar = regexp.MustCompile(`^[0-9a-zA-Z]{1,1}$`)

	// FQDN 域名
	regexFQDN = regexp.MustCompile(`^([a-zA-Z0-9]{1}[a-zA-Z0-9-]{0,62})(\.[a-zA-Z0-9]{1}[a-zA-Z0-9-]{0,62})*?(\.[a-zA-Z]{1}[a-zA-Z0-9]{0,62})\.?$`)
)

// IsEmail 校验给定字符串是否为格式正确的邮件地址。
func IsEmail(s string) bool {
	_, err := mail.ParseAddress(strings.TrimSpace(s))

	if err != nil {
		return false
	}

	// net/mail 认为 `user@domain` 是合法的邮件地址。
	_, domain, found := strings.Cut(s, "@")
	if !found {
		return false
	}

	if !IsDomain(domain) {
		return false
	}

	return true
}

func IsFQDN(s string) bool {
	return regexFQDN.MatchString(s)
}

// IsDomain 校验给定字符串是否为格式正确的邮件域名。
func IsDomain(s string) bool {
	if len(s) < 4 || len(s) > 254 {
		return false
	}

	if goutils.IsIPv4(s) {
		return false
	}

	return regexDomain.MatchString(s)
}

func IsTLDDomain(d string) bool {
	return regexTLDDomain.MatchString(d)
}

func IsWildcardAddr(s string) bool {
	s = strings.ReplaceAll(s, "*", "1")

	return net.ParseIP(s) != nil
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
	_, domain, found := strings.Cut(e, "@")
	if !found {
		return ""
	}

	// IP address.
	if strings.HasPrefix(domain, "[") && strings.HasSuffix(domain, "]") {
		domain = strings.Trim(domain, "[]")
	}

	return strings.ToLower(domain)
}

// ExtractDomains 从多个邮件地址里提取邮件域名并转换为小写。
func ExtractDomains(emails []string) (domains []string) {
	for _, addr := range emails {
		domains = append(domains, ExtractDomain(addr))
	}
	slices.Sort(domains)

	return slices.Compact(domains)
}

// ExtractUsernameAndDomain 从给定的 s 里提取用户名和域名。
// 如果 `s` 不是有效的邮件地址，`isValidEmail` 为 false。
func ExtractUsernameAndDomain(s string) (username, domain string, isValidEmail bool) {
	s = StripExtension(s)

	return strings.Cut(s, "@")
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

// StripExtension 移除邮件地址里的 `+extension` 扩展。
// 注意：始终将 `email` 转换为小写再返回。
func StripExtension(email string) string {
	email = strings.ToLower(email)

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

	return username + "@" + domain
}

// ParseAddress 是 `mail.ParseAddress()` 的简单封装：
// - 去除首尾的引号
// - 将邮件地址转换为小写
// 注意：
//   - 自 Go 1.22.2 起，邮件地址的域名部分可以用 `[IP]` 格式。
//   - `mail.ParseAddress()` 处理的是 RFC 5322 address，是经过 base64 encode 后的字符串。
func ParseAddress(address string) (addr *mail.Address, err error) {
	// FIXME 考虑用第三方库代替，否则配置参数里的 archiving_domain 归档邮件域名不能用内部 IP 地址。
	addr, err = mail.ParseAddress(address)
	if err != nil {
		// 移除错误信息前面的 `mail: ` 字符
		return nil, errors.New(strings.TrimPrefix(err.Error(), "mail: "))
	}

	// 去掉首尾的引号。部分 Microsoft Outlook 客户端会带上引号。
	addr.Name = strings.Trim(addr.Name, `'"`)
	addr.Address = strings.Trim(addr.Address, `'"`)

	return
}

// ExtractEmailsFromAddressList 从 `To:`, `Cc:` 等含有多个邮件地址的邮件头的值里提取完整邮件地址。
// 注意：返回的邮件地址都是小写、不包含地址扩展。
func ExtractEmailsFromAddressList(s string) (emails []string, err error) {
	addrs, err := mail.ParseAddressList(s)
	if err != nil {
		return
	}

	for _, addr := range addrs {
		// 去掉���址扩展（并转换为小写）
		emails = append(emails, StripExtension(addr.Address))
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
	for i := range len(name) {
		if name[i] > unicode.MaxASCII {
			return false
		}
	}

	return true
}

// ToLowerWithExt 将邮件地址转换为小写，但保留地址扩展部分（+extension）的大小写。
// 例如：UsEr+LoG@ExAmPlE.CoM -> user+LoG@example.com。
// 注意：传入的 `s` 必须是合法的邮件地址，ToLowerWithExt 内部不检查其是否合法。
func ToLowerWithExt(s string) string {
	userExt, domain, _ := strings.Cut(s, "@")
	username, ext, found := strings.Cut(userExt, "+")
	if found {
		return fmt.Sprintf("%s+%s@%s", strings.ToLower(username), ext, strings.ToLower(domain))
	} else {
		return strings.ToLower(s)
	}
}

// ToLowerWithoutExt 将邮件地址转换为小写，并且移除地址扩展（+extension）。
// 例如：UsEr+LoG@ExAmPlE.CoM -> user@example.com。
func ToLowerWithoutExt(s string) string {
	if !IsEmail(s) {
		return s
	}

	return StripExtension(s)
}

func ObfuscateAddresses(emails ...string) (obfuscated []string) {
	if len(emails) == 0 {
		return
	}

	slices.Sort(emails)
	for _, email := range emails {
		if !IsEmail(email) {
			continue
		}

		email = ToLowerWithExt(email)

		username, domain, _ := strings.Cut(email, "@")
		if len(username) == 1 {
			// u@ -> u*@
			username = username[:1] + "*"
		} else {
			// user@ -> us*@
			username = username[:2] + "*"
		}

		if len(domain) == 3 || len(domain) == 4 || len(domain) == 5 {
			// x.y -> *.y
			// x.io -> *.io
			domain = "*" + domain[1:]
		} else if len(domain) == 6 {
			// abc.io -> **c.io
			domain = "**" + domain[2:]
		} else {
			// abcdefg.io -> ***defg.io
			domain = "***" + domain[3:]
		}

		obfuscated = append(obfuscated, username+"@"+domain)
	}

	return
}

/*
func ReverseDomain(domain string) string {
	split := strings.Split(domain, ".")
	slices.Reverse(split)

	return strings.Join(split, ".")
}

func ReverseDomains(domains []string) []string {
	for i, d := range domains {
		domains[i] = ReverseDomain(d)
	}

	return domains
}
*/

// ExtractEmailsInCommaString extracts email addresses from a string which
// contains one or multiple email address separated by comma.
//
// Notes:
//   - Invalid and duplicate emails will be discarded.
//   - Username and domain parts will be converted to lower cases.
//   - Address extension will be kept (with same upper/lower cases).
func ExtractEmailsInCommaString(s string) (mails []string) {
	for _, addr := range strings.Split(s, ",") {
		addr = strings.TrimSpace(addr)

		if IsEmail(addr) {
			addr = ToLowerWithExt(addr)

			if !slices.Contains(mails, addr) {
				mails = append(mails, addr)
			}
		}
	}

	return
}

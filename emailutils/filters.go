package emailutils

import (
	"slices"
	"strings"
)

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

// FilterSameDomainEmails 返回同域名内的邮件地址和外部域名的邮件地址。
// same 里所有邮件地址里的地址扩展（+ext）都被移除，但在 `orig` 的 Value 里保留了地址扩展。
// orig 是以移除地址扩展后的地址作为 key，以原始
func FilterSameDomainEmails(domain string, emails []string) (same []string, others []string, orig map[string]string) {
	orig = make(map[string]string)

	for _, e := range emails {
		if !IsEmail(e) {
			continue
		}

		// 移除地址扩展
		addr := StripExtension(e)
		e = ToLowerWithExt(e)

		if strings.HasSuffix(addr, "@"+domain) {
			if !slices.Contains(same, addr) {
				same = append(same, addr)
				orig[addr] = e
			}
		} else {
			others = append(others, addr)
			orig[addr] = e
		}
	}

	return
}

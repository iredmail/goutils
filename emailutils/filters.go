package emailutils

import (
	"slices"
	"strings"
)

// FilterValidEmails 从给定的邮件地址列表里过滤出有效的和无效的邮件地址。
// 注意：邮件地址扩展会被保留。
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
//
// - same 和 others 里的邮件地址都已转换为小写并移除地址扩展（+ext），但在 `orig` 的 Value 里保留了地址扩展。
// - orig 以移除地址扩展后的地址作为 key，以原始邮件地址作为 value。
func FilterSameDomainEmails(domain string, mails []string) (same []string, others []string, orig map[string]string) {
	orig = make(map[string]string)

	for _, e := range mails {
		if !IsEmail(e) {
			continue
		}

		// 转换为小写并移除地址扩展
		addr := StripExtension(e)

		// 转换为小写但保留地址扩展
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

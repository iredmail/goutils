package emailutils

import (
	"slices"
	"strings"
)

// FilterValidEmails 从给定的邮件地址列表里过滤出有效的和无效的邮件地址。
// 注意：邮件地址扩展会被保留。
func FilterValidEmails(addrs []string) (valid []string, invalid []string) {
	for _, addr := range addrs {
		addr = strings.TrimSpace(addr)

		if IsEmail(addr) {
			valid = append(valid, ToLowerWithExt(addr))
		} else {
			invalid = append(invalid, ToLowerWithExt(addr))
		}
	}

	return
}

func FilterValidDomains(domains []string) (valid []string, invalid []string) {
	for _, d := range domains {
		d = strings.ToLower(strings.TrimSpace(d))

		if IsDomain(d) {
			valid = append(valid, d)
		} else {
			invalid = append(invalid, d)
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

	if domain == "" || len(mails) == 0 {
		return
	}

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

// FilterEmailsInDomains 过滤出属于指定域名列表的邮件地址和不属于指定域名列表的邮件地址。
//
// - 无效的邮件地址和域名都会被忽略。
// - 域名比较不区分大小写。
// - 返回的邮件地址去除了首尾空白、转换为小写，但保留地址扩展。重复地址只保留一个。
func FilterEmailsInDomains(mails, domains []string, removeExt bool) (in, notIn []string) {
	mDomains := make(map[string]bool)
	for _, d := range domains {
		d = strings.ToLower(strings.TrimSpace(d))
		if d != "" {
			mDomains[d] = true
		}
	}

	for _, orig := range mails {
		orig = strings.TrimSpace(orig)

		if !IsEmail(orig) {
			continue
		}

		var e string
		if removeExt {
			e = StripExtension(orig)
		} else {
			e = ToLowerWithExt(orig)
		}

		if mDomains[ExtractDomainFromEmail(e)] {
			if !slices.Contains(in, e) {
				in = append(in, e)
			}
		} else {
			if !slices.Contains(notIn, e) {
				notIn = append(notIn, e)
			}
		}
	}

	return
}

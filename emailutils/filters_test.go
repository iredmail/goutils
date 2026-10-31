package emailutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilterEmailsInDomains(t *testing.T) {
	// domains 包含重复项、大小写不一致、前后空格以及空串。
	domains := []string{"a.io", "A.IO", " b.com ", "B.Com", "a.io", "", "  "}

	tests := []struct {
		name      string
		mails     []string
		domains   []string
		removeExt bool
		in        []string
		notIn     []string
	}{
		{
			name:    "基本过滤：大小写、空格、重复",
			domains: domains,
			mails: []string{
				"user1@a.io",
				"User2@A.IO",    // 大写域名
				"USER1@a.io",    // 与 user1@a.io 仅大小写不同，视为重复
				"user1@a.io",    // 完全重复
				" user3@b.com ", // 前后空格
				"other@c.net",   // 不在 domains 内
				"OTHER@C.NET",   // 与 other@c.net 仅大小写不同，视为重复
			},
			in:    []string{"user1@a.io", "user2@a.io", "user3@b.com"},
			notIn: []string{"other@c.net"},
		},
		{
			name:      "保留地址扩展（removeExt=false）",
			domains:   domains,
			removeExt: false,
			mails: []string{
				"user+Ext@a.io",
				"user+ext@a.io",   // 扩展大小写不同，视为不同地址
				"user@a.io",       // 无扩展，与带扩展的地址不同
				"Other+Tag@D.ORG", // 不在 domains 内，带扩展
			},
			in:    []string{"user+Ext@a.io", "user+ext@a.io", "user@a.io"},
			notIn: []string{"other+Tag@d.org"},
		},
		{
			name:      "移除地址扩展（removeExt=true）",
			domains:   domains,
			removeExt: true,
			mails: []string{
				"user+Ext@a.io",
				"user+ext@a.io",   // 移除扩展后与上一条重复
				"USER@A.IO",       // 移除扩展后与上一条重复
				"user+a+b@B.COM",  // 多个 + 号
				"Other+Tag@D.ORG", // 不在 domains 内
				"other@d.org",     // 移除扩展后与上一条重复
			},
			in:    []string{"user@a.io", "user@b.com"},
			notIn: []string{"other@d.org"},
		},
		{
			name:    "子域名不视为匹配",
			domains: []string{"a.io"},
			mails:   []string{"u@sub.a.io", "u@a.io", "u@xa.io"},
			in:      []string{"u@a.io"},
			notIn:   []string{"u@sub.a.io", "u@xa.io"},
		},
		{
			name:    "IP 字面量域名",
			domains: []string{"192.168.1.1"},
			mails:   []string{"u@[192.168.1.1]", "u@[10.0.0.1]"},
			in:      []string{"u@[192.168.1.1]"},
			notIn:   []string{"u@[10.0.0.1]"},
		},
		{
			name:    "无效邮件地址被忽略",
			domains: domains,
			mails:   []string{"invalid", "bad@domain", "@a.io", "user@", "", "  ", "u@a.io."},
			in:      nil,
			notIn:   nil,
		},
		{
			name:    "空 domains：所有有效地址均为 notIn",
			domains: nil,
			mails:   []string{"u@a.io", "x", "v@b.com"},
			in:      nil,
			notIn:   []string{"u@a.io", "v@b.com"},
		},
		{
			name:    "空 mails",
			domains: domains,
			mails:   nil,
			in:      nil,
			notIn:   nil,
		},
		{
			name:    "mails 与 domains 均为空",
			domains: nil,
			mails:   nil,
			in:      nil,
			notIn:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in, notIn := FilterEmailsInDomains(tt.mails, tt.domains, tt.removeExt)
			assert.Equal(t, tt.in, in, "in")
			assert.Equal(t, tt.notIn, notIn, "notIn")
		})
	}
}

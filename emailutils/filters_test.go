package emailutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilterEmailsInDomains(t *testing.T) {
	// domains 包含重复项、大小写不一致以及前后空格。
	domains := []string{"a.io", "A.IO", " b.com ", "B.Com", "a.io", ""}
	mails := []string{
		"user1@a.io",
		"User2@A.IO",       // 大写域名
		"user3+Ext@b.com",  // 带扩展
		"user1@a.io",       // 重复
		"USER4@B.COM",      // 大写
		"other@c.net",      // 不在 domains 内
		"Other2+Tag@D.ORG", // 不在 domains 内，带扩展
		"other@c.net",      // 重复
		"invalid",          // 无效地址
		"bad@domain",       // 无效地址
		" user5@a.io ",     // 前后空格
	}

	valid, invalid := FilterEmailsInDomains(mails, domains)
	assert.Equal(t, []string{"user1@a.io", "User2@A.IO", "user3+Ext@b.com", "USER4@B.COM", "user5@a.io"}, valid)
	assert.Equal(t, []string{"other@c.net", "Other2+Tag@D.ORG"}, invalid)

	// 空 domains：所有有效地址均为 invalid。
	valid, invalid = FilterEmailsInDomains([]string{"u@a.io", "x"}, nil)
	assert.Empty(t, valid)
	assert.Equal(t, []string{"u@a.io"}, invalid)

	// 空 mails。
	valid, invalid = FilterEmailsInDomains(nil, domains)
	assert.Empty(t, valid)
	assert.Empty(t, invalid)
}

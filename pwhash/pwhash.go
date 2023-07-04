package pwhash

import (
	"strings"

	"golang.org/x/exp/slices"

	"github.com/iredmail/goutils/respcode"
)

const (
	SchemePlain       = "PLAIN"
	SchemeCrypt       = "CRYPT"
	SchemeMD5         = "MD5"
	SchemePlainMD5    = "PLAIN-MD5"
	SchemeSHA         = "SHA"
	SchemeSSHA        = "SSHA"
	SchemeSHA512      = "SHA512"
	SchemeSSHA512     = "SSHA512"
	SchemeSHA512Crypt = "SHA512-CRYPT"
	SchemeBcrypt      = "BCRYPT"
	SchemeCramMD5     = "CRAM-MD5"
	SchemeNTLM        = "NTLM"
)

var (
	SupportedPasswordSchemes = []string{
		SchemePlain,
		SchemeCrypt,
		SchemeMD5,
		SchemePlainMD5,
		SchemeSHA,
		SchemeSSHA,
		SchemeSHA512,
		SchemeSSHA512,
		SchemeSHA512Crypt,
		SchemeBcrypt,
		SchemeCramMD5,
		SchemeNTLM,
	}
)

// ExtractSchemeFromPasswordHash 从密码哈希中提取哈希算法名称。例如：`{SSHA}xxx` -> `SSHA`。
// 注意：返回的 schema 名称是大写的。
func ExtractSchemeFromPasswordHash(pwHash string) (scheme string) {
	_, after, found := strings.Cut(pwHash, "{")
	if !found {
		// no password scheme name.
		return
	}

	scheme, _, found = strings.Cut(after, "}")
	if !found {
		// no password scheme name.
		return
	}

	return strings.ToUpper(scheme)
}

func IsSupportedPasswordScheme(scheme string) bool {
	return slices.Contains(SupportedPasswordSchemes, scheme)
}

func GeneratePassword(scheme string, plainPassword string) (hash string, err error) {
	if len(plainPassword) == 0 {
		err = respcode.ErrEmptyPassword

		return
	}

	scheme = strings.ToUpper(scheme)

	if !slices.Contains(SupportedPasswordSchemes, scheme) {
		err = respcode.ErrUnsupportedPasswordScheme

		return
	}

	switch strings.ToUpper(scheme) {
	case SchemePlain:
		hash = plainPassword
	case SchemeCrypt:
		// TODO
	case SchemeMD5:
		hash = GenerateMD5Password(plainPassword)
	case SchemePlainMD5:
		hash, err = GeneratePlainMD5Password(plainPassword)
	case SchemeSHA:
		// TODO
	case SchemeSSHA:
		hash, err = GenerateSSHAPassword(plainPassword)
	case SchemeSHA512:
		hash, err = GenerateSHA512Password(plainPassword)
	case SchemeSSHA512:
		hash, err = GenerateSSHA512Password(plainPassword)
	case SchemeSHA512Crypt:
		// TODO
	case SchemeBcrypt:
		hash, err = GenerateBcryptPassword(plainPassword)
	case SchemeCramMD5:
		// TODO
	case SchemeNTLM:
		// TODO
	}

	return
}

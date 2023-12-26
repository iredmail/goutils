package pwhash

import (
	"slices"
	"strings"

	"github.com/iredmail/goutils/respcode"
)

// TODO 支持 argon2

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
	// SchemeCramMD5     = "CRAM-MD5"
	// SchemeNTLM        = "NTLM"

	// SchemeBcrypt Blowfish crypt.
	// bcrypt is not available in libc `crypt()` on old Linux distributions.
	// Since v2.3.0 this is provided by dovecot.
	SchemeBcrypt  = "BLF-CRYPT"
	SchemeBcrypt2 = "CRYPT"
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
		// SchemeCramMD5,
		// SchemeNTLM,
	}
)

// extractSchemeAndHash 从密码哈希中提取哈希算法名称及哈希字符串。例如：`{ssha}xxx` -> `SSHA`, `xxx`。
// 注意：返回的 schema 名称是大写的。
func extractSchemeAndHash(s string) (scheme, hash string) {
	_, after, found := strings.Cut(s, "{")
	if !found {
		// no password scheme name.
		return
	}

	scheme, hash, found = strings.Cut(after, "}")
	if !found {
		// no password scheme name.
		return
	}

	return strings.ToUpper(scheme), hash
}

// GeneratePassword 加密密码。注意：带有哈希算法前缀，如 `{SSHA512}`。
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

	switch scheme {
	case SchemePlain:
		hash = "{PLAIN}" + plainPassword
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

		// case SchemeCramMD5:
		// TODO
		// case SchemeNTLM:
		// TODO
	}

	return
}

func VerifyPassword(hashedPassword, plainPassword string) (matched bool, err error) {
	if len(hashedPassword) == 0 || len(plainPassword) == 0 {
		err = respcode.ErrEmptyPassword

		return
	}

	// 明文密码不带 scheme。
	if hashedPassword == plainPassword {
		return true, nil
	}

	scheme, _ := extractSchemeAndHash(hashedPassword)

	if !slices.Contains(SupportedPasswordSchemes, scheme) {
		err = respcode.ErrUnsupportedPasswordScheme

		return
	}

	switch scheme {
	case SchemePlain:
		if hashedPassword == "{PLAIN}"+plainPassword || hashedPassword == "{plain}"+plainPassword {
			matched = true
		}
	case SchemeCrypt:
		// TODO
	case SchemeMD5:
		matched = VerifyMD5Password(hashedPassword, plainPassword)
	case SchemePlainMD5:
		matched = VerifyPlainMD5Password(hashedPassword, plainPassword)
	case SchemeSHA:
		// TODO
	case SchemeSSHA:
		matched = VerifySSHAPassword(hashedPassword, plainPassword)
	case SchemeSHA512:
		matched = VerifySHA512Password(hashedPassword, plainPassword)
	case SchemeSSHA512:
		matched = VerifySSHA512Password(hashedPassword, plainPassword)
	case SchemeSHA512Crypt:
		// TODO
	case SchemeBcrypt:
		matched = VerifyBcryptPassword(hashedPassword, plainPassword)
		// case SchemeCramMD5:
		// TODO
		// case SchemeNTLM:
		// TODO
	}

	return
}

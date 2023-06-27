package pwhash

import (
	"strings"
)

const (
	SchemePlain       = "PLAIN"
	SchemeCrypt       = "CRYPT"
	SchemeMD5         = "MD5"
	SchemePlainMD5    = "PLAIN-MD5"
	SchemeSHA         = "SHA"
	SchemeSHA512      = "SHA512"
	SchemeSSHA512     = "SSHA512"
	SchemeSHA512Crypt = "SHA512-CRYPT"
	SchemeBcrypt      = "BCRYPT"
	SchemeCramMD5     = "CRAM-MD5"
	SchemeNTLM        = "NTLM"
)

type Scheme string

func IsSupportedPasswordScheme(pwHash string) bool {
	if !(strings.HasPrefix(pwHash, "{") && strings.Contains(pwHash, "}")) {
		return false
	}

	// Extract scheme name from password hash: "{SSHA}xxxx" -> "SSHA"
	scheme := strings.Split(strings.Split(pwHash, "}")[0], "{")[1]
	scheme = strings.ToUpper(scheme)

	supportedSchemes := []string{
		SchemePlain,
		SchemeCrypt,
		SchemeMD5,
		SchemePlainMD5,
		SchemeSHA,
		SchemeSHA512,
		SchemeSSHA512,
		SchemeSHA512Crypt,
		SchemeBcrypt,
		SchemeCramMD5,
		SchemeNTLM,
	}

	for _, supportedScheme := range supportedSchemes {
		if scheme == supportedScheme {
			return true
		}
	}

	return false
}

func GeneratePasswordHash(scheme Scheme, password string) (challenge string, err error) {
	// TODO

	switch scheme {
	case SchemePlain:

	}

	return
}

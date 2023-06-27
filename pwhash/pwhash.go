package pwhash

import (
	"strings"
)

const (
	SchemaPlain       = "PLAIN"
	SchemaCrypt       = "CRYPT"
	SchemaMD5         = "MD5"
	SchemaPlainMD5    = "PLAIN-MD5"
	SchemaSHA         = "SHA"
	SchemaSHA512      = "SHA512"
	SchemaSSHA512     = "SSHA512"
	SchemaSHA512Crypt = "SHA512-CRYPT"
	SchemaBcrypt      = "BCRYPT"
	SchemaCramMD5     = "CRAM-MD5"
	SchemaNTLM        = "NTLM"
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
		SchemaPlain,
		SchemaCrypt,
		SchemaMD5,
		SchemaPlainMD5,
		SchemaSHA,
		SchemaSHA512,
		SchemaSSHA512,
		SchemaSHA512Crypt,
		SchemaBcrypt,
		SchemaCramMD5,
		SchemaNTLM,
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
	case SchemaPlain:

	}

	return
}

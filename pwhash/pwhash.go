package pwhash

import (
	"strings"
)

func IsSupportedPasswordScheme(pwHash string) bool {
	if !(strings.HasPrefix(pwHash, "{") && strings.Contains(pwHash, "}")) {
		return false
	}

	// Extract scheme name from password hash: "{SSHA}xxxx" -> "SSHA"
	scheme := strings.Split(strings.Split(pwHash, "}")[0], "{")[1]
	scheme = strings.ToUpper(scheme)

	supportedSchemes := []string{
		"PLAIN",
		"CRYPT",
		"MD5",
		"PLAIN-MD5",
		"SHA",
		"SSHA",
		"SHA512",
		"SSHA512",
		"SHA512-CRYPT",
		"BCRYPT",
		"CRAM-MD5",
		"NTLM",
	}

	for _, supportedScheme := range supportedSchemes {
		if scheme == supportedScheme {
			return true
		}
	}

	return false
}

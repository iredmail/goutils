package pwhash

import (
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// Blowfish crypt (bcrypt) scheme.
// It is generally considered to be very secure.
// The encrypted password will start with $2y$ (other generators can generate
// passwords that have other letters after $2, those should work too.)
//
// FYI
//
//	- https://en.wikipedia.org/wiki/Bcrypt
//	- https://doc.dovecot.org/configuration_manual/authentication/password_schemes/

func IsBcryptHash(s string) (ok bool, hash string) {
	scheme, hash := extractSchemeAndHash(s)

	switch scheme {
	case "BLF-CRYPT", "CRYPT":
		// $2a$
		// $2x$, $2y$ (June 2011)
		// $2b$ (February 2014)
		if strings.HasPrefix(hash, `$2a$`) ||
			strings.HasPrefix(hash, `$2b$`) ||
			strings.HasPrefix(hash, `$2x$`) ||
			strings.HasPrefix(hash, `$2y$`) {
			ok = true
		}
	}

	return
}

func GenerateBcryptPassword(plain string) (hash string, err error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return
	}

	hash = "{BLF-CRYPT}" + string(hashed)

	return
}

func VerifyBcryptPassword(challengePassword, plainPassword string) (matched bool) {
	_, hash := extractSchemeAndHash(challengePassword)

	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plainPassword)) == nil
}

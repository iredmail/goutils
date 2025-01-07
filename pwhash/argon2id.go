package pwhash

import (
	"strings"

	"github.com/alexedwards/argon2id"
)

func GenArgon2IDPassword(plain string, prefixScheme bool) (hash string, err error) {
	hash, err = argon2id.CreateHash(plain, argon2id.DefaultParams)
	if err != nil {
		return
	}

	if prefixScheme {
		hash = "{ARGON2ID}" + hash
	}

	return
}

func VerifyArgon2IDPassword(plain, hash string) (matched bool, err error) {
	if strings.HasPrefix(hash, "{ARGON2ID}") || strings.HasPrefix(hash, "{argon2id}") {
		hash = hash[10:]
	}

	return argon2id.ComparePasswordAndHash(plain, hash)
}

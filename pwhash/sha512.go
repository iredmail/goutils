package pwhash

import (
	"bytes"
	"crypto/sha512"
	"encoding/base64"
	"strings"
)

func GenerateSHA512Password(password string) (challenge string, err error) {
	sha := sha512.New()
	if _, err = sha.Write([]byte(password)); err != nil {
		return
	}

	calculatedHash := sha.Sum(nil)
	challenge = "{SHA512}" + base64.StdEncoding.EncodeToString(calculatedHash)

	return
}

func VerifySHA512Password(challengePassword, plainPassword string) bool {
	if strings.HasPrefix(challengePassword, "{SHA512}") ||
		strings.HasPrefix(challengePassword, "{sha512}") {
		challengePassword = challengePassword[8:]
	}

	hash, err := base64.StdEncoding.DecodeString(challengePassword)
	if err != nil {
		return false
	}

	password := []byte(plainPassword)
	sha := sha512.New()
	sha.Write(password)
	calculatedHash := sha.Sum(nil)

	return bytes.Equal(hash, calculatedHash)
}

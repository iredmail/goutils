package pwhash

import (
	"bytes"
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"strings"
)

func GenerateSSHA512Password(password string) (challenge string, err error) {
	salt := make([]byte, 8)
	if _, err = rand.Read(salt); err != nil {
		return
	}

	sha := sha512.New()
	sha.Write([]byte(password))
	sha.Write(salt)
	hash := sha.Sum(nil)

	hashWithSalt := append(hash, salt...)
	encodedHash := base64.StdEncoding.EncodeToString(hashWithSalt)

	return "{SSHA512}" + encodedHash, nil
}

func VerifySSHA512Password(challengePassword, plainPassword string) bool {
	if strings.HasPrefix(challengePassword, "{SSHA512}") ||
		strings.HasPrefix(challengePassword, "{ssha512}") {
		challengePassword = challengePassword[9:]
	}

	hashWithSalt, err := base64.StdEncoding.DecodeString(challengePassword)
	if err != nil {
		return false
	}

	hash := hashWithSalt[:64]
	salt := hashWithSalt[64:]

	sha := sha512.New()
	sha.Write([]byte(plainPassword))
	sha.Write(salt)
	calculatedHash := sha.Sum(nil)

	return bytes.Equal(hash, calculatedHash)
}

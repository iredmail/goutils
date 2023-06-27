package pwhash

import (
	"bytes"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"strings"
)

func GenerateSSHAPassword(password string) (hashStr string, err error) {
	passwordBytes := []byte(password)
	salt := make([]byte, 8)
	if _, err = rand.Read(salt); err != nil {
		panic(err)
	}

	hash := sha1.New()
	hash.Write(passwordBytes)
	hash.Write(salt)
	hashedPassword := hash.Sum(nil)

	hashSSHA := append(hashedPassword, salt...)
	hashStr = "{SSHA}" + base64.StdEncoding.EncodeToString(hashSSHA)

	return
}

func VerifySSHAPassword(challengePassword, plainPassword string) bool {
	index := strings.LastIndex(challengePassword, "}")
	if index < 0 {
		return false
	}

	if strings.Index(challengePassword, "{SSHA}") == 0 ||
		strings.Index(challengePassword, "{ssha}") == 0 {
		challengePassword = challengePassword[6:]
	} else if strings.Index(challengePassword, "{SHA}") == 0 ||
		strings.Index(challengePassword, "{sha}") == 0 {
		challengePassword = challengePassword[5:]
	}

	challengeBytes, err := base64.StdEncoding.DecodeString(challengePassword)
	if err != nil {
		return false
	}

	digest := challengeBytes[:20]
	salt := challengeBytes[20:]
	hash := sha1.New()
	hash.Write([]byte(plainPassword))
	hash.Write(salt)
	hashedPassword := hash.Sum(nil)

	return bytes.Equal(digest, hashedPassword)
}

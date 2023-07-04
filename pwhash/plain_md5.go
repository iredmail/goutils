package pwhash

import (
	"crypto/md5"
	"encoding/hex"
	"strings"
)

func GeneratePlainMD5Password(password string) (challenge string, err error) {
	hash := md5.New()
	if _, err = hash.Write([]byte(password)); err != nil {
		return
	}

	hashBytes := hash.Sum(nil)

	return hex.EncodeToString(hashBytes), nil
}

func VerifyPlainMD5Password(challengePassword, plainPassword string) bool {
	if strings.HasPrefix(challengePassword, "{PLAIN-MD5}") ||
		strings.HasPrefix(challengePassword, "{plain-md5}") {
		challengePassword = challengePassword[11:]
	}

	challenge, err := GeneratePlainMD5Password(plainPassword)
	if err != nil {
		return false
	}

	return strings.EqualFold(challenge, challengePassword)
}

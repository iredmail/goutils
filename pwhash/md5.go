package pwhash

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strings"
)

func GenerateMD5Password(password string) string {
	md5Bytes := md5.Sum([]byte(password))

	return fmt.Sprintf("{MD5}%x", md5Bytes)
}

func VerifyMD5Password(challengePassword, plainPassword string) bool {
	return strings.EqualFold(GenerateMD5Password(plainPassword), challengePassword)
}

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

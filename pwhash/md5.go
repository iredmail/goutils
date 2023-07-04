package pwhash

import (
	"crypto/md5"
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

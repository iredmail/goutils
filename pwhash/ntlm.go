package pwhash

import (
	"encoding/hex"
	"strings"
	"unicode/utf16"

	"golang.org/x/crypto/md4"
)

func GenerateNTLMPassword(password string) (challenge string, err error) {
	hash := md4.New()
	if _, err = hash.Write([]byte(password)); err != nil {
		return
	}

	ntlmHash := hash.Sum(nil)

	return hex.EncodeToString(ntlmHash), nil
}

func VerifyNTLMPassword(challengePassword, plainPassword string) bool {
	// Convert password to UTF-16 little-endian byte array
	passwordUTF16 := utf16.Encode([]rune(plainPassword))
	passwordBytes := make([]byte, len(passwordUTF16)*2)
	for i, v := range passwordUTF16 {
		passwordBytes[i*2] = byte(v)
		passwordBytes[i*2+1] = byte(v >> 8)
	}

	// Convert NTLM hash to byte array
	ntlmHash := strings.ToUpper(challengePassword)
	ntlmHashBytes, err := hex.DecodeString(ntlmHash)
	if err != nil {
		return false
	}

	// Compare password and NTLM hash
	if len(passwordBytes) != len(ntlmHashBytes) {
		return false
	}

	equal := true
	for i := 0; i < len(passwordBytes); i++ {
		equal = equal && (passwordBytes[i] == ntlmHashBytes[i])
	}

	return equal
}

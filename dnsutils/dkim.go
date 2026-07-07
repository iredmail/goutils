package dnsutils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"
)

const (
	DefaultDKIMSelector   = "dkim"
	DefaultDKIMKeyLength  = 2048
	dkimSelectorMaxLength = 20
)

// ValidateDKIMSelector 检查 DKIM selector 字符串。
// 仅允许最多20个字符，且只能包含小写字母、数字和短横线。
func ValidateDKIMSelector(s string) error {
	s = strings.ToLower(strings.TrimSpace(s))

	if s == "" {
		return errors.New("selector cannot be empty")
	}

	if len(s) > dkimSelectorMaxLength {
		return fmt.Errorf("selector too long (max %d characters)", dkimSelectorMaxLength)
	}

	for _, ch := range s {
		if (ch >= 'a' && ch <= 'z') ||
			(ch >= '0' && ch <= '9') ||
			ch == '-' {
			continue
		}

		return fmt.Errorf("selector contains invalid characters, only lowercase letters, digits and dashes are allowed")
	}

	return nil
}

func GenDKIMKey(keyLength int) (privateKey, publicKey string, err error) {
	pk, err := rsa.GenerateKey(rand.Reader, keyLength)
	if err != nil {
		return
	}

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(pk.Public())
	if err != nil {
		return
	}

	// 注意：目前 DKIM key 只适合用 PKCS1 格式
	privateKeyBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(pk),
	})

	publicKey = base64.StdEncoding.EncodeToString(publicKeyBytes)
	privateKey = string(privateKeyBytes)

	return
}

package enc

import (
	"github.com/Luzifer/go-openssl/v4"
)

func Encrypt(secretKey string, data []byte) ([]byte, error) {
	return openssl.New().EncryptBytes(secretKey, data, openssl.PBKDF2SHA256)
}

func Decrypt(secretKey string, data []byte) ([]byte, error) {
	return openssl.New().DecryptBytes(secretKey, data, openssl.PBKDF2SHA256)
}

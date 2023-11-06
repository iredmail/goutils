package enc

import (
	"github.com/Luzifer/go-openssl/v4"
)

// Encrypt 加密 `[]byte`, 加密成功返回 base64 encode 后的 `[]byte`。
func Encrypt(secretKey string, data []byte) ([]byte, error) {
	return openssl.New().EncryptBytes(secretKey, data, openssl.PBKDF2SHA256)
}

// EncryptBinaryBytes 加密 `[]byte`, 加密成功返回 `[]byte`。
func EncryptBinaryBytes(secretKey string, data []byte) ([]byte, error) {
	return openssl.New().EncryptBinaryBytes(secretKey, data, openssl.PBKDF2SHA256)
}

// Decrypt 解密 base64 encode 之后的 `[]byte`, 解密成功返回 `[]byte`。
func Decrypt(secretKey string, data []byte) ([]byte, error) {
	return openssl.New().DecryptBytes(secretKey, data, openssl.PBKDF2SHA256)
}

// DecryptBinaryBytes 解密 `[]byte`, 解密成功返回 `[]byte`。
func DecryptBinaryBytes(secretKey string, data []byte) ([]byte, error) {
	return openssl.New().DecryptBinaryBytes(secretKey, data, openssl.PBKDF2SHA256)
}

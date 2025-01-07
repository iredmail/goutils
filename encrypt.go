package goutils

import (
	"github.com/Luzifer/go-openssl/v4"
)

// EncryptWithSecret 使用 openssl 加密数据。
// 注意：`openssl.New().EncryptBytes()` 方法会自动将数据 base64 encode。
func EncryptWithSecret(secret string, d []byte) (encrypted []byte, err error) {
	return openssl.New().EncryptBytes(secret, d, openssl.PBKDF2SHA256)
}

// DecryptWithSecret 使用 openssl 解密数据。
// 注意：数据使用 `openssl.New().EncryptBytes()` 处理，会自动将解密后的数据做 base64 decode 处理。
func DecryptWithSecret(secret string, d []byte) (decrypted []byte, err error) {
	return openssl.New().DecryptBytes(secret, d, openssl.PBKDF2SHA256)
}

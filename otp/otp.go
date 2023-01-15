package otp

import (
	"encoding/base32"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/dgryski/dgoogauth"
	"rsc.io/qr"
)

var ErrInvalidOtpSecret = errors.New("invalid otp secret")

// Gen2FAImageData 生成 2FA 扫描二维码的内容
// - HMAC-SHA-1 加密后的长度得到一个20字节的密串
// - 取这个20字节的密串的最后一个字节，取这个字节的低4位，作为截取加密串的下标偏移量
// - 按照下标偏移量开始，获取4个字节，按照大端方式组成一个整数
// - 截取这个整数的后6位或者后8位转成字符串返回
func Gen2FAImageData(productName, account, secret string) (encode string, err error) {
	length := len(secret)
	if length < 4 {
		err = ErrInvalidOtpSecret

		return
	}

	prefix := secret[length-4:]
	last4 := secret[:length-4]

	_secret := []byte(last4 + hex.EncodeToString([]byte(prefix)))
	b32Secret := base32.StdEncoding.EncodeToString(_secret)
	_url := fmt.Sprintf(
		"otpauth://totp/%s:%s?issuer=%s&secret=%s",
		productName,
		account,
		productName,
		b32Secret,
	)

	code, err := qr.Encode(_url, qr.Q)
	if err != nil {
		return
	}
	encode = base64.StdEncoding.EncodeToString(code.PNG())

	return
}

// Authenticate 执行 OTP 验证。
func Authenticate(secret, password string) (authed bool) {
	length := len(secret)
	if length < 4 {
		return false
	}

	last4 := secret[:len(secret)-4]
	prefix := secret[length-4:]

	_secret := []byte(last4 + hex.EncodeToString([]byte(prefix)))
	b32Secret := base32.StdEncoding.EncodeToString(_secret)

	c := &dgoogauth.OTPConfig{
		Secret:     b32Secret,
		WindowSize: 3,
		UTC:        true,
	}

	authed, err := c.Authenticate(password)
	if err != nil {
		return false
	}

	return authed
}

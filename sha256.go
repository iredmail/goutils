package goutils

import (
	"bytes"
	"crypto/sha256"
	"io"
	"regexp"
)

// 匹配 sha256 checksum 的正则表达式。
// - SHA-256 固定 64 个十六进制字符
// - 只接受 0-9、a-f、A-F，拒绝 g-z 和 Base64 字符（+、/）
// 必须全量匹配 `^...$`，否则 len(s) > 64 时 MatchString 会在前缀命中（正则的锚点问题很常见，多写一层 len(s) == 64 做双保险也行）
var regxSha256Checksum = regexp.MustCompile(`^[a-fA-F0-9]{64}$`)

func GenSha256(content []byte) (shasum []byte, err error) {
	h := sha256.New()
	if _, err = io.Copy(h, bytes.NewReader(content)); err != nil {
		return
	}

	return h.Sum(nil), nil
}

func GenSha256FromReader(r io.Reader) (shasum []byte, err error) {
	h := sha256.New()
	if _, err = io.Copy(h, r); err != nil {
		return
	}

	return h.Sum(nil), nil
}

func IsSameSha256(a, b []byte) bool {
	return bytes.Equal(a, b)
}

func IsValidSHA256Checksum(s string) bool {
	return regxSha256Checksum.MatchString(s)
}

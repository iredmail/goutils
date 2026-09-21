package goutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidSHA256Checksum(t *testing.T) {
	// 合法的 SHA-256 校验值：全小写十六进制。
	assert.Equal(t, true, IsValidSHA256Checksum("e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"))
	// 合法的 SHA-256 校验值：全大写十六进制。
	assert.Equal(t, true, IsValidSHA256Checksum("E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855"))
	// 长度不足 64 位，应该判定为无效。
	assert.Equal(t, false, IsValidSHA256Checksum("e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b85"))
	// 长度超过 64 位，应该判定为无效。
	assert.Equal(t, false, IsValidSHA256Checksum("e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855aa"))
	// 出现非法字符 g，应该判定为无效。
	assert.Equal(t, false, IsValidSHA256Checksum("e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b85g"))
	// 空字符串不可能是合法的 SHA-256 校验值。
	assert.Equal(t, false, IsValidSHA256Checksum(""))
	// 前后带空格时，不应被当作合法值。
	assert.Equal(t, false, IsValidSHA256Checksum(" e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855 "))
}

package goutils

import (
	"crypto/rand"
	"math/big"
	mRand "math/rand"
	"time"
)

const (
	charsForRandomString = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

// genWeakerRandomString 使用 `math/rand` 包生成指定长度的随机字符串。
// `math/rand` 被认为比 `crypt/rand` 弱。
func genWeakerRandomString(length int) string {
	mRand.Seed(time.Now().UnixNano())

	s := []rune(charsForRandomString)
	b := make([]rune, length)

	for i := range b {
		b[i] = s[mRand.Intn(len(s))]
	}

	return string(b)
}

func GenRandomString(length int) string {
	ret := make([]byte, length)
	charLen := int64(len(charsForRandomString))
	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(charLen))

		if err != nil {
			return genWeakerRandomString(length)
		}
		ret[i] = charsForRandomString[num.Int64()]
	}

	return string(ret)
}

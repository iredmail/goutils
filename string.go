package goutils

import (
	"crypto/rand"
	"math/big"
	mRand "math/rand"
	"reflect"
	"slices"
	"strings"
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
	for i := range length {
		num, err := rand.Int(rand.Reader, big.NewInt(charLen))

		if err != nil {
			return genWeakerRandomString(length)
		}
		ret[i] = charsForRandomString[num.Int64()]
	}

	return string(ret)
}

func SplitLines(s string) (lines []string) {
	newLines := strings.Split(s, "\n")

	for _, i := range newLines {
		lines = append(lines, strings.TrimSpace(i))
	}

	return
}

// StringSliceToLower 将 slice 里的元素都转换为小写。
func StringSliceToLower(ss []string) {
	for i := range len(ss) {
		ss[i] = strings.ToLower(ss[i])
	}
}

func Flatten(v any) (flattened []string) {
	if v == nil {
		return
	}

	var results []string
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.String:
		results = append(results, rv.String())
	case reflect.Slice:
		for i := 0; i < rv.Len(); i++ {
			results = append(results, Flatten(rv.Index(i).Interface())...)
		}
	default:
		break
	}

	// Remove empty and duplicate value.
	for _, result := range results {
		if result != "" && !slices.Contains(flattened, result) {
			flattened = append(flattened, result)
		}
	}

	return flattened
}

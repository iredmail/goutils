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

// SplitLines 按行拆分字符串，并去除每行的首尾空白字符。
func SplitLines(s string) (lines []string) {
	// Handle both Unix-style (\n) and Windows-style (\r\n) newlines
	// by first replacing Windows-style newlines with Unix-style ones.
	s = strings.ReplaceAll(s, "\r\n", "\n")

	for _, i := range strings.Split(s, "\n") {
		lines = append(lines, strings.TrimSpace(i))
	}

	return
}

// RemoveSpacesAndNewLines 去除字符串中的多余的空白字符和换行符，并使用指定的分隔符连接剩余的字段。
func RemoveSpacesAndNewLines(s string, sep ...string) string {
	fields := strings.Fields(s)

	if len(sep) > 0 {
		return strings.Join(fields, sep[0])
	} else {
		return strings.Join(fields, "")
	}
}

// StringSliceToLower 将 slice 里的元素都转换为小写。
func StringSliceToLower(ss []string) {
	for i := range len(ss) {
		ss[i] = strings.ToLower(ss[i])
	}
}

// FlattenStrings 函数将输入的任意类型 v 展平为一个字符串切片，空字符串和重复的字符串会被移除。
// 对于嵌套的切片，函数会递归地展平所有层级。
func FlattenStrings(v any) (flattened []string) {
	if v == nil {
		return
	}

	var results []string
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.String:
		results = append(results, rv.String())
	case reflect.Slice:
		for i := range rv.Len() {
			results = append(results, FlattenStrings(rv.Index(i).Interface())...)
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

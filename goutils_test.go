package goutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGoUtils(t *testing.T) {
	var arrInt []int
	var arrStr []string
	var ptr *string
	var m map[string]string
	mm := make(map[string]string)
	assert.True(t, IsEmpty(""))
	assert.True(t, IsEmpty(arrInt))
	assert.True(t, IsEmpty(arrStr))
	assert.True(t, IsEmpty(ptr))
	assert.True(t, IsEmpty(m))
	assert.True(t, IsEmpty(mm))
	arrInt = []int{1, 2, 3}
	arrStr = []string{"a", "b", "c"}
	str := "ptr"
	ptr = &str
	mm["a"] = "A"
	assert.True(t, NotEmpty("str"))
	assert.True(t, NotEmpty(arrInt))
	assert.True(t, NotEmpty(arrStr))
	assert.True(t, NotEmpty(ptr))
	assert.True(t, NotEmpty(mm))

	assert.True(t, IsIPv4("192.168.2.4"))
	assert.False(t, IsIPv4("2001:0db8:85a3:0000:0000:8a2e:0370:7334"))

	assert.True(t, IsUUID("DA14F3A5-AAEB-44C1-8A14-5146FA60B7DD"))

	assert.True(t, CalculateTotalPages(10, 50) == 1)
	assert.True(t, CalculateTotalPages(50, 50) == 1)
	assert.True(t, CalculateTotalPages(52, 50) == 2)
	assert.True(t, CalculateTotalPages(200, 50) == 4)
}

package goutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGoUtils(t *testing.T) {
	var arrInt []int
	var arrStr []string
	var ptr *string
	assert.True(t, IsEmpty(""))
	assert.True(t, IsEmpty(arrInt))
	assert.True(t, IsEmpty(arrStr))
	assert.True(t, IsEmpty(ptr))
	arrInt = []int{1, 2, 3}
	arrStr = []string{"a", "b", "c"}
	str := "ptr"
	ptr = &str
	assert.True(t, NotEmpty("str"))
	assert.True(t, NotEmpty(arrInt))
	assert.True(t, NotEmpty(arrStr))
	assert.True(t, NotEmpty(ptr))

	assert.True(t, IsIPv4("192.168.2.4"))
	assert.False(t, IsIPv4("2001:0db8:85a3:0000:0000:8a2e:0370:7334"))
}

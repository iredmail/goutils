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
	assert.True(t, IsNotEmpty("str"))
	assert.True(t, IsNotEmpty(arrInt))
	assert.True(t, IsNotEmpty(arrStr))
	assert.True(t, IsNotEmpty(ptr))
}

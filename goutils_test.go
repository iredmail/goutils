package goutils

import (
	"os"
	"path/filepath"
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

	t.Run("Test DestExists Func", func(t *testing.T) {
		tmpPth1 := filepath.Join(os.TempDir(), "abc.txt")
		tmpPth2 := filepath.Join(os.TempDir(), "def.txt")
		_ = os.Remove(tmpPth1)
		_ = os.Remove(tmpPth2)
		assert.False(t, DestExists(tmpPth1))
		_ = os.WriteFile(tmpPth2, []byte("test"), 0700)
		assert.True(t, DestExists(tmpPth2))
		_ = os.Remove(tmpPth2)
	})
}

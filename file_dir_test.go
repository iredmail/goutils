package goutils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDestExists(t *testing.T) {
	pth1 := filepath.Join(os.TempDir(), "1.txt")
	pth2 := filepath.Join(os.TempDir(), "2.txt")

	// Make sure files are absent.
	_ = os.Remove(pth1)
	_ = os.Remove(pth2)

	assert.False(t, DestExists(pth1))

	err := os.WriteFile(pth2, []byte("test"), 0700)
	assert.Nil(t, err)
	assert.True(t, DestExists(pth2))

	err = os.Remove(pth2)
	assert.Nil(t, err)
	assert.False(t, DestExists(pth2))
}

func TestReadFullFileContent(t *testing.T) {
	var content []byte
	var s string
	var err error

	pth := filepath.Join(os.TempDir(), "1.txt")
	err = os.WriteFile(pth, []byte("test"), 0700)
	assert.Nil(t, err)

	content, err = ReadFullFileContent(pth)
	assert.Nil(t, err)
	assert.Equal(t, []byte("test"), content)

	s, err = ReadFullFileContentInString(pth)
	assert.Nil(t, err)
	assert.Equal(t, "test", s)

	err = os.WriteFile(pth, []byte("\n\ttest\r\n"), 0700)
	assert.Nil(t, err)
	assert.Equal(t, []byte("test"), content)

	s, err = ReadFullFileContentInString(pth)
	assert.Nil(t, err)
	assert.Equal(t, "test", s)

	_ = os.Remove(pth)
	assert.False(t, DestExists(pth))
}

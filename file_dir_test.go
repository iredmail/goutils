package goutils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileDir(t *testing.T) {
	pth1 := filepath.Join(os.TempDir(), "1.txt")
	pth2 := filepath.Join(os.TempDir(), "2.txt")

	// Make sure files are absent.
	_ = os.Remove(pth1)
	_ = os.Remove(pth2)

	assert.False(t, DestExists(pth1))

	_ = os.WriteFile(pth2, []byte("test"), 0700)
	assert.True(t, DestExists(pth2))

	_ = os.Remove(pth2)
	assert.False(t, DestExists(pth2))
}

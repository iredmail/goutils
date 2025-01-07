package pwhash

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenArgon2IDPassword(t *testing.T) {
	plainPassword := "plainPassword"

	// With prefixed scheme.
	hash, err := GenArgon2IDPassword(plainPassword, true)
	assert.Nil(t, err)
	assert.True(t, strings.HasPrefix(hash, "{ARGON2ID}"))

	matched, err := VerifyArgon2IDPassword(plainPassword, hash)
	assert.Nil(t, err)
	assert.True(t, matched)

	// Without prefixed scheme.
	hash, err = GenArgon2IDPassword(plainPassword)
	assert.Nil(t, err)
	assert.False(t, strings.HasPrefix(hash, "{ARGON2ID}"))

	matched, err = VerifyArgon2IDPassword(plainPassword, hash)
	assert.Nil(t, err)
	assert.True(t, matched)
}

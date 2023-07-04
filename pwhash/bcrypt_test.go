package pwhash

import (
	"testing"

	"github.com/iredmail/goutils"
	"github.com/stretchr/testify/assert"
)

func TestBcrypt(t *testing.T) {
	// key is plain password, value is password hash.
	data := map[string]string{
		"test1": "{BLF-CRYPT}$2a$10$yxGkpyOkEjBJ81YG7z6N7.OZDgdC7dsrFi54fCxXYGmHCAixjxeTK",
		"test2": "{CRYPT}$2y$10$c5km82jGk1Iw75I5wL31Juw9mRQNW6XKVoLC5T.jB3yrxD1GYcWyu",
	}

	var ok, matched bool
	var hash string

	for plain, hashed := range data {
		ok, _ = IsBcryptHash(hashed)
		assert.True(t, ok)

		_, hash = extractSchemeAndHash(hashed)
		matched = VerifyBcryptPassword(hashed, plain)
		assert.True(t, matched)
	}

	plain := goutils.GenRandomString(12)
	hash, err := GenerateBcryptPassword(plain)
	assert.Nil(t, err)
	assert.True(t, VerifyBcryptPassword(hash, plain))
}

package pwhash

import (
	"testing"

	"github.com/iredmail/goutils"
	"github.com/stretchr/testify/assert"
)

func TestBcrypt(t *testing.T) {
	plainPassword := goutils.GenRandomString(12)
	challengePassword, err := GenerateBcryptPassword(plainPassword)
	assert.Nil(t, err)
	matched := VerifyBcryptPassword(challengePassword, plainPassword)
	assert.True(t, matched)
}

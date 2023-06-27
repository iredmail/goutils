package pwhash

import (
	"testing"

	"github.com/iredmail/goutils"
	"github.com/stretchr/testify/assert"
)

func TestMD5(t *testing.T) {
	plainPassword := goutils.GenRandomString(12)
	challengePassword := GenerateMD5Password(plainPassword)
	assert.True(t, VerifyMD5Password(challengePassword, plainPassword))

	// plain
	plainPassword = goutils.GenRandomString(12)
	challengePassword, err := GeneratePlainMD5Password(plainPassword)
	assert.Nil(t, err)
	assert.True(t, VerifyPlainMD5Password(challengePassword, plainPassword))
}

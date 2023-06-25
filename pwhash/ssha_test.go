package pwhash

import (
	"testing"

	"github.com/iredmail/goutils"
	"github.com/stretchr/testify/assert"
)

func TestSSHA(t *testing.T) {
	plainPassword := goutils.GenRandomString(12)
	challengePassword, err := GenerateSSHAPassword(plainPassword)
	assert.Nil(t, err)
	matched, err := VerifySSHAPassword(challengePassword, plainPassword)
	assert.Nil(t, err)
	assert.True(t, matched)
}

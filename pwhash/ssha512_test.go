package pwhash

import (
	"testing"

	"github.com/iredmail/goutils"
	"github.com/stretchr/testify/assert"
)

func TestSSHA512(t *testing.T) {
	plainPassword := goutils.GenRandomString(12)
	challengePassword, err := GenerateSSHA512Password(plainPassword)
	assert.Nil(t, err)
	assert.True(t, VerifySSHA512Password(challengePassword, plainPassword))
}

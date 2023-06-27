package pwhash

import (
	"testing"

	"github.com/iredmail/goutils"
	"github.com/stretchr/testify/assert"
)

func TestSHA512(t *testing.T) {
	plainPassword := goutils.GenRandomString(12)
	challengePassword, err := GenerateSHA512Password(plainPassword)
	assert.Nil(t, err)
	assert.True(t, VerifySHA512Password(challengePassword, plainPassword))
}

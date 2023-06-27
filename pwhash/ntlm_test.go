package pwhash

import (
	"testing"

	"github.com/iredmail/goutils"
	"github.com/stretchr/testify/assert"
)

func TestNTLM(t *testing.T) {
	plainPassword := goutils.GenRandomString(12)
	challengePassword, err := GenerateNTLMPassword(plainPassword)
	assert.Nil(t, err)
	matched := VerifyNTLMPassword(challengePassword, plainPassword)
	assert.True(t, matched)
}

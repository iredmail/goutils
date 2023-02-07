package goutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSemVer(t *testing.T) {
	assert.True(t, IsValidSemVerion("0.9.0"))
	assert.True(t, IsValidSemVerion("v0.9.0"))
	assert.True(t, IsValidSemVerion("v1.0"))
	assert.True(t, IsValidSemVerion("v1.0.0"))

	assert.False(t, HasNewVersion("1.0.0", "1.0.0"))
	assert.True(t, HasNewVersion("0.9.0", "1.0.0"))
	assert.True(t, HasNewVersion("1.0.0", "1.0.1"))
	assert.True(t, HasNewVersion("1.0.0", "1.1.0"))
	assert.True(t, HasNewVersion("1.0.0", "2.0.0"))
}

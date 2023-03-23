package goutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringSliceToLower(t *testing.T) {
	ss := []string{"A", "b", "C"}

	StringSliceToLower(ss)
	assert.Equal(t, "a", ss[0])
	assert.Equal(t, "b", ss[1])
	assert.Equal(t, "c", ss[2])
}

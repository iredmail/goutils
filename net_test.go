package goutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNet(t *testing.T) {
	assert.Equal(t, true, IsIP("192.168.0.1"))
	assert.Equal(t, true, IsIP("192.168.0.0"))
	assert.Equal(t, false, IsIP("192.168.0.0/24"))

	assert.Equal(t, true, IsCIDR("192.168.0.0/24"))
	assert.Equal(t, false, IsCIDR("192.168.0.0"))
}

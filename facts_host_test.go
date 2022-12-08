package goutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGatherOSInfo(t *testing.T) {
	oi, err := GatherOSInfo()
	assert.Nil(t, err)
	// fmt.Println(osInfo.JSON())

	m, err := oi.ToMap()
	assert.Nil(t, err)
	assert.NotNil(t, m)
}

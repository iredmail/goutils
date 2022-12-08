package goutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGatherOSInfo(t *testing.T) {
	_, err := GatherOSInfo()
	assert.Nil(t, err)
	// fmt.Println(osInfo.JSON())
}

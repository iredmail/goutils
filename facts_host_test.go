package goutils

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGatherOSInfo(t *testing.T) {
	osInfo, err := GatherOSInfo()
	assert.Nil(t, err)
	fmt.Println(osInfo.ToMap())
}

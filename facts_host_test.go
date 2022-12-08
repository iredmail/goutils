package goutils

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGatherOSInfo(t *testing.T) {
	info, err := GatherOSInfo()
	assert.Nil(t, err)
	fmt.Println(info)
}

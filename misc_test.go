package goutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetNewAndRemoved(t *testing.T) {
	added, removed, remained := GetNewAndRemoved([]string{"a", "c", "b", "d"}, []string{"a", "x", "b", "y"})
	assert.Equal(t, added, []string{"x", "y"})
	assert.Equal(t, removed, []string{"c", "d"})
	assert.Equal(t, remained, []string{"a", "b"})

	added, removed, remained = GetNewAndRemoved([]string{}, []string{"a", "b", "x", "y"})
	assert.Equal(t, added, []string{"a", "b", "x", "y"})
	assert.Equal(t, len(removed), 0)
	assert.Equal(t, len(remained), 0)

	added, removed, remained = GetNewAndRemoved([]string{"a", "b", "c", "d"}, []string{})
	assert.Equal(t, len(added), 0)
	assert.Equal(t, len(removed), 4)
	assert.Equal(t, len(remained), 0)

	addedInt, removedInt, remainedInt := GetNewAndRemoved([]int{1, 2, 3, 4}, []int{2, 4, 8, 9})
	assert.Equal(t, addedInt, []int{8, 9})
	assert.Equal(t, removedInt, []int{1, 3})
	assert.Equal(t, remainedInt, []int{2, 4})

	addedInt, removedInt, remainedInt = GetNewAndRemoved([]int{}, []int{1, 2, 3, 4})
	assert.Equal(t, addedInt, []int{1, 2, 3, 4})
	assert.Equal(t, len(removedInt), 0)
	assert.Equal(t, len(remainedInt), 0)

	addedInt, removedInt, remainedInt = GetNewAndRemoved([]int{1, 2, 3, 4}, []int{})
	assert.Equal(t, len(addedInt), 0)
	assert.Equal(t, removedInt, []int{1, 2, 3, 4})
	assert.Equal(t, len(remainedInt), 0)
}

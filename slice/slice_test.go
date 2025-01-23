package slice

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddMissingElems(t *testing.T) {
	s := []int{1, 2, 1, 3, 5, 4, 2, 3, 5, 4, 5}
	elems := []int{2, 3, 7, 4, 6}

	newS := AddMissingElems(s, elems...)
	assert.Equal(t, []int{1, 2, 1, 3, 5, 4, 2, 3, 5, 4, 5, 7, 6}, newS)
}

func TestDeleteElems(t *testing.T) {
	s := []int{1, 2, 1, 3, 5, 4, 2, 3, 5, 4, 5}
	elems := []int{2, 3, 4}

	newS := DeleteElems(s, elems...)
	assert.Equal(t, []int{1, 1, 5, 5, 5}, newS)
}

func TestDeduplicateAndSort(t *testing.T) {
	assert.Equal(t, []int{1, 2, 3, 4, 5}, DeduplicateAndSort([]int{1, 5, 4, 2, 3, 1, 4, 5, 3, 4, 2, 3, 5, 4, 5}))
}

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

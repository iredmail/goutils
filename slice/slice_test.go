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

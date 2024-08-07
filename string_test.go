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

func TestFlattenToStrings(t *testing.T) {
	errValue := 1
	str := "str"
	array := []string{"1", "2", "3"}
	arrayAny := []any{"1", 2, "3", 4, "5"}
	twoArrayString := [][]string{
		{"1", "2"},
		{"4", "5"},
	}
	twoArrayAny := [][]any{
		{"1", "2", 3},
		{"4", "5", nil},
	}

	_, err := FlattenToStrings(errValue)
	assert.True(t, err != nil)

	values, err := FlattenToStrings(str)
	assert.Nil(t, err)
	assert.Equal(t, len(values), 1)

	values, err = FlattenToStrings(array)
	assert.Nil(t, err)
	assert.Equal(t, len(values), 3)

	values, err = FlattenToStrings(arrayAny)
	assert.Nil(t, err)
	assert.Equal(t, len(values), 3)

	values, err = FlattenToStrings(twoArrayAny)
	assert.Nil(t, err)
	assert.Equal(t, len(values), 4)

	values, err = FlattenToStrings(twoArrayString)
	assert.Nil(t, err)
	assert.Equal(t, len(values), 4)
}

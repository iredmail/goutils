package goutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFlattenStrings(t *testing.T) {
	var empty []string

	assert.Equal(t, empty, FlattenStrings(nil))
	assert.Equal(t, []string{"1"}, FlattenStrings("1"))
	assert.Equal(t,
		[]string{"1", "2"},
		FlattenStrings([]string{"1", "2", "", "1", "2"}), // empty and duplicate values
	)

	// numbers will be ignored.
	assert.Equal(t, []string{"1"}, FlattenStrings([]any{"1", 1, 2, 3}))

	// Mixed.
	assert.Equal(t,
		[]string{"1", "2", "3", "4"},
		FlattenStrings([]any{
			"1",
			[]string{"2", "3"},
			"4"},
		),
	)

	// Nested.
	assert.Equal(t,
		[]string{"1", "2", "3", "4", "5", "6", "7"},
		FlattenStrings(
			[]any{
				"1",
				"",                          // empty
				"1",                         // duplicate
				[]string{"2", "3", "", "2"}, // empty + duplicate

				// 2 levels.
				[][]string{
					{"4", "5", "", "5"}, // empty + duplicate
				},

				// 3 levels.
				[][][]string{
					{
						{"6", "7", "", "7"}, // empty + duplicate
					},
				},
			},
		),
	)

	// Unsupported type.
	assert.Equal(t, empty, FlattenStrings(1))
}

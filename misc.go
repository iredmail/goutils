package goutils

import "slices"

// GetNewAndRemoved compares old (o) and new (n) slices, returns new, removed and remained items.
func GetNewAndRemoved[T comparable](o, n []T) (added, removed, remained []T) {
	for _, v := range o {
		if slices.Contains(n, v) {
			remained = append(remained, v)
		} else {
			removed = append(removed, v)
		}
	}

	for _, v := range n {
		if !slices.Contains(o, v) {
			added = append(added, v)
		}
	}

	return
}

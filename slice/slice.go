package slice

import (
	"slices"

	"golang.org/x/exp/constraints"
)

// AddMissingElems 添加 slice `s` 里缺失的所有 `elems` 元素。
func AddMissingElems[T comparable](s []T, elems ...T) []T {
	for _, elem := range elems {
		if !slices.Contains(s, elem) {
			s = append(s, elem)
		}
	}

	return s
}

// DeleteElems 移除 slice `s` 里出现的所有 `elems` 元素。如果不存在则忽略。
func DeleteElems[T comparable](s []T, elems ...T) (newS []T) {
	if len(elems) == 0 {
		return s
	}

	for _, elem := range s {
		if !slices.Contains(elems, elem) {
			newS = append(newS, elem)
		}
	}

	return
}

// DeduplicateAndSort 移除 slice `s` 里的所有重复元素，并按升序排序。
func DeduplicateAndSort[T constraints.Ordered](s []T) (newS []T) {
	m := make(map[T]bool)

	for _, elem := range s {
		m[elem] = true
	}

	for k := range m {
		newS = append(newS, k)
	}

	slices.Sort(newS)

	return
}

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

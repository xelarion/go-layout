package util

import "slices"

// DifferenceString returns the difference between two collections of strings.
func DifferenceString(x []string, y []string) ([]string, []string) {
	var leftSlice []string
	var rightSlice []string

	for _, v := range x {
		if !slices.Contains(y, v) {
			leftSlice = append(leftSlice, v)
		}
	}

	for _, v := range y {
		if !slices.Contains(x, v) {
			rightSlice = append(rightSlice, v)
		}
	}

	return leftSlice, rightSlice
}

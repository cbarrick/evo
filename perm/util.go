package perm

import (
	"fmt"
)

// Search searches an int slice for a particular value and returns the index.
// If the value is not found, search returns -1.
func search(slice []int, val int) (idx int) {
	for idx = range slice {
		if slice[idx] == val {
			return idx
		}
	}
	return -1
}

// Reverse reverses an int slice.
func reverse(slice []int) {
	i := 0
	j := len(slice) - 1
	for i < j {
		slice[i], slice[j] = slice[j], slice[i]
		i++
		j--
	}
}

// Validate panics if the argument is not a permutation.
func validate(slice []int) {
	n := len(slice)
	for i := 0; i < n; i++ {
		if search(slice, i) == -1 {
			fmt.Println()
			fmt.Println(slice)
			panic("invalid permutation")
		}
	}
}

package perm

import (
	"math/rand"
)

// RandInvert reverses a random slice of the argument.
func RandInvert(gene []int) {
	slice, _, _ := RandSlice(gene)
	Reverse(slice)
}

// RandSwap swaps two random elements of the argument.
func RandSwap(gene []int) {
	size := len(gene)
	i := rand.Intn(size)
	j := i
	for j == i {
		j = rand.Intn(size)
	}
	gene[i], gene[j] = gene[j], gene[i]
}

package perm

import (
	"math/rand"
)

// search searches an int slice for a particular value and returns the index.
// If the value is not found, search returns -1.
func search(slice []int, val int) (idx int) {
	for idx = range slice {
		if slice[idx] == val {
			return idx
		}
	}
	return -1
}

// sublist takes a slice and returns a random subslice along with the boundaries.
func sublist(slice []int) (sub []int, left, right int) {
	left = rand.Intn(len(slice) - 1)
	right = left
	for right == left {
		right = rand.Intn(len(slice))
	}
	if right < left {
		left, right = right, left
	}
	return slice[left:right], left, right
}

// OrderX performs order crossover on two parents to create a child.
// Order crossover is analogous to 1-point crossover repaired for permutations.
// Order crossover is a very simple crossover technique for permutations.
func OrderX(mom, dad []int) (child []int) {
	child = make([]int, len(mom))
	sub, left, right := sublist(mom)
	copy(child[left:right], sub)
	i, j := right, right
	for i < left || right <= i {
		if search(sub, dad[j]) == -1 {
			child[i] = dad[j]
			i = (i + 1) % len(child)
		}
		j = (j + 1) % len(child)
	}
	return child
}

// PMX performs partially mapped crossover on two parents to create a child.
// PMX is often a good choice for a variety of permutation problems.
func PMX(mom, dad []int) (child []int) {
	child = make([]int, len(mom))
	_, left, right := sublist(mom)

	for i := range child {
		child[i] = -1
	}
	copy(child[left:right], mom[left:right])

	for i := left; i < right; i++ {
		if search(child, dad[i]) == -1 {
			j := i
			for left <= j && j < right {
				j = search(dad, mom[j])
			}
			child[j] = dad[i]
		}
	}

	for i := range child {
		if child[i] == -1 {
			child[i] = dad[i]
		}
	}

	return child
}

// CycleX performs cycle crossover on two parents to produce a child.
// Cycle crossover is a good choice when you want the inherited alleals to keep
// the position inherited from the parents.
func CycleX(mom, dad []int) (child []int) {
	var cycles [][]int
	taken := make([]bool, len(mom))
	for i := range mom {
		if !taken[i] {
			var cycle []int
			for j := i; !taken[j]; {
				taken[j] = true
				cycle = append(cycle, j)
				j = search(mom, dad[j])
			}
			cycles = append(cycles, cycle)
		}
	}

	child = make([]int, len(mom))
	var who bool
	for i := range cycles {
		var parent []int
		if who {
			parent = mom
		} else {
			parent = dad
		}
		for _, j := range cycles[i] {
			child[j] = parent[j]
		}
		if len(cycles[i]) > 1 {
			who = !who
		}
	}

	return child
}

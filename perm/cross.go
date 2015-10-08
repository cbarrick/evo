package perm

import (
	"math/rand"
)

// OrderX performs order crossover. Order crossover is a good choice when you
// want to inherit the relative order of values.
func OrderX(child, mom, dad []int) {
	if rand.Float64() < 0.5 {
		mom, dad = dad, mom
	}
	sub, left, right := RandSlice(mom)
	copy(child[left:right], sub)
	i, j := right, right
	for i < left || right <= i {
		if Search(sub, dad[j]) == -1 {
			child[i] = dad[j]
			i = (i + 1) % len(child)
		}
		j = (j + 1) % len(child)
	}
}

// PMX performs partially mapped crossover. PMX inherits a random slice of one
// parent. The position of the other values is more random when there is greater
// difference between the parents.
func PMX(child, mom, dad []int) {
	if rand.Float64() < 0.5 {
		mom, dad = dad, mom
	}
	_, left, right := RandSlice(mom)

	for i := range child {
		child[i] = -1
	}
	copy(child[left:right], mom[left:right])

	for i := left; i < right; i++ {
		if Search(child, dad[i]) == -1 {
			j := i
			for left <= j && j < right {
				j = Search(dad, mom[j])
			}
			child[j] = dad[i]
		}
	}

	for i := range child {
		if child[i] == -1 {
			child[i] = dad[i]
		}
	}
}

// CycleX performs cycle crossover. Cycle crossover is a good choice when you
// want to inherit the absolute position of values.
func CycleX(child, mom, dad []int) {
	if rand.Float64() < 0.5 {
		mom, dad = dad, mom
	}
	var cycles [][]int
	taken := make([]bool, len(mom))
	for i := range mom {
		if !taken[i] {
			var cycle []int
			for j := i; !taken[j]; {
				taken[j] = true
				cycle = append(cycle, j)
				j = Search(mom, dad[j])
			}
			cycles = append(cycles, cycle)
		}
	}

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
}

// EdgeX performs edge recombination. Edge recombination is a good choice when
// you want to inherit adjacency information.
func EdgeX(child, mom, dad []int) {
	dim := len(mom)
	child = child[0:0]

	if rand.Float64() < 0.5 {
		mom, dad = dad, mom
	}

	// build the table
	// doubles are marked by negating the entry
	table := make([][]int, dim)
	for i := range table {
		table[i] = make([]int, 0, 4)
	}
	for i := range table {
		var j int

		var mnext, mprev int
		j = Search(mom, i)
		if j == 0 {
			mnext = 1
			mprev = dim - 1
		} else if j == dim-1 {
			mnext = 0
			mprev = dim - 2
		} else {
			mnext = j + 1
			mprev = j - 1
		}
		table[i] = append(table[i], mom[mnext], mom[mprev])

		var dnext, dprev int
		j = Search(dad, i)
		if j == 0 {
			dnext = 1
			dprev = dim - 1
		} else if j == dim-1 {
			dnext = 0
			dprev = dim - 2
		} else {
			dnext = j + 1
			dprev = j - 1
		}
		if table[i][0] == dad[dnext] {
			table[i][0] = -table[i][0]
		} else if table[i][1] == dad[dnext] {
			table[i][1] = -table[i][1]
		} else {
			table[i] = append(table[i], dad[dnext])
		}
		if table[i][0] == dad[dprev] {
			table[i][0] = -table[i][0]
		} else if table[i][1] == dad[dprev] {
			table[i][1] = -table[i][1]
		} else {
			table[i] = append(table[i], dad[dprev])
		}
	}

	// clear removes all occurences of x in the table
	clear := func(x int) {
		for i := range table {
			newrow := table[i][0:0]
			pos := Search(table[i], x)
			neg := Search(table[i], -x)
			for j := range table[i] {
				if j != pos && j != neg {
					newrow = append(newrow, table[i][j])
				}
			}
			table[i] = newrow
		}
	}

	// main loop
	var reversed bool
	current := rand.Intn(dim)
	child = append(child, current)
	clear(current)
	for len(child) < dim {
		next := -1
		shortest := 5
		row := table[current]
		if len(row) == 0 {
			if !reversed {
				Reverse(child)
				reversed = true
				current = child[len(child)-1]
				continue
			} else {
				for next == -1 || Search(child, next) != -1 {
					next = rand.Intn(len(table))
				}
			}
		} else {
			for i := range row {
				if row[i] < 0 {
					next = -row[i]
					break
				} else if len(table[row[i]]) < shortest {
					shortest = len(table[row[i]])
					next = row[i]
				} else if len(table[row[i]]) == shortest {
					if rand.Float32() < 0.5 {
						next = row[i]
					}
				}
			}
		}
		reversed = false
		child = append(child, next)
		clear(next)
		current = next
	}
}

package integer

import "math/rand"

// UniformX performs a uniform crossover of some parents into a child.
func UniformX(child []int, parents ...[]int) {
	n := len(parents)
	for i := range child {
		child[i] = parents[rand.Intn(n)][i]
	}
}

// PointX performs n-point crossover of two parents into a child.
func PointX(n int, child, mom, dad []int) {
	if rand.Intn(2) == 0 {
		mom, dad = dad, mom
	}
	for 0 < n {
		i := rand.Intn(len(child)-n) + 1
		copy(child, mom[:i])
		child = child[i:]
		mom, dad = dad[i:], mom[i:]
		n--
	}
	copy(child, mom)
}

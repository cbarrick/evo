package real

import (
	"math/rand"
)

// UniformX performs a uniform crossover of some parents into a child.
func UniformX(child Vector, parents ...Vector) {
	n := len(parents)
	for i := range child {
		child[i] = parents[rand.Intn(n)][i]
	}
}

// ArithX performs arithmetic crossover. When the scale is 1, a child is chosen
// uniformly at random from the line segment between the parents. The scale
// affects the length of the segment about the midpoint. Thus when the scale is
// 0, the child is always the midpoint.
func ArithX(scale float64, child, mom, dad Vector) {
	// special case when scale == 0, we can find the midpoint in constant space
	if scale == 0 {
		copy(child, mom)
		child.Subtract(dad)
		child.Scale(0.5)
		child.Add(dad)
		return
	}

	copy(child, mom)
	child.Subtract(dad)
	mid := child.Copy()
	mid.Scale(0.5)
	child.Scale(scale*rand.Float64() - scale/2)
	child.Add(dad)
	child.Add(mid)
}

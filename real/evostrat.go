package real

import (
	"math"
)

// Adapt performs a lognormal scaling of the vector using a global learning
// rate of 1/sqrt(n) and a local learning rate of 1/sqrt(2*sqrt(n)). This is
// commonly used in evolution strategies to learn the strategy parameters.
func (v Vector) Adapt() {
	n := float64(len(v))
	globalrate := 1 / math.Sqrt(n)
	localrate := 1 / math.Sqrt(2*math.Sqrt(n))
	global := Lognormal(globalrate)
	for i := range v {
		v[i] *= Lognormal(localrate) * global
	}
}

// Step performs a gausian purterbation of the vector using position-wise
// step-sizes. This is commonly used in evolution strategies to mutate the
// object parameters, using the strategy parameters as the step-sizes.
func (v Vector) Step(steps Vector) {
	for i := range v {
		v[i] += Normal(steps[i])
	}
}

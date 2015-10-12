package real

import (
	"math/rand"
)

type Vector []float64

// Random generates a random vector of length n. Values are taken uniformly
// between [0,1) and multiplied by some scale.
func Random(n int, scale float64) (v Vector) {
	v = make(Vector, n)
	for i := range v {
		v[i] = rand.Float64() * scale
	}
	return v
}

func (v Vector) Copy() Vector {
	w := make(Vector, len(v))
	copy(w, v)
	return w
}

func (v Vector) Add(w Vector) {
	for i := range v {
		v[i] += w[i]
	}
}

func (v Vector) Subtract(w Vector) {
	for i := range v {
		v[i] -= w[i]
	}
}

func (v Vector) Scale(s float64) {
	for i := range v {
		v[i] *= s
	}
}

func (v Vector) LowPass(min float64) {
	for i := range v {
		if v[i] < min {
			v[i] = min
		}
	}
}

func (v Vector) HighPass(max float64) {
	for i := range v {
		if v[i] > max {
			v[i] = max
		}
	}
}

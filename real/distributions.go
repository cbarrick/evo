package real

import (
	"math"
	"math/rand"
)

func Normal(stdv float64) float64 {
	return stdv * rand.NormFloat64()
}

func Lognormal(rate float64) float64 {
	return math.Exp(Normal(rate))
}

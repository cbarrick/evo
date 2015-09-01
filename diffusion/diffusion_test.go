package diffusion_test

import (
	"math"
	"testing"

	"github.com/cbarrick/evo"
	"github.com/cbarrick/evo/diffusion"
	"github.com/cbarrick/evo/example/ackley" // imported as "main"
)

func BenchmarkAckley(b *testing.B) {
	size := 32
	dim := 16
	accuracy := 0.001
	convergence := math.Inf(+1)

	acks := make([]evo.Genome, size)
	for i := range acks {
		acks[i] = main.Random(dim)
	}
	population := diffusion.Hypercube(acks)

	b.ResetTimer()
	for convergence > accuracy {
		members := population.Members()
		max := evo.Max(members...).Fitness()
		min := evo.Min(members...).Fitness()
		convergence = max - min
	}
	population.Close()
	b.StopTimer()
}

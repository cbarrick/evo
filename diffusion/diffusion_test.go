package diffusion_test

import (
	"testing"

	"github.com/cbarrick/evo"
	"github.com/cbarrick/evo/diffusion"
	"github.com/cbarrick/evo/example/ackley" // imported as "main"
)

func BenchmarkAckley(b *testing.B) {
	size := 32
	dim := 16
	accuracy := 0.001

	acks := make([]evo.Genome, size)
	for i := range acks {
		acks[i] = main.Random(dim)
	}
	population := diffusion.Hypercube(acks)

	b.ResetTimer()
	for population.Stats().N["convergence"] > accuracy {
	}
	population.Close()
	b.StopTimer()
}

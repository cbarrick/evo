package graph_test

import (
	"testing"

	"github.com/cbarrick/evo"
	"github.com/cbarrick/evo/diffusion"
	"github.com/cbarrick/evo/example/ackley" // imported as "main"
)

func BenchmarkAckley(b *testing.B) {
	var (
		size      = 32
		dim       = 16
		accuracy  = 0.0001
		deviation float64
	)

	acks := make([]evo.Genome, size)
	for i := range acks {
		acks[i] = main.Random(dim)
	}
	population := diffusion.Hypercube(acks)

	update := func() {
		view := population.View()
		deviation = view.StdDeviation()
		view.Close()
	}

	update()
	for deviation > accuracy {
		update()
	}
	population.Close()
}

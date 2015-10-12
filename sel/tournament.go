package sel

import (
	"math"
	"math/rand"

	"github.com/cbarrick/evo"
)

func Tournament(genomes ...evo.Genome) (best evo.Genome) {
	var fit, bestfit float64
	bestfit = math.Inf(-1)
	for i := range genomes {
		fit = genomes[i].Fitness()
		if fit > bestfit {
			bestfit = fit
			best = genomes[i]
		}
	}
	return best
}

func BinaryTournament(suitors ...evo.Genome) evo.Genome {
	var x, y, size int
	size = len(suitors)
	if size > 2 {
		x = rand.Intn(size)
		y = x
		for y == x {
			y = rand.Intn(size)
		}
	} else {
		x, y = 0, 1
	}
	if suitors[x].Fitness() < suitors[y].Fitness() {
		return suitors[y]
	}
	return suitors[x]
}

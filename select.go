package evo

import (
	"math/rand"
)

func BinaryTournament(suiters ...Genome) Genome {
	if len(suiters) > 2 {
		x := rand.Intn(len(suiters))
		y := x
		for y == x {
			y = rand.Intn(len(suiters))
		}
	} else {
		x, y = 0, 1
	}
	if suiters[x].Fitness() < suiters[y].Fitness() {
		return suiters[y]
	}
	return suiters[x]
}

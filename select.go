package evo

import (
	"math/rand"
)

// BinaryTournament randomly selects two suitors and returns the one with the
// highest fitness.
func BinaryTournament(suitors ...Genome) Genome {
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

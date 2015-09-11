package evo

// Genomes must be comparable, i.e. pointer to slice instead of slice
type Genome interface {
	Fitness() float64 // must continue to work after close
	Cross(...Genome) Genome
	Close()
}

type Population interface {
	Genome
	View() View
}

// Functions
// -------------------------

func Tournament(suiters ...Genome) (max Genome) {
	max = suiters[0]
	for i := range suiters {
		if suiters[i].Fitness() > max.Fitness() {
			max = suiters[i]
		}
	}
	return max
}

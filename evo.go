package evo

// Genomes must be comparable, i.e. pointer to slice instead of slice
type Genome interface {
	Fitness() float64
	Cross(...Genome) Genome
	Difference(Genome) float64
	Close() error
}

type Population interface {
	Genome
	Stats() Stats
}

type Stats struct {
	Members []Genome
	Max     Genome
	Min     Genome
	N       map[string]float64
}

// Selection strategies
// -------------------------

// Tournament returns the genome with the highest fitness.
// Nil genomes are ignored. If all genomes are nil, nil is returned.
func Tournament(gs ...Genome) Genome {
	switch {
	case len(gs) == 1:
		return gs[0]

	case gs[0] == nil:
		return Tournament(gs[1:]...)

	case gs[1] == nil:
		gs[1] = gs[0]
		return Tournament(gs[1:]...)

	default:
		if gs[0].Fitness() > gs[1].Fitness() {
			gs[1] = gs[0]
		}
		return Tournament(gs[1:]...)
	}
}

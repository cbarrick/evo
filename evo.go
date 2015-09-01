package evo

// Genomes must be comparable, i.e. pointer to slice instead of slice
type Genome interface {
	Fitness() float64
	Cross(...Genome) Genome
	Close() error
}

type Population interface {
	Genome
	Members() []Genome
}

// Functions
// -------------------------

// Max returns the genome with the highest fitness.
func Max(gs ...Genome) (max Genome) {
	max = gs[0]
	for i := range gs {
		if gs[i].Fitness() > max.Fitness() {
			max = gs[i]
		}
	}
	return max
}

// Min returns the genome with the lowest fitness.
func Min(gs ...Genome) (min Genome) {
	min = gs[0]
	for i := range gs {
		if gs[i].Fitness() < min.Fitness() {
			min = gs[i]
		}
	}
	return min
}

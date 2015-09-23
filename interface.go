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

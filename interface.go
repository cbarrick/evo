package evo

// An EvolveFn describes an iteration of the evolution loop. The evolve function
// is called once for each member of the population, possibly in parrallel, and
// is responsible for producing new Genomes given some subset of the population,
// called the suitors. The replacement Genome replaces the current Genome within
// the population.
type EvolveFn func(current Genome, suitors []Genome) (replacement Genome)

// A Genome describes the function being optimized and the representation of
// solutions. Genomes are provided by the user, and Evo provides convenience
// packages for common representations.
type Genome interface {
	// The Fitness method is the function being maximized.
	// For minimization problems, return the negative or the inverse.
	Fitness() float64
}

// A Population models the interaction between Genomes during evolution. In
// practice, this determines the kind of parallelism and number of suitors
// during the optimization.
//
// Populations implement Genome, making them composable. For example, an island
// model can be built by composing generational populations into a graph
// population.
type Population interface {
	// Fitness returns the maximum fitness of the population.
	Fitness() float64

	// Evolve starts the evolution of the population in a separate goroutine.
	// Genomes are evolved in place; it is not safe to access the genome slice
	// while the evolution is running.
	Evolve([]Genome, EvolveFn)

	// Stop terminates the optimization.
	Stop()

	// Stats returns various statistics about the population.
	Stats() Stats
}

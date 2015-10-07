package evo

// Genomes are members of populations and define the inner loop of the
// evolutionary algorithm. The inner loop is defined by the Cross method: For
// each iteration, the Cross method is called in parallel on some subset of the
// population. The method receives some subset of the population as arguments,
// and the value returned will replace the caller in the population. The user is
// expected to define a genome for their search space.
//
// Genomes may implement two optional methods, Fitness and Close. The fitness
// method has no effect on the population, but may be used by helper methods
// included within Evo. The Stats method of populations returns statistics of
// the population's fitness. You may simply return 0 if you do not wish to
// implement Fitness.
//
// The Close method will be registered as a finalizer on the genome. This can
// be used to recycle memory from dying genomes to newborn genomes. The Close
// method is not guarenteed to be called. See https://godoc.org/runtime#SetFinalizer
//
// Genomes must be pointer types.
type Genome interface {
	Cross(...Genome) Genome
	Fitness() float64
	Close()
}

// Populations control a set of genomes and define the architecture of the
// evolutionary algorithm. Populations implement Genome, so novel architectures,
// like the island model, can be implemented by nesting populations.
//
// Populations execute the evolutionary algorithm in a separate goroutine. The
// Close method stops that process.
//
// The Stats method returns statistics about the fitness of genomes in the
// population.
//
// The Iter method returns an iterator over the genomes in the population. If
// called on a meta-population, i.e. one whose members are themselves
// populations, then the iterator walks over the leaf-level genomes.
type Population interface {
	Genome
	Stats() Stats
	Iter() Iterator
}

// Iterators are returned by Population.Iter(). To iterate over a population:
//     for i := pop.Iter(); i.Value() != nil; i.Next() {
//         // do something with i.Value()
//     }
type Iterator interface {
	// Value returns nil once the iterator has reached the end.
	Value() Genome

	// Next advances the iterator. A nil-pointer panic occurs when advancing
	// beyond the end of the population.
	Next()
}

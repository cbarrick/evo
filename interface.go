package evo

import (
	"math"
)

// Genomes represent candidate solutions and are implemented by the user. Your
// genome type defines the inner loop of the evolutionary algorithm as the
// Evolve method: For each genome in a population, the Evolve method is called,
// passing some subset of the population, called the suitors, as arguments.
// The value returned is will replace the caller within the population. In a
// typical geneticl algorithm, the Evolve method will select one or more parents
// from the suiters, apply some kind of recombination operator to generate a new
// child genome, possibly apply a mutation operator, and return the child. A
// more complicated replacement scheme may require coordinating with other
// genomes in the population.
//
// Genomes may implement two optional methods, Fitness and Close. The Fitness
// method has no effect on the population, but may be used by helper methods
// included within Evo. The Stats method of the Population type returns
// statistics of the population's fitness, thus Fitness must be non-trivial to
// use this feature. You may simply return 0 if you do not wish to implement
// Fitness.
//
// The Close method will be registered as a finalizer on the genome. This can
// be used to recycle memory from dying genomes to newborn genomes. The Close
// method is not guarenteed to be called.
// See https://godoc.org/runtime#SetFinalizer
//
// Genomes must be pointer types.
type Genome interface {
	Evolve(...Genome) Genome
	Fitness() float64
	Close()
}

// Populations orchestrate the evolution of genomes.
//
// Populations implement Genome by providing uniform random migration as the
// Evolve method. Thus architectures like the island model can be implemented by
// nesting populations.
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
//
// The builtin populations live in the package evo/pop.
type Population interface {
	Genome
	Stats() Stats
	Iter() Iterator
}

// Max returns the genome with the highest fitness.
func Max(pop Population) (best Genome) {
	var val Genome
	var fit, bestfit = float64(0), math.Inf(-1)
	for i := pop.Iter(); i.Value() != nil; i.Next() {
		val = i.Value()
		fit = val.Fitness()
		if fit > bestfit {
			best = val
			bestfit = fit
		}
	}
	return best
}

// Iterators iterate over the genomes in a population and are returned by
// Population.Iter().
//
// The Value method returns the current Genome pointed to by the iterator, or
// nil if the iterator has reached the end of the population.
//
// The Next method advances the iterator.
//
// To iterate over a population:
//     for i := pop.Iter(); i.Value() != nil; i.Next() {
//         // do something with i.Value()
//     }
type Iterator interface {
	Value() Genome
	Next()
}

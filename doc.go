// Evo is a framework for implementing evolutionary algorithms in Go.
//
// Evo exposes a clean and flexible API oriented around two interfaces: `Genome`
// and `Population`. Genomes represent both the function being optimized and the
// representation of solutions. Populations represent the architecture under
// which genomes are evolved. Multiple population types are provided by Evo to
// enable the construction of both common and novel architectures.
//
// The body of the evolutionary loop is defined by an evolve function. For each
// genome in a population, the evolve function is called, receiving some subset
// of the population, called the suitors, as arguments. The evolve function then
// applies the user's variation operators (selection, mutation, etc) and returns
// a genome for the next iteration. common operators for a variety of
// representations are provided as subpackages of Evo.
//
// Populations model the evolution patterns of genomes. A few different
// population types are provided by Evo under the package `evo/pop`. Populations
// themselves implement the Genome interface, making them composeable. Migration
// functions are provided to be used in this context, allowing go novel
// architectures like the island model.
package evo

// TODO: Keep this in sync with the readme

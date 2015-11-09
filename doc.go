// Evo is a framework for implementing evolutionary algorithms in Go.
//
// Evo exposes a clean and flexible API oriented around two interfaces: `Genome`
// and `Population`. Genomes represent candidate solutions to the user's problem
// and are implemented by the user. Genomes define their own means of evolution,
// allowing for a multiplicity of techniques ranging from genetic algorithms to
// evolution strategies and beyond. Populations represent the architecture under
// which genomes are evolved. Multiple population types are provided by Evo to
// enable the construction of both common and novel architectures.
//
// The body of the evolutionary loop is defined by the Evolve method of the
// Genome type being evolved. For each genome in a population, the Evolve method
// is called, receiving some subset of the population, called the suitors, as
// arguments. The Evolve method then applies operators to the suiters
// (selection, mutation, etc) and returns a genome that will replace the caller
// within the population for the next iteration. The concrete genome type is
// problem specific and defined by the user, while common operators for a
// variety of domains are provided as subpackages of Evo.
//
// Populations orchestrate the evolution of genomes. A few different population
// types are provided by Evo under the package `evo/pop`. Populations themselves
// implement the Genome interface, making them composeable. The Evolve method of
// the builtin populations implements uniform random migration: A random
// population is chosen from the pool of suitors. Then the first population and
// its suitor exchange random members. This allows novel architectures like the
// island model to be implemented by nesting populations.
package evo

// TODO: Keep this in sync with the readme

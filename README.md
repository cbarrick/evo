# Evo

Evo is a package to assist the implementation of evolutionary algorithms in Go.

```
go get github.com/cbarrick/evo
```


## Status

Evo is a young project under active development. Reviews and comments on both the interface and implementation are welcome. An official (beta) release is planned for December 2015.


## Documentation

https://godoc.org/github.com/cbarrick/evo


## Overview

Evo exposes a clean and flexible API oriented around two interfaces: `Genome` and `Population`. Genomes represent candidate solutions to the user's problem and are implemented by the user. Genomes define their own means of evolution, allowing for a multiplicity of techniques ranging from genetic algorithms to evolution strategies and beyond. Several composeable population types are provided by Evo to enable the construction of both common and novel architectures.

Genomes define the body of the evolutionary loop as the Evolve method. For each genome in a population, the Evolve method is called, passing some subset of the population, called the suitors, as arguments. The Evolve method then applies operators to the suiters (selection, mutation, etc) and returns a genome that will replace the caller within the population for the next iteration. Common operators for a variety of representations are provided by Evo.

Populations orchestrate the evolution of genomes. Populations provided by Evo live under the package `evo/pop`. The `generational` population implements a traditional generation-based loop with master-slave parallelism. Each genome receives the entire population as suitors, and the population is only updated after all genomes have returned. The `graph` population maps each genome to a node in a graph. Each genome only receives the neighboring genomes as suitors, and each node is evolved in parallel.

Populations themselves implement the Genome interface. The Evolve method on populations implements uniform random migration: A random suitor is chosen and asserted to be a population of the same type. Then the population and its suitor exchange random members. This allows novel architectures like the island model to be implemented by nesting populations.


## Examples

You can browse example problems in the [example subpackage](https://github.com/cbarrick/evo/tree/master/example). The examples are maintained as a development tool rather than to provide optimal solutions to the problems they tackle. Reading the examples should give you a good idea of how easy it is to write code with Evo.


## License

This program is free software: you can redistribute it and/or modify it under the terms of the GNU Lesser General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU Lesser General Public License for more details.

You should have received a copy of the GNU Lesser General Public License along with this program. If not, see <http://www.gnu.org/licenses/>.

# Evo

Evo is a library in the Go programming language for genetic algorithms and other evolutionary computation techniques.


## Overview

Working with Evo is clean and flexible. The API is oriented around two interfaces: `Genome` and `Population`. Genomes represent candidate solutions to the user's problem and are implemented by the user. The genetic encoding of the solution and the domain-specific operation of the evolution loop are defined in the context of a genome type. Populations are collections of genomes that evolve over time. Different populations provided different kinds of evolution loops and are provided by Evo.

### Genomes

Genomes are defined by the user and both encode a candidate solution and define the procedure for evolving better solutions. The evolution procedure is a method called `Cross` defined on your genome type. For each position in a population, the Cross method is called, passing a mating pool as the argument. It is important to note that no selection pressure is applied to that mating pool by the system. It is the responsibility of the Cross method to select zero or more "parent" genomes and produce a "child" genome by applying domain-specific operators (crossover, mutation, etc). The method receiver is the genome currently occupying the position, and the return value will replace that genome in the population.

The Cross method will be called in parallel for different positions of the population. If you desire greater control, you can coordinate the evolution of a population using a goroutine. However, by synchronizing you will be removing an important opportunity for massive parallelism.

Genomes may optionally implement a `Close` method if they control any closable resources. The Close method is called in two cases. First, if a call to Cross results in a genome being replaced, that genome is closed. Second, when the population is closed, all member genomes are closed.

### Populations

Populations are collections of genomes that evolve over time. Each population provides slightly different semantics for the evolution loop. For specific details, see the documentation for each population type.

Populations are composable, allowing the user to design new and novel evolution loops. Composability is accomplished because each population type also implements the genome interface. Typically, when `Cross` is called on a population, the receiver picks a random population from the mating pool and incorporates the best genome of that population into itself. The `Close` method of all populations simply stops the underlying evolution loop.


## Examples

You can browse example problems in the [example subpackage](https://github.com/cbarrick/evo/tree/master/example)


## License

This program is free software: you can redistribute it and/or modify it under the terms of the GNU Lesser General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU Lesser General Public License for more details.

You should have received a copy of the GNU Lesser General Public License along with this program. If not, see <http://www.gnu.org/licenses/>.

# Evo

Evo is a framework for implementing evolutionary algorithms in Go.

```
go get github.com/cbarrick/evo
```


## Documentation

https://godoc.org/github.com/cbarrick/evo


## Status & Contributing

Evo is a general framework for developing genetic algorithms and more. It began life in fall 2015 as I studied evolutionary algorithms as an undergrad. The most recent release is [v0.1.1].

Contributions are welcome! I am currently a student and inexperienced maintainer, so please bear with me as I learn. The focus of the project thus far has been on the API design and less on performance. I am particularly interested in hearing about use-cases that are not well covered and success stories of where Evo excels. Testing, code reviews, and performance audits are always welcome and needed.

[v0.1.1]: https://github.com/cbarrick/evo/tree/v0.1.1


## Overview

Evo exposes a clean and flexible API oriented around two interfaces: `Genome` and `Population`. Genomes represent candidate solutions to the user's problem and are implemented by the user. Genomes define their own means of evolution, allowing for a multiplicity of techniques ranging from genetic algorithms to evolution strategies and beyond. Populations represent the architecture under which genomes are evolved. Multiple population types are provided by Evo to enable the construction of both common and novel architectures.

The body of the evolutionary loop is defined by the Evolve method of the Genome type being evolved. For each genome in a population, the Evolve method is called, receiving some subset of the population, called the suitors, as arguments. The Evolve method then applies operators to the suiters (selection, mutation, etc) and returns a genome that will replace the caller within the population for the next iteration. The concrete genome type is problem specific and defined by the user, while common operators for a variety of domains are provided as subpackages of Evo.

Populations orchestrate the evolution of genomes. A few different population types are provided by Evo under the package `evo/pop`. Populations themselves implement the Genome interface, making them composeable. The Evolve method of the builtin populations implements uniform random migration: A random population is chosen from the pool of suitors. Then the first population and its suitor exchange random members. This allows novel architectures like the island model to be implemented by nesting populations.


## Examples

You can browse example problems in the [example subpackage]. The examples are maintained as a development tool rather than to provide optimal solutions to the problems they tackle. Reading the examples should give you a good idea of how easy it is to write code with Evo.

[example subpackage]: https://github.com/cbarrick/evo/tree/master/example


## License

This program is free software: you can redistribute it and/or modify it under the terms of the GNU Lesser General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU Lesser General Public License for more details.

You should have received a copy of the GNU Lesser General Public License along with this program. If not, see <http://www.gnu.org/licenses/>.

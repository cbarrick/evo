# Evo Examples

This package contains examples of Evo being used on well know problems. The goal of the examples is not to be the best at solving the particular problem but instead to highlight various use-cases of Evo. Each example is implemented as a unit test and can be easily run with the `go test` command. For example, to run the `ackley` example from the source root:

    go test ./example/ackley -v

## Descriptions

- `ackley`: This example minimizes the Ackley function, a standard benchmark function for real-valued optimization. The problem is highly multimodal with a global minimum of 0 at the origin. The example minimizes the function in 30 dimensions with a self-adaptive (40/2,280)-evolution strategy.

- `queens`: This example solves the 128-queens problem by minimizing the number of conflicts on the board. The example highlights nested populations by implementing an island model where the population is divided among several sub-populations, called islands, and each island is evolved independently and in parallel. Occasionally migrations of individuals occur between the islands to serve as sources of new genes.

- `tsp`: This example searches for a minimal tour of the capitals of the 48 contiguous American states (dataset ATT48 of [TSPLIB]). The example uses a diffusion model, where is population is arranged in a hypercube and individuals breed only with their neighbors. The example also highlights hybridization with local search by using a 2-opt hillclimber as a mutation.

[TSPLIB]: http://comopt.ifi.uni-heidelberg.de/software/TSPLIB95/

package main

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/cbarrick/evo"
	"github.com/cbarrick/evo/perm"
	"github.com/cbarrick/evo/pop/gen"
	"github.com/cbarrick/evo/pop/graph"
	"github.com/cbarrick/evo/sel"
)

// Tuneables
const (
	dim  = 256     // the dimension of the board
	size = dim * 2 // the size of the population
	isl  = 4       // the number of islands in which to divide the population

	// the delay between island communications
	delay = 500 * time.Millisecond
)

// Global objects
var (
	// Count of the number of fitness evaluations.
	count struct {
		sync.Mutex
		n int
	}

	// A free-list used to recycle memory.
	pool = sync.Pool{
		New: func() interface{} {
			return rand.Perm(dim)
		},
	}
)

// The queens type is our genome type. We evolve a permuation of [0,n)
// representing the position of queens on an n x n board
type queens struct {
	gene    []int     // permutation representation of an n-queens solution
	fitness float64   // the negative of the number of conflicts in the solution
	once    sync.Once // used to compute fitness lazily
}

// Close recycles the memory of this genome to be use for new genomes.
func (q *queens) Close() {
	pool.Put(q.gene)
	q.gene = nil
}

// String returns the gene contents and fitness.
func (q *queens) String() string {
	return fmt.Sprintf("%v@%v", q.gene, q.Fitness())
}

// Fitness returns the negative of the number of conflicts in the solution.
// The fitness of a genome is only computed once across all calls to Fitness by
// using a sync.Once.
func (q *queens) Fitness() float64 {
	q.once.Do(func() {
		for i := range q.gene {
			for j, k := 1, i-1; k >= 0; j, k = j+1, k-1 {
				if q.gene[k] == q.gene[i]+j || q.gene[k] == q.gene[i]-j {
					q.fitness--
				}
			}
			for j, k := 1, i+1; k < len(q.gene); j, k = j+1, k+1 {
				if q.gene[k] == q.gene[i]+j || q.gene[k] == q.gene[i]-j {
					q.fitness--
				}
			}
		}
		q.fitness /= 2

		count.Lock()
		count.n++
		count.Unlock()
	})
	return q.fitness
}

// Evolve implements the inner loop of the evolutionary algorithm.
// The population calls the Evolve method of each genome, in parallel. Then,
// each receiver returns a value to replace it in the next generation.
func (q *queens) Evolve(matingPool ...evo.Genome) evo.Genome {
	// Crossover:
	// We're implementing a diffusion model. For each member of the population,
	// we receive a small mating pool containing only our neighbors. We choose
	// a mate using a random binary tournament and create a child with
	// partially mapped crossover.
	mate := sel.BinaryTournament(matingPool...).(*queens)
	child := &queens{gene: pool.Get().([]int)}
	perm.PMX(child.gene, q.gene, mate.gene)

	// Mutation:
	// Perform n random swaps where n is taken from an exponential distribution.
	mutationCount := math.Ceil(rand.ExpFloat64() - 0.5)
	for i := float64(0); i < mutationCount; i++ {
		j := rand.Intn(len(child.gene))
		k := rand.Intn(len(child.gene))
		child.gene[j], child.gene[k] = child.gene[k], child.gene[j]
	}

	// Replacement:
	// Only replace if the child is better or equal.
	if q.Fitness() > child.Fitness() {
		return q
	}
	return child
}

func main() {
	fmt.Println("Dimension:", dim)

	// Setup:
	// We create a random initial population and divide it into islands. Each
	// island is evolved independently. The islands are grouped together into
	// a wrapping population. The wrapper coordiates migrations between the
	// islands according to the delay period.
	init := make([]evo.Genome, size)
	for i := range init {
		init[i] = &queens{gene: pool.Get().([]int)}
	}
	islands := make([]evo.Genome, isl)
	islSize := size / isl
	for i := range islands {
		islands[i] = gen.New(init[i*islSize : (i+1)*islSize])
		islands[i].(evo.Population).Start()
	}
	pop := graph.Ring(islands)
	pop.SetDelay(delay)
	pop.Start()

	// Tear-down:
	// Upon returning, we cleanup our resources and print the solution.
	defer func() {
		pop.Close()
		fmt.Println("\nSolution:", evo.Max(pop))
	}()

	// Run:
	// We continuously poll the population for statistics used in the
	// termination conditions.
	for {
		count.Lock()
		n := count.n
		count.Unlock()
		stats := pop.Stats()

		// "\x1b[2K" is the escape code to clear the line
		fmt.Printf("\x1b[2K\rCount: %7d | Max: %4.0f | Min: %4.0f | SD: %6.6g",
			n,
			stats.Max(),
			stats.Min(),
			stats.StdDeviation())

		// We've found the solution when max is 0
		if stats.Max() == 0 {
			return
		}

		// We've converged once the deviation is less than 0.01
		if stats.StdDeviation() < 1e-2 {
			return
		}

		// Force stop after 2,000,000 fitness evaluations
		if n > 2e6 {
			return
		}

		// var blocker chan struct{}
		// <-blocker
	}
}

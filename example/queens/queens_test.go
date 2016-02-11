package queens

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/cbarrick/evo"
	"github.com/cbarrick/evo/perm"
	"github.com/cbarrick/evo/pop/gen"
	"github.com/cbarrick/evo/pop/graph"
	"github.com/cbarrick/evo/sel"
)

// Tuneables
const (
	dim  = 128     // the dimension of the problem
	size = dim * 4 // the size of the population
	isl  = 4       // the number of islands in which to divide the population

	migration = size / isl / 8  // the size of migrations
	delay     = 1 * time.Second // the delay between migrations
)

// Global objects
var (
	// counts the number of fitness evaluations
	count struct {
		sync.Mutex
		n int
	}
)

// The queens type is our genome. We evolve a permuation of [0,n)
// representing the position of queens on an n x n board
type queens struct {
	gene    []int     // permutation representation of an n-queens solution
	fitness float64   // the negative of the number of conflicts in the solution
	once    sync.Once // used to compute fitness lazily
}

// String returns the gene contents and number of conflicts.
func (q *queens) String() string {
	return fmt.Sprintf("%v@%v", q.gene, -q.Fitness())
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

// Evolution implements the body of the evolution loop.
func Evolution(q evo.Genome, suitors []evo.Genome) evo.Genome {
	// Crossover:
	// We're implementing a diffusion model. For each member of the population,
	// we receive a small mating pool containing only our neighbors. We choose
	// a mate using a random binary tournament and create a child with
	// partially mapped crossover.
	mom := q.(*queens)
	dad := sel.BinaryTournament(suitors...).(*queens)
	child := &queens{gene: make([]int, len(mom.gene))}
	perm.PMX(child.gene, mom.gene, dad.gene)

	// Mutation:
	// Perform n random swaps where n is taken from an exponential distribution.
	// mutationCount := math.Ceil(rand.ExpFloat64() - 0.5)
	for i := float64(0); i < 5; i++ {
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

func TestQueens(t *testing.T) {
	fmt.Printf("Find a solution to %d-queens\n", dim)

	// Setup:
	// We create an initial set of random candidates and divide them into "islands".
	// Each island is evolved independently in a generational population.
	// The islands are then linked together into a graph population with
	seed := make([]evo.Genome, size)
	for i := range seed {
		seed[i] = &queens{gene: perm.New(dim)}
	}
	islands := make([]evo.Genome, isl)
	islSize := size / isl
	for i := range islands {
		var island gen.Population
		island.Evolve(seed[i*islSize:(i+1)*islSize], Evolution)
		islands[i] = &island
	}
	pop := graph.Ring(isl)
	pop.Evolve(islands, gen.Migrate(migration, delay))

	// Continuously print statistics while the optimization runs.
	pop.Poll(0, func() bool {
		count.Lock()
		n := count.n
		count.Unlock()
		stats := pop.Stats()

		// "\x1b[2K" is the xterm escape code to clear the line
		// Because this is a minimization problem, the fitness is negative.
		// Thus we update the statistics accordingly.
		fmt.Printf("\x1b[2K\rCount: %7d | Max: %3.0f | Mean: %3.0f | Min: %3.0f | RSD: %9.2e",
			n,
			-stats.Min(),
			-stats.Mean(),
			-stats.Max(),
			-stats.RSD())

		return false
	})

	// Terminate when we've found the solution (when max is 0)
	pop.Poll(0, func() bool {
		stats := pop.Stats()
		return stats.Max() == 0
	})

	// Terminate if We've converged to a deviation is less than 0.01
	pop.Poll(0, func() bool {
		stats := pop.Stats()
		return stats.SD() < 1e-2
	})

	// Terminate after 2,000,000 fitness evaluations.
	pop.Poll(0, func() bool {
		count.Lock()
		n := count.n
		count.Unlock()
		return n > 2e6
	})

	pop.Wait()
	best := seed[0]
	bestFit := seed[0].Fitness()
	for i := range seed {
		fit := seed[i].Fitness()
		if fit > bestFit {
			best = seed[i]
			bestFit = fit
		}
	}
	fmt.Println("\nSolution:", best)
}

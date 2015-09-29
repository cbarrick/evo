package queens

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/cbarrick/evo"
	"github.com/cbarrick/evo/perm"
	"github.com/cbarrick/evo/pop/graph"
)

// Count counts the number of fitness evaluations.
// It is syncronised because fitness evaluations may happen in parrallel.
var count struct{
	sync.RWMutex
	val int
}

// The queens type is our genome type
type queens struct {
	gene    []int     // permutation representation of n-queens
	fitness float64   // cache of the fitness
	once    sync.Once // used to sync fitness computations
}

// String produces the gene contents and fitness.
// Useful for debugging.
func (q *queens) String() string {
	return fmt.Sprintf("%v@%v", q.gene, q.Fitness())
}

// The genome does not use the close method.
// This could be used to implement genome recycling and reduce allocation.
func (q *queens) Close() {}

// Fitness returns the negative count of conflicts.
// The value is cached so the computation only occurs once.
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

		// we count the number of fitness evaluations
		count.Lock()
		count.val++
		count.Unlock()
	})
	return q.fitness
}

// Cross is the implementation of the inner loop of the GA
//
// Cross is called in-parallel for each position in the population. The receiver
// of the method is the genome currently occupying that position. The genome
// returned will occupy that position next iteration.
func (q *queens) Cross(matingPool ...evo.Genome) evo.Genome {

	// Selection:
	// Select each parent using a simple random binary tournament
	mom := evo.BinaryTournament(matingPool...).(*queens)
	dad := evo.BinaryTournament(matingPool...).(*queens)

	// Crossover:
	// Cycle crossover
	// We randomly decide who is the first and second parent
	child := &queens{gene: perm.CycleX(mom.gene, dad.gene)}

	// Mutation:
	// Perform n random swaps where n is taken from an exponential distribution
	mutationCount := math.Ceil(rand.ExpFloat64() - 0.5)
	for i := float64(0); i < mutationCount; i++ {
		j := rand.Intn(len(child.gene))
		k := rand.Intn(len(child.gene))
		child.gene[j], child.gene[k] = child.gene[k], child.gene[j]
	}

	// Replacement:
	// Only replace if the child is better or equal
	if q.Fitness() > child.Fitness() {
		return q
	}
	return child
}

// Main is the entry point of our program.
//
// We construct an island model population where each island is a diffusion
// population. The islands are arranged in a ring, and the nodes of each
// diffusion population are arranged in a hypercube.
func Main(dim int) {
	if dim <= 0 {
		dim = 256
	}

	// tunables
	var (
		pop   evo.Population
		size  = dim * 2         // the size of the population
		isl   = 8               // the number of islands
		delay = 1 * time.Second // delay between island communication
	)

	fmt.Printf("queens: dimension=%d population=%d\n", dim, size)

	// random initial population
	init := make([]evo.Genome, size)
	for i := range init {
		init[i] = &queens{gene: rand.Perm(dim)}
	}

	// construct the population
	isleSize := size/isl
	islands := make([]evo.Genome, isl)
	for i := range islands {
		islands[i] = graph.Hypercube(init[i*isleSize:(i+1)*isleSize])
	}
	pop = graph.Ring(islands).SetDelay(delay)

	// prints summary stats
	// "\x1b[2K" is the escape code to clear the line
	print := func(v evo.View) {
		count.RLock()
		fmt.Printf("\x1b[2K\rCount: %d | Max: %d | Min: %d | SD: %f",
			count.val,
			int(v.Max().Fitness()),
			int(v.Min().Fitness()),
			v.StdDeviation())
		count.RUnlock()
	}

	// control loop
	// continuously query the population for stats and print them
	// kill the population when done
	for {
		view := pop.View()
		print(view)

		// stop when fitness is 0
		// or we count 1 million fitness computations
		count.RLock()
		if view.Max().Fitness() == 0 || count.val >= 1e6 {
			fmt.Printf("\nSolution: %v\n", view.Max())
			count.RUnlock()
			pop.Close()
			view.Recycle()
			return
		}
		count.RUnlock()

		// recycling the view reduces allocation cost of creating the next view
		view.Recycle()

		// sleep before next poll
		<-time.After(500 * time.Millisecond)
	}
}

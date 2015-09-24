package main

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/cbarrick/evo"
	"github.com/cbarrick/evo/pop/graph"
)

const (
	dim = 256 // the dimension to be solved
)

// queens is our genome type
type queens struct {
	gene    []int     // permutation representation of n-queens
	fitness float64   // cache of the fitness
	once    sync.Once // used to sync fitness computations
}

// The genePool is used to cache allocated but unused genes. By using a
// genePool, we greatly reduce the allocation cost of the program.
var genePool = sync.Pool{
	New: func() interface{} {
		return rand.Perm(dim)
	},
}

// random creates a random n-queens genome.
// We use Fisher-Yates shuffle to produce random permutations.
func random(n int) *queens {
	gene := genePool.Get().([]int)
	for i := dim-1; 1 <= i; i-- {
		j := rand.Intn(i+1)
		gene[i], gene[j] = gene[j], gene[i]
	}
	return &queens{
		gene: gene,
	}
}

// String produces the gene contents
func (q *queens) String() string {
	return fmt.Sprintf("%v@%v", q.gene, q.Fitness())
}

// Cross is the implmentation of the inner loop of the GA
//
// Cross is called in-parallel for each position in the population. The receiver
// of the method is the genome currently occupying that position. The genome
// returned will occupy that position next iteration.
func (q *queens) Cross(matingPool ...evo.Genome) evo.Genome {

	// Selection:
	// TODO: Document
	mom := evo.BinaryTournament(matingPool...).(*queens)
	dad := evo.BinaryTournament(matingPool...).(*queens)

	// Crossover:
	// Take the best child from partially mapped crossover (PMX)
	// We randomly decide who is the first and second parent
	var p1, p2 []int
	if rand.Float64() > 0.5 {
		p1, p2 = mom.gene, dad.gene
	} else {
		p1, p2 = dad.gene, mom.gene
	}
	child := &queens{
		gene: evo.CycleX(p1, p2),
	}

	// Mutation:
	// Perform n random swaps where n is taken from an exponential distribution
	mutationCount := math.Ceil(rand.ExpFloat64() - 0.5)
	for i := float64(0); i < mutationCount; i++ {
		j := rand.Intn(len(child.gene))
		k := rand.Intn(len(child.gene))
		child.gene[j], child.gene[k] = child.gene[k], child.gene[j]
	}

	// Replacement:
	// Only replace if the child is better
	return evo.BinaryTournament(q, child)
}

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
	})
	return q.fitness
}

// Close returns the gene to the genePool
func (q *queens) Close() {
	genePool.Put(q.gene)
}

func main() {
	var (
		pop  evo.Population
		size = 1024 // the size of the population
		isl  = 8    // the number of islands
	)

	fmt.Printf("queens: dimension=%d population=%d\n", dim, size)

	// random initial population
	init := make([]evo.Genome, size)
	for i := range init {
		init[i] = random(dim)
	}

	// construct the population
	isleSize := size/isl
	islands := make([]evo.Genome, isl)
	for i := range islands {
		islands[i] = graph.Hypercube(init[i*isleSize:(i+1)*isleSize])
	}
	pop = graph.Ring(islands).SetDelay(1 * time.Second)

	// control loop
	// continuously query the population for stats and print them
	// reseed the population if it converges early
	// kill the population when done
	for {
		view := pop.View()

		// print stats
		// "\x1b[2K" is the escape code to clear the line
		fmt.Printf("\x1b[2K\r%v", view)

		// stop when fitness is 0
		if view.Max().Fitness() == 0 {
			fmt.Printf("\nSolution: %v\n", view.Max())
			pop.Close()
			view.Recycle()
			return
		}

		// recycling the view reduces allocation cost of creating the next view
		view.Recycle()
	}
}

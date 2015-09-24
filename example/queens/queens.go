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

// We count the number of fitness computations
var count int

// queens is our genome type
type queens struct {
	gene    []int     // permutation representation of n-queens
	fitness float64   // cache of the fitness
	once    sync.Once // used to sync fitness computations
}

// The genePool is used to cache allocated but unused genes. By using a
// genePool, we greatly reduce the allocation cost of the program.
var genePool sync.Pool

// String produces the gene contents and fitness
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
	// Select each parent using a simple random binary tournament
	mom := evo.BinaryTournament(matingPool...).(*queens)
	dad := evo.BinaryTournament(matingPool...).(*queens)

	// Crossover:
	// Cycle crossover
	// We randomly decide who is the first and second parent
	var p1, p2 []int
	if rand.Float64() > 0.5 {
		p1, p2 = mom.gene, dad.gene
	} else {
		p1, p2 = dad.gene, mom.gene
	}
	child := &queens{
		gene: perm.CycleX(p1, p2),
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
	// Only replace if the child is better or equal
	if q.Fitness() > child.Fitness() {
		return q
	}
	return child
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

		// we count the number of fitness evaluations
		count++
	})
	return q.fitness
}

// Close returns the gene to the genePool
func (q *queens) Close() {
	genePool.Put(q.gene)
}

func Main(dim int) {
	// default dimension is 256
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

	// the genePool lets us recycle genes to reduce allocation cost
	genePool = sync.Pool{
		New: func() interface{} {
			return rand.Perm(dim)
		},
	}

	fmt.Printf("queens: dimension=%d population=%d\n", dim, size)

	// random initial population
	// we pull a gene from the genePool and apply Fisher-Yates shuffle
	init := make([]evo.Genome, size)
	for i := range init {
		gene := genePool.Get().([]int)
		for i := dim-1; 1 <= i; i-- {
			j := rand.Intn(i+1)
			gene[i], gene[j] = gene[j], gene[i]
		}
		init[i] = &queens{gene: gene}
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
		fmt.Printf("\x1b[2K\rCount: %d | Max: %d | Min: %d | SD: %f",
			count,
			int(v.Max().Fitness()),
			int(v.Min().Fitness()),
			v.StdDeviation())
	}

	// control loop
	// continuously query the population for stats and print them
	// reseed the population if it converges early
	// kill the population when done
	for {
		view := pop.View()
		print(view)

		// stop when fitness is 0
		// or we count 1 million fitness computations
		if view.Max().Fitness() == 0 || count >= 1000000 {
			fmt.Printf("\nSolution: %v\n", view.Max())
			pop.Close()
			view.Recycle()
			return
		}

		// recycling the view reduces allocation cost of creating the next view
		view.Recycle()

		// sleep before next poll
		<-time.After(1 * time.Second)
	}
}

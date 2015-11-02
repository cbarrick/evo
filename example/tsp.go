package main

import (
	"fmt"
	"math"
	"math/rand"
	"sync"

	"github.com/cbarrick/evo"
	"github.com/cbarrick/evo/perm"
	"github.com/cbarrick/evo/pop/graph"
	"github.com/cbarrick/evo/sel"
)

// Constants
const (
	size = dim * 5     // the size of the population
	dim  = len(cities) // the dimension of the problem
	best = 118282      // shortest known tour of the cities
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

// A city is a coordinate pair giving the position of the city.
type city struct {
	x, y float64
}

// Cities is the list of cities to tour for this example.
var cities = [...]city{
	{9860, 14152},
	{9396, 14616},
	{11252, 14848},
	{11020, 13456},
	{9512, 15776},
	{10788, 13804},
	{10208, 14384},
	{11600, 13456},
	{11252, 14036},
	{10672, 15080},
	{11136, 14152},
	{9860, 13108},
	{10092, 14964},
	{9512, 13340},
	{10556, 13688},
	{9628, 14036},
	{10904, 13108},
	{11368, 12644},
	{11252, 13340},
	{10672, 13340},
	{11020, 13108},
	{11020, 13340},
	{11136, 13572},
	{11020, 13688},
	{8468, 11136},
	{8932, 12064},
	{9512, 12412},
	{7772, 11020},
	{8352, 10672},
	{9164, 12876},
	{9744, 12528},
	{8352, 10324},
	{8236, 11020},
	{8468, 12876},
	{8700, 14036},
	{8932, 13688},
	{9048, 13804},
	{8468, 12296},
	{8352, 12644},
	{8236, 13572},
	{9164, 13340},
	{8004, 12760},
	{8584, 13108},
	{7772, 14732},
	{7540, 15080},
	{7424, 17516},
	{8352, 17052},
	{7540, 16820},
	{7888, 17168},
	{9744, 15196},
	{9164, 14964},
	{9744, 16240},
	{7888, 16936},
	{8236, 15428},
	{9512, 17400},
	{9164, 16008},
	{8700, 15312},
	{11716, 16008},
	{12992, 14964},
	{12412, 14964},
	{12296, 15312},
	{12528, 15196},
	{15312, 6612},
	{11716, 16124},
	{11600, 19720},
	{10324, 17516},
	{12412, 13340},
	{12876, 12180},
	{13688, 10904},
	{13688, 11716},
	{13688, 12528},
	{11484, 13224},
	{12296, 12760},
	{12064, 12528},
	{12644, 10556},
	{11832, 11252},
	{11368, 12296},
	{11136, 11020},
	{10556, 11948},
	{10324, 11716},
	{11484, 9512},
	{11484, 7540},
	{11020, 7424},
	{11484, 9744},
	{16936, 12180},
	{17052, 12064},
	{16936, 11832},
	{17052, 11600},
	{13804, 18792},
	{12064, 14964},
	{12180, 15544},
	{14152, 18908},
	{5104, 14616},
	{6496, 17168},
	{5684, 13224},
	{15660, 10788},
	{5336, 10324},
	{812, 6264},
	{14384, 20184},
	{11252, 15776},
	{9744, 3132},
	{10904, 3480},
	{7308, 14848},
	{16472, 16472},
	{10440, 14036},
	{10672, 13804},
	{1160, 18560},
	{10788, 13572},
	{15660, 11368},
	{15544, 12760},
	{5336, 18908},
	{6264, 19140},
	{11832, 17516},
	{10672, 14152},
	{10208, 15196},
	{12180, 14848},
	{11020, 10208},
	{7656, 17052},
	{16240, 8352},
	{10440, 14732},
	{9164, 15544},
	{8004, 11020},
	{5684, 11948},
	{9512, 16472},
	{13688, 17516},
	{11484, 8468},
	{3248, 14152},
}

// Dist returns the distance between two cities.
func dist(a, b city) float64 {
	return math.Sqrt((a.x-b.x)*(a.x-b.x) + (a.y-b.y)*(a.y-b.y))
}

// The tsp type is our genome type.
type tsp struct {
	gene    []int     // permutation representation of a tour
	fitness float64   // the negative length of the tour
	once    sync.Once // used to compute fitness lazily
}

// String returns the gene contents and fitness.
func (t *tsp) String() string {
	return fmt.Sprintf("%v@%v", t.gene, t.Fitness())
}

// Close recycles the memory of this genome to be use for new genomes.
func (t *tsp) Close() {
	pool.Put(t.gene)
	t.gene = nil
}

// Fitness returns the negative length of the tour represented by a tsp genome.
// The fitness of a genome is only computed once across all calls to Fitness by
// using a sync.Once.
func (t *tsp) Fitness() float64 {
	t.once.Do(func() {
		for i := range t.gene {
			a := cities[t.gene[i]]
			b := cities[t.gene[(i+1)%dim]]
			t.fitness -= dist(a, b)
		}

		count.Lock()
		count.n++
		count.Unlock()
	})
	return t.fitness
}

// Evolve implements the inner loop of the evolutionary algorithm.
// The population calls the Evolve method of each genome, in parallel. Then,
// each receiver returns a value to replace it in the next generation.
func (t *tsp) Evolve(matingPool ...evo.Genome) evo.Genome {
	// Selection:
	// Select each parent using a simple random binary tournament
	mom := sel.BinaryTournament(matingPool...).(*tsp)
	dad := sel.BinaryTournament(matingPool...).(*tsp)

	// Crossover:
	// Edge recombination
	child := &tsp{gene: pool.Get().([]int)}
	perm.EdgeX(child.gene, mom.gene, dad.gene)

	// Mutation:
	// 10% chance to have a random inversion
	if rand.Float32() < 0.1 {
		perm.RandInvert(child.gene)
	}

	// Replacement:
	// Only replace if the child is better or equal
	if t.Fitness() > child.Fitness() {
		return t
	}
	return child
}

func main() {
	// Setup:
	// We create a random initial population
	// and evolve it using a generational model.
	init := make([]evo.Genome, size)
	for i := range init {
		init[i] = &tsp{gene: pool.Get().([]int)}
	}
	pop := graph.Hypercube(init)
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
		fmt.Printf("\x1b[2K\rCount: %7d | Max: %10.2f | Min: %10.2f | SD: %7.2e",
			n,
			stats.Max(),
			stats.Min(),
			stats.StdDeviation())

		// We've found the solution when max is -best
		if stats.Max() >= -best {
			return
		}

		// We've converged once the deviation is less than 1e-12
		if stats.StdDeviation() < 1e-12 {
			return
		}

		// Force stop after 2,000,000 fitness evaluations
		if n > 2e6 {
			return
		}
	}
}

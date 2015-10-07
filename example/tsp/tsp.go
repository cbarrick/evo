package tsp

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/cbarrick/evo"
	"github.com/cbarrick/evo/perm"
	"github.com/cbarrick/evo/pop/gen"
)

// Best is the shortest known tour for the cities below.
const best float64 = 118282

// Count counts the number of fitness evaluations.
// It is syncronised because fitness evaluations may happen in parrallel.
var count struct {
	sync.RWMutex
	val int
}

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

// Dist returns the distance between two cities
func dist(a, b city) float64 {
	return math.Sqrt((a.x-b.x)*(a.x-b.x) + (a.y-b.y)*(a.y-b.y))
}

// The tsp type is our genome type
type tsp struct {
	gene    []int      // permutation representation of traveling salesman
	fitness float64    // cache of the fitness
	once    sync.Once  // used to sync fitness computations
	pool    *sync.Pool // used to recycle genes
}

// String produces the gene contents and fitness.
// Useful for debugging.
func (t *tsp) String() string {
	return fmt.Sprintf("%v@%v", t.gene, t.Fitness())
}

// The genome does not use the close method.
// This could be used to implement genome recycling and reduce allocation.
func (t *tsp) Close() {
	t.pool.Put(t.gene)
	t.gene = nil
}

// Fitness returns the negative length of the tour represented by a tsp genome.
// The value is cached so the computation only occurs once.
func (t *tsp) Fitness() float64 {
	t.once.Do(func() {
		for i := range t.gene {
			a := cities[t.gene[i]]
			b := cities[t.gene[(i+1)%len(t.gene)]]
			t.fitness -= dist(a, b)
		}

		// count this fitness evaluation
		count.Lock()
		count.val++
		count.Unlock()
	})
	return t.fitness
}

// Cross is the implementation of the inner loop of the GA
//
// Cross is called in-parallel for each position in the population. The receiver
// of the method is the genome currently occupying that position. The genome
// returned will occupy that position next iteration.
func (t *tsp) Cross(matingPool ...evo.Genome) evo.Genome {

	// Selection:
	// Select each parent using a simple random binary tournament
	mom := evo.BinaryTournament(matingPool...).(*tsp)
	dad := evo.BinaryTournament(matingPool...).(*tsp)

	// Crossover:
	// Edge recombination
	child := &tsp{gene: mom.pool.Get().([]int), pool: mom.pool}
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

// Main is the entry point of our program.
//
// We construct an island model population where each island is a diffusion
// population. The islands are arranged in a ring, and the nodes of each
// diffusion population are arranged in a hypercube.
func Main() {
	dim := len(cities)

	// tunables
	var (
		pop   evo.Population
		size  = dim * 5 // the size of the population
	)

	fmt.Printf("tsp: dimension=%d population=%d\n", dim, size)

	// pool of genes. lets us reuse genes, reducing allocation costs
	pool := sync.Pool{
		New: func() interface{} {
			return make([]int, dim)
		},
	}

	// random initial population
	init := make([]evo.Genome, size)
	for i := range init {
		init[i] = &tsp{gene: rand.Perm(dim), pool: &pool}
	}

	// construct the population
	pop = gen.New(init)

	// prints summary stats
	// "\x1b[2K" is the escape code to clear the line
	print := func(stats evo.Stats) {
		count.RLock()
		fmt.Printf("\x1b[2K\rCount: %d | Max: %d | Min: %d | SD: %f",
			count.val,
			int(stats.Max()),
			int(stats.Min()),
			stats.StdDeviation())
		count.RUnlock()
	}

	// control loop
	// continuously query the population for stats and print them
	// kill the population when done
	for {
		stats := pop.Stats()
		print(stats)

		// stop when fitness is 0
		// or we count 1 million fitness computations
		count.RLock()
		if stats.Max() == 0 || count.val >= 1e6 {
			pop.Close()
			fmt.Println()
			for i := pop.Iter(); i.Value() != nil; i.Next() {
				if i.Value().Fitness() == 0 {
					fmt.Println(i.Value())
					break
				}
			}
			count.RUnlock()
			return
		}
		count.RUnlock()

		// sleep before next poll
		<-time.After(100 * time.Millisecond)
	}
}

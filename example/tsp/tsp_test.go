package tsp

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
	"testing"

	"github.com/cbarrick/evo"
	"github.com/cbarrick/evo/perm"
	"github.com/cbarrick/evo/pop/graph"
	"github.com/cbarrick/evo/sel"
)

// Constants
const (
	dim  = len(cities) // the dimension of the problem
	size = 256         // the size of the population
	stop = 2e6         // terminate after this number of fitness evalutations
)

// Global objects
var (
	// The evolutionary loop managed by the population
	pop evo.Population

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

// Dist returns the pseudo-euclidian distance between two cities. This is the
// distance function used in the literature on this problem instance.
// http://comopt.ifi.uni-heidelberg.de/software/TSPLIB95/DOC.PS
func dist(a, b city) float64 {
	xd := a.x - b.x
	yd := a.y - b.y
	d := math.Sqrt(xd*xd/10 + yd*yd/10)
	return math.Ceil(d)
}

// The tsp type is our genome type.
type tsp struct {
	gene    []int     // permutation representation of a tour
	fitness float64   // the negative length of the tour
	once    sync.Once // used to compute fitness lazily
}

// String returns the gene contents and length of the tour.
func (t *tsp) String() string {
	return fmt.Sprintf("%v@%v", t.gene, -t.Fitness())
}

// Close recycles the memory of this genome to be use for new genomes.
func (t *tsp) Close() {
	pool.Put(t.gene)
	t.gene = nil
}

// Fitness returns the negative length of the tour represented by a tsp genome.
// The fitness is negative because TSP is a minimization problem, but the Evo
// API is phrased in terms of maximization. As a consequence, fitness statistics
// (i.e. the result of pop.Stats()) are also negative: the shortest know path
// would be -stats.Max().
func (t *tsp) Fitness() float64 {
	t.once.Do(func() {
		t.fitness = 0
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

// TwoOpt performs a 2-opt local search for improvement of the gene. The first
// edge is selected at random and inversions between all other edges are
// evaluated in random order. Even if an improvement is not found, the gene will
// be rotated by an uniform-random amount. We use this search as a form of
// mutation.
func (t *tsp) TwoOpt() {
	t.once = sync.Once{}
	perm.Rotate(t.gene, rand.Intn(dim))
	for _, i := range rand.Perm(dim) {
		if i < 2 {
			continue
		}
		a := cities[t.gene[0]]
		b := cities[t.gene[i-1]]
		y := cities[t.gene[i]]
		z := cities[t.gene[dim-1]]
		before := dist(b, y) + dist(z, a)
		after := dist(a, y) + dist(z, b)
		if after < before {
			perm.Reverse(t.gene[:i])
			return
		}
	}
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
	// There is an n% chance for the gene to have n random swaps
	// and an n% chance to undergo n steps of a greedy 2-opt hillclimber
	for rand.Float64() < 0.1 {
		perm.RandSwap(child.gene)
	}
	for rand.Float64() < 0.1 {
		child.TwoOpt()
	}

	// Replacement:
	// Only replace if the child is better or equal
	if t.Fitness() > child.Fitness() {
		return t
	}
	return child
}

func TestTSP(t *testing.T) {
	fmt.Println("Minimize tour of US capitals - optimal is", best)

	// Setup:
	// We create a random initial population
	// and evolve it using a generational model.
	init := make([]evo.Genome, size)
	for i := range init {
		init[i] = &tsp{gene: pool.Get().([]int)}
	}
	pop = graph.Hypercube(init)
	pop.Start()

	// Tear-down:
	// Upon returning, we cleanup our resources and print the solution.
	defer func() {
		pop.Close()
		fmt.Println("\nTour:", evo.Max(pop))
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
		// The fitness of minimization problems is negative
		fmt.Printf("\x1b[2K\rCount: %7d | Max: %6.0f | Mean: %6.0f | Min: %6.0f | RSD: %7.2e",
			n,
			-stats.Min(),
			-stats.Mean(),
			-stats.Max(),
			stats.RSD())

		// Stop when we get close. Finding the true minimum could take a while.
		if -stats.Max() < best*1.1 {
			return
		}

		// Force stop after some number fitness evaluations
		if n > stop {
			t.Fail()
			return
		}
	}
}

// Best is the minimum tour of the cities.
const best = 10628

// Cities is a list of the capitals of the 48 contiguous American states. The
// minimum tour length is 10628. This is dataset ATT48 from TSPLIB, a collection
// of traveling salesman problem datasets maintained by Dr. Gerhard Reinelt:
// "http://comopt.ifi.uni-heidelberg.de/software/TSPLIB95/"
var cities = [48]city{
	{6734, 1453},
	{2233, 10},
	{5530, 1424},
	{401, 841},
	{3082, 1644},
	{7608, 4458},
	{7573, 3716},
	{7265, 1268},
	{6898, 1885},
	{1112, 2049},
	{5468, 2606},
	{5989, 2873},
	{4706, 2674},
	{4612, 2035},
	{6347, 2683},
	{6107, 669},
	{7611, 5184},
	{7462, 3590},
	{7732, 4723},
	{5900, 3561},
	{4483, 3369},
	{6101, 1110},
	{5199, 2182},
	{1633, 2809},
	{4307, 2322},
	{675, 1006},
	{7555, 4819},
	{7541, 3981},
	{3177, 756},
	{7352, 4506},
	{7545, 2801},
	{3245, 3305},
	{6426, 3173},
	{4608, 1198},
	{23, 2216},
	{7248, 3779},
	{7762, 4595},
	{7392, 2244},
	{3484, 2829},
	{6271, 2135},
	{4985, 140},
	{1916, 1569},
	{7280, 4899},
	{7509, 3239},
	{10, 2676},
	{6807, 2993},
	{5185, 3258},
	{3023, 1942},
}

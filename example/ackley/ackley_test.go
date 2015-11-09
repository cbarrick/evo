package ackley

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
	"testing"

	"github.com/cbarrick/evo"
	"github.com/cbarrick/evo/pop/gen"
	"github.com/cbarrick/evo/real"
	"github.com/cbarrick/evo/sel"
)

// Tuneables
const (
	dim       = 30    // Dimension of the problem.
	bounds    = 30    // Bounds of object parameters.
	precision = 1e-16 // Desired precision.
)

// Global objects
var (
	// Count of the number of fitness evaluations.
	count struct {
		sync.Mutex
		n int
	}

	// Each of the 40 members of the population generates 7 children and adds
	// them to this pool. This pool returns to each member a different one of
	// most fit to be their replacement in the next generation.
	selector = sel.ElitePool(40, 280)

	// A free-list used to recycle memory.
	vectors = sync.Pool{
		New: func() interface{} {
			return make(real.Vector, dim)
		},
	}
)

// The ackley type specifies our genome. We evolve a real-valued vector that
// optimizes the ackley function. Each genome also contains a vector of strategy
// parameters used with a self-adaptive evolution strategy.
type ackley struct {
	gene  real.Vector // the object vector to optimize
	steps real.Vector // strategy parameters for mutation
	fit   float64     // the ackley function of the gene
	once  sync.Once   // used to compute fitness lazily
}

// Close recycles the memory of this genome to be use for new genomes.
func (ack *ackley) Close() {
	vectors.Put(ack.gene)
	vectors.Put(ack.steps)
}

// Returns the fitness as a string.
func (ack *ackley) String() string {
	return fmt.Sprint(-ack.Fitness())
}

// Fitness returns the ackley function of the gene. We are trying to solve a
// minimization problem, so we return the negative of the traditional formula.
// The fitness of a genome is only computed once across all calls to Fitness by
// using a sync.Once.
func (ack *ackley) Fitness() float64 {
	const a, b = 20, 0.2
	ack.once.Do(func() {
		var sum1, sum2 float64
		n := float64(dim)
		for _, x := range ack.gene {
			sum1 += x * x
			sum2 += math.Cos(2 * math.Pi * x)
		}

		ack.fit -= a
		ack.fit *= math.Exp(-b * math.Sqrt(sum1/n))
		ack.fit -= math.Exp(sum2 / n)
		ack.fit += a
		ack.fit += math.E
		ack.fit *= -1

		count.Lock()
		count.n++
		count.Unlock()
	})
	return ack.fit
}

// Evolve implements the inner loop of the evolutionary algorithm.
// The population calls the Evolve method of each genome, in parallel. Then,
// each receiver returns a value to replace it in the next generation. A global
// selector object synchronises replacement among the parallel parents.
func (ack *ackley) Evolve(suitors ...evo.Genome) evo.Genome {
	for i := 0; i < 7; i++ {
		// Creation:
		// We create the child genome from recycled memory when we can.
		var child ackley
		child.gene = vectors.Get().(real.Vector)
		child.steps = vectors.Get().(real.Vector)

		// Crossover:
		// Select two parents at random.
		// Uniform crossover of object parameters.
		// Arithmetic crossover of strategy parameters.
		mom := suitors[rand.Intn(len(suitors))].(*ackley)
		dad := suitors[rand.Intn(len(suitors))].(*ackley)
		real.UniformX(child.gene, mom.gene, dad.gene)
		real.ArithX(1, child.steps, mom.steps, dad.steps)

		// Mutation: Evolution Strategy
		// Lognormal scaling of strategy parameters.
		// Gausian perturbation of object parameters.
		child.steps.Adapt()
		child.steps.LowBound(precision)
		child.gene.Step(child.steps)
		child.gene.HighBound(bounds)
		child.gene.LowBound(-bounds)

		// Replacement: (40,280)
		// Each child is added to the global selection pool.
		selector.Put(&child)
	}

	// Finally, block until all parallel calls have added their children to the
	// selection pool and return one of the selected replacements.
	return selector.Get()
}

func TestAckley(t *testing.T) {
	fmt.Printf("Minimize the Ackley function with n=%d\n", dim)

	// Setup:
	// We initialize a set of 40 random solutions,
	// then add them to a generational population.
	init := make([]evo.Genome, 40)
	for i := range init {
		init[i] = &ackley{
			gene:  real.Random(dim, 30),
			steps: real.Random(dim, 1),
		}
	}
	pop := gen.New(init)
	pop.Start()

	// Tear-down:
	// Upon returning, we cleanup our resources and print the solution.
	defer func() {
		pop.Close()
		selector.Close()
		fmt.Println("\nSolution:", evo.Max(pop))
	}()

	// Run:
	// We continuously poll the population for statistics and terminate when we
	// have a solution or after 200,000 evaluations.
	for {
		count.Lock()
		n := count.n
		count.Unlock()
		stats := pop.Stats()

		// "\x1b[2K" is the escape code to clear the line
		// The fitness of minimization problems is negative
		fmt.Printf("\x1b[2K\rCount: %7d | Max: %8.3g | Mean: %8.3g | Min: %8.3g | RSD: %9.2e",
			n,
			-stats.Min(),
			-stats.Mean(),
			-stats.Max(),
			stats.RSD())

		// We've converged once the deviation is within the precision
		if stats.SD() < precision {
			return
		}

		// Force stop after 200,000 fitness evaluations
		if n > 200000 {
			return
		}
	}
}

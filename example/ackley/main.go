package main

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/cbarrick/evo"
	"github.com/cbarrick/evo/pop/gen"
	"github.com/cbarrick/evo/real"
	"github.com/cbarrick/evo/sel"
)

// Tuneables
const (
	dim       = 30   // Dimension of the problem.
	bounds    = 30   // Bounds of object parameters.
	precision = 1e-6 // Desired precision.
)

// Global objects
var (
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
	gene  real.Vector // The object vector to optimize
	steps real.Vector // Strategy parameters for mutation
}

// When a genome is garbage collected, we recycle its vectors for new genomes.
func (ack *ackley) Close() {
	vectors.Put(ack.gene)
	vectors.Put(ack.steps)
}

// Returns the fitness as a string.
// In our case, we only care about the optimum value, not the parameters used
// to get there. Obviously you can/should return more details as needed.
func (ack *ackley) String() string {
	return fmt.Sprint(ack.Fitness())
}

// The fitness being maximized is the ackley function. Technically we care
// about the minimum value of the ackley function, so we maximize the negative.
// Fitness evaluation can be expensive, so we use a sync.Once to illistrate
// caching of the fitness. Caching isn't important for this application, but it
// can be when the fitness function is more expensive.
func (ack *ackley) Fitness() (f float64) {
	const a, b = 20, 0.2
	var sum1, sum2 float64
	n := float64(dim)
	for _, x := range ack.gene {
		sum1 += x * x
		sum2 += math.Cos(2 * math.Pi * x)
	}

	f -= a
	f *= math.Exp(-b * math.Sqrt(sum1/n))
	f -= math.Exp(sum2 / n)
	f += a
	f += math.E
	f *= -1
	return f
}

// Evolve implements the inner loop of the evolutionary algorithm. It is called
// in parallel for each member of the population, and the genome returned
// replaces the method receiver in the next generation. We use a 40 member
// population and generate 7 new competing children per call, then return one
// of the best.
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
		child.steps.LowPass(precision)
		child.gene.Step(child.steps)
		child.gene.HighPass(bounds)
		child.gene.LowPass(-bounds)

		// Replacement: (40,280)
		// Each child is added to a shared selection pool, defined globally.
		// The pool decides which children should be used in the next generation.
		selector.Put(&child)
	}

	// This blocks until all parallel calls to Evolve have added their children
	// to the pool. Then it returns one of the most fit to be the replacement.
	return selector.Get()
}

func main() {
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

	// Termination:
	// Stop when we reach the desired precision or after some timeout.
	timeout := time.After(5 * time.Second)
	defer func() {
		pop.Close()
		selector.Close()
		fmt.Println("\nSolution:", evo.Max(pop))
	}()
	for {
		select {
		case <-timeout:
			return
		default:
			stats := pop.Stats()
			// "\x1b[2K" is the escape code to clear the line
			fmt.Printf("\x1b[2K\rMax: %f | Min: %f | SD: %f",
				stats.Max(),
				stats.Min(),
				stats.StdDeviation())
			if stats.StdDeviation() < precision {
				return
			}
		}
	}
}

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

	// Max run time.
	timeout = 100 * time.Second
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

type ackley struct {
	gene  real.Vector
	steps real.Vector
}

func (ack *ackley) Close() {
	vectors.Put(ack.gene)
	vectors.Put(ack.steps)
}

func (ack *ackley) String() string {
	return fmt.Sprint(ack.Fitness())
}

func (ack *ackley) Fitness() (f float64) {
	var a, b float64
	a = 20
	b = 0.2

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

func (ack *ackley) Evolve(suitors ...evo.Genome) evo.Genome {
	// Replacement: (40,280)
	// Evolve is called in parallel for each of the 40 members of the population.
	// Since we want to generate 280 children, we generate 7 children per call.
	for i := 0; i < 7; i++ {

		// Creation:
		// We create the child genome from recycled memory as best we can.
		var child ackley
		child.gene = vectors.Get().(real.Vector)
		child.steps = vectors.Get().(real.Vector)

		// Crossover:
		// Select two parents at random.
		// Uniform crossover of object vars.
		// Arithmetic crossover of strategy vars
		mom := suitors[rand.Intn(len(suitors))].(*ackley)
		dad := suitors[rand.Intn(len(suitors))].(*ackley)
		real.UniformX(child.gene, mom.gene, dad.gene)
		real.ArithX(1, child.steps, mom.steps, dad.steps)

		// Mutation:
		// Lognormal scaling of strategy vars
		// Gausian perturbation of object vars
		child.steps.Adapt()
		child.steps.LowPass(precision)
		child.gene.Step(child.steps)
		child.gene.HighPass(bounds)
		child.gene.LowPass(-bounds)

		// Replacement: (40,280)
		// Each child is added to a selection pool, defined globally.
		// After all children are generated, we return one child from the pool.
		// selector.Get() may block until enough children have been put.
		selector.Put(&child)
	}
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
	// Stop when we reach the desired precision or after the timeout.
	timer := time.After(timeout)
	defer func() {
		pop.Close()
		selector.Close()
		fmt.Println("\nSolution:", Max(pop))
	}()
	for {
		select {
		case <-timer:
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

// TODO: put this somewhere in the library
func Max(pop evo.Population) evo.Genome {
	var val, best evo.Genome
	var fit, bestfit = float64(0), math.Inf(-1)
	for i := pop.Iter(); i.Value() != nil; i.Next() {
		val = i.Value()
		fit = val.Fitness()
		if fit > bestfit {
			best = val
			bestfit = fit
		}
	}
	return best
}

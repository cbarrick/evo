// Package gen provides a traditional generational population.
//
// Generational populations evolve their genomes in generations. Each genome
// is given the entire population as suitors to evolve the next generation. Once
// the new generation is constructed, the old generation is replaced. Each
// genome is evolved in parallel, similar to the textbook master-slave
// parallelism.
package gen

import (
	"math/rand"
	"sync"
	"time"

	"github.com/cbarrick/evo"
)

type Population struct {
	members []evo.Genome        // the individuals, not safe to touch while running
	getc    chan chan int       // used to access members while running
	setc    chan chan int       // used to mutate members while running
	valuec  chan evo.Genome     // sends/receives genomes for get/set
	statsc  chan chan evo.Stats // used to get stats while running
	stopc   chan chan struct{}  // used to stop the goroutine
}

// Evolve initiates the optimization in a separate goroutine.
func (pop *Population) Evolve(members []evo.Genome, body evo.EvolveFn) {
	pop.members = members
	pop.statsc = make(chan chan evo.Stats)
	pop.setc = make(chan chan int)
	pop.getc = make(chan chan int)
	pop.valuec = make(chan evo.Genome)
	pop.stopc = make(chan chan struct{})
	go run(*pop, body)
}

// Stop terminates the evolution loop.
func (pop *Population) Stop() {
	ch := make(chan struct{})
	pop.stopc <- ch
	<-ch
	close(pop.statsc)
	close(pop.setc)
	close(pop.getc)
	close(pop.valuec)
	close(pop.stopc)
}

// Stats returns statistics on the fitness of genomes in the population.
func (pop *Population) Stats() (s evo.Stats) {
	statsc := <-pop.statsc
	if statsc == nil {
		for i := range pop.members {
			s = s.Put(pop.members[i].Fitness())
		}
		return s
	}
	return <-statsc
}

// Fitness returns the maximum fitness within the population.
func (pop *Population) Fitness() float64 {
	return pop.Stats().Max()
}

// get returns the ith member of the population.
func (pop *Population) get(i int) (val evo.Genome) {
	getter := <-pop.getc
	if getter == nil {
		val = pop.members[i]
	} else {
		getter <- i
		val = <-pop.valuec
	}
	return val
}

// set sets the ith member of the population.
func (pop *Population) set(i int, val evo.Genome) {
	setter := <-pop.setc
	if setter == nil {
		pop.members[i] = val
	} else {
		setter <- i
		pop.valuec <- val
	}
}

// Migrate returns an EvolveFn for using generational populations as genomes.
// The returned migration function exchanges n individuals between the target
// population and one neighboring population. Migration can be slowed down by
// a delay period that occurs before the migration is performed.
//
// The returned migration function can be used to implement an island population
// model where the individuals are divided between some number of generational
// populations, which are themselves linked together into a graph population.
// The generational populations evolve using the user's EvolveFn while the graph
// population evolves using a migration function.
//
//     var evolution evo.EvolveFn    // the body of the evolution
//     var seed []evo.Genome         // the initial solutions
//     var islands []evo.Genome      // the islands
//     n := len(seed) / len(islands) // number of solutions per island
//
//     for i := range islands {
//     	var island gen.Population
//     	island.Evolve(seed[i*n:(i+1)*n], evolution)
//     	islands[i] = &island
//     }
//     pop := graph.Ring(len(islands))
//     pop.Evolve(islands, gen.Migrate(5, 1*time.Second))
func Migrate(n int, delay time.Duration) evo.EvolveFn {
	return func(current evo.Genome, suitors []evo.Genome) evo.Genome {
		<-time.After(delay)
		var a, b *Population
		a = current.(*Population)
		for b = a; b == a; {
			b = suitors[rand.Intn(len(suitors))].(*Population)
		}
		for i := 0; i < n; i++ {
			ai := rand.Intn(len(a.members))
			bi := rand.Intn(len(b.members))
			av := a.get(ai)
			bv := b.get(bi)
			a.set(ai, bv)
			b.set(bi, av)
		}
		return current
	}
}

// run implements the main goroutine.
func run(pop Population, body evo.EvolveFn) {
	var (
		// drives the main loop
		loop = make(chan struct{}, 1)

		// receives the results of evolutions
		nextgen = make(chan evo.Genome, len(pop.members))

		// synchronizes pending evolutions
		pending sync.WaitGroup

		// used to access/mutate pop.members
		getter = make(chan int)
		setter = make(chan int)
		statsc = make(chan evo.Stats)
	)

	for i := range pop.members {
		nextgen <- pop.members[i]
	}
	loop <- struct{}{}

	for {
		select {
		case <-loop:
			for i := range pop.members {
				pop.members[i] = <-nextgen
			}
			pending.Add(len(pop.members))
			for i := range pop.members {
				val := pop.members[i]
				go func() {
					nextgen <- body(val, pop.members)
					pending.Done()
				}()
			}
			go func() {
				pending.Wait()
				loop <- struct{}{}
			}()

		case pop.getc <- getter:
			i := <-getter
			pop.valuec <- pop.members[i]

		case pop.setc <- setter:
			i := <-setter
			pop.members[i] = <-pop.valuec

		case pop.statsc <- statsc:
			var s evo.Stats
			for i := range pop.members {
				s = s.Put(pop.members[i].Fitness())
			}
			statsc <- s

		case ch := <-pop.stopc:
			pending.Wait()
			for i := range pop.members {
				if subpop, ok := pop.members[i].(evo.Population); ok {
					subpop.Stop()
				}
			}
			ch <- struct{}{}
			return
		}
	}
}

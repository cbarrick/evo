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
	"runtime"
	"time"

	"github.com/cbarrick/evo"
)

// New constructs a generational population starting from the given members.
func New(members []evo.Genome) evo.Population {
	pop := new(population)
	pop.members = members
	return pop
}

type population struct {
	running  bool
	members  []evo.Genome
	membersc chan []evo.Genome
	delay    time.Duration
	delayc   chan time.Duration
	updates  chan evo.Genome
	migrate  chan chan evo.Genome
	closec   chan chan struct{}
}

// Start initiates the evolution loop in a separate goroutine.
func (pop *population) Start() {
	if !pop.running {
		pop.membersc = make(chan []evo.Genome)
		pop.delayc = make(chan time.Duration)
		pop.updates = make(chan evo.Genome, len(pop.members))
		pop.migrate = make(chan chan evo.Genome)
		pop.closec = make(chan chan struct{})
		pop.running = true
		go pop.run()
	}
}

// SetDelay sets a delay between each iteration of the evolution loop.
func (pop *population) SetDelay(d time.Duration) {
	if pop.running {
		pop.delayc <- d
	} else {
		pop.delay = d
	}
}

// Close terminates the evolution loop.
func (pop *population) Close() {
	if pop.running {
		ch := make(chan struct{})
		pop.closec <- ch
		<-ch
		close(pop.membersc)
		close(pop.delayc)
		close(pop.updates)
		close(pop.migrate)
		close(pop.closec)
		pop.running = false
	}
}

// Iter returns an iterator ranging over the values of the population.
func (pop *population) Iter() evo.Iterator {
	return iterate(pop)
}

// Stats returns statistics on the fitness of genomes in the population.
func (pop *population) Stats() (s evo.Stats) {
	for i := pop.Iter(); i.Value() != nil; i.Next() {
		s = s.Put(i.Value().Fitness())
	}
	return s
}

// Fitness returns the maximum fitness within the population.
func (pop *population) Fitness() float64 {
	return pop.Stats().Max()
}

// Evolve performs a random migration between this population and a random suiter.
func (pop *population) Evolve(suiters ...evo.Genome) evo.Genome {
	var other = pop
	for other == pop {
		other = suiters[rand.Intn(len(suiters))].(*population)
	}
	chA := make(chan evo.Genome)
	chB := make(chan evo.Genome)
	pop.migrate <- chA
	other.migrate <- chB
	a := <-chA
	b := <-chB
	chA <- b
	chB <- a
	return pop
}

// The main goroutine.
func (pop *population) run() {
	var (
		// receives when the next iteration should start
		mate <-chan time.Time

		// holds the next generation as it is being built
		nextgen = make([]evo.Genome, len(pop.members))

		// the number of updates until the next generation is complete
		wait int
	)

	for i := range pop.members {
		runtime.SetFinalizer(pop.members[i], nil)
		runtime.SetFinalizer(pop.members[i], func(val evo.Genome) {
			val.Close()
		})
	}

	cross := func() {
		wait = 0
		for i := range pop.members {
			wait++
			go func(i int, members []evo.Genome) {
				pop.updates <- members[i].Evolve(members...)
			}(i, pop.members)
		}
	}
	cross()

	for {
		select {

		case pop.delay = <-pop.delayc:

		case pop.membersc <- pop.members:
			memcopy := make([]evo.Genome, len(pop.members))
			copy(memcopy, pop.members)
			pop.members = memcopy

		case <-mate:
			cross()

		case ch := <-pop.migrate:
			go func(ch chan evo.Genome) {
				ch <- <-pop.updates
				pop.updates <- <-ch
			}(ch)

		case val := <-pop.updates:
			runtime.SetFinalizer(val, nil)
			runtime.SetFinalizer(val, func(val evo.Genome) {
				val.Close()
			})

			wait--
			nextgen[wait] = val

			if wait == 0 {
				pop.members, nextgen = nextgen, pop.members
				for i := range nextgen {
					nextgen[i] = nil
				}
				mate = time.After(pop.delay)
			}

		case ch := <-pop.closec:
			for 0 < wait {
				<-pop.updates
				wait--
			}
			ch <- struct{}{}
			return

		}
	}
}

// Iterator
// -------------------------

type iter struct {
	sub  evo.Iterator
	idx  int
	vals []evo.Genome
}

func iterate(pop *population) evo.Iterator {
	var it iter
	if vals, ok := <-pop.membersc; ok {
		it.vals = vals
	} else {
		it.vals = pop.members
	}
	if pop, ok := it.vals[it.idx].(evo.Population); ok {
		it.sub = pop.Iter()
	}
	return &it
}

func (it *iter) Value() evo.Genome {
	if it.sub != nil {
		return it.sub.Value()
	}
	if it.idx == len(it.vals) {
		return nil
	}
	return it.vals[it.idx]
}

func (it *iter) Next() {
	switch {
	case it.sub != nil:
		it.sub.Next()
		if it.sub.Value() != nil {
			break
		}
		it.sub = nil
		fallthrough
	default:
		it.idx++
		if it.idx < len(it.vals) {
			if pop, ok := it.vals[it.idx].(evo.Population); ok {
				it.sub = pop.Iter()
			}
		}
	}
}

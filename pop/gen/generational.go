// Package gen provides a traditional generational population.
//
// Generational populations evolve their genomes in generations. Each genome
// is given the entire population as suitors to evolve the next generation. Once
// the new generation is constructed, the old generation is replaced. Each
// genome is evolved in parallel, similar to the textbook master-slave
// parallelism.
package gen

import (
	"runtime"
	"time"

	"github.com/cbarrick/evo"
)

// Population
// -------------------------

type population struct {
	members  []evo.Genome
	membersc chan []evo.Genome
	updates  chan evo.Genome
	migrate  chan chan evo.Genome
	closec   chan chan struct{}
	closed   bool
}

func (pop *population) init(members []evo.Genome) {
	pop.members = members
	pop.membersc = make(chan []evo.Genome)
	pop.updates = make(chan evo.Genome, len(pop.members))
	pop.migrate = make(chan chan evo.Genome)
	pop.closec = make(chan chan struct{})
}

func (pop *population) run() {
	var (
		delay   time.Duration
		mate    = time.After(delay)
		nextgen = make([]evo.Genome, len(pop.members))
		pos     = 0
	)

	for i := range pop.members {
		runtime.SetFinalizer(pop.members[i], nil)
		runtime.SetFinalizer(pop.members[i], func(val evo.Genome) {
			val.Close()
		})
	}

	for {
		select {

		case pop.membersc <- pop.members:
			memcopy := make([]evo.Genome, len(pop.members))
			copy(memcopy, pop.members)
			pop.members = memcopy

		case <-mate:
			for i := range pop.members {
				go func(i int, members []evo.Genome) {
					pop.updates <- members[i].Evolve(members...)
				}(i, pop.members)
			}

		case val := <-pop.updates:
			runtime.SetFinalizer(val, nil)
			runtime.SetFinalizer(val, func(val evo.Genome) {
				val.Close()
			})
			nextgen[pos] = val
			pos++
			if pos == len(nextgen) {
				pop.members, nextgen = nextgen, pop.members
				for i := range nextgen {
					nextgen[i] = nil
				}
				pos = 0
				mate = time.After(delay)
			}

		case ch := <-pop.closec:
			close(pop.membersc)
			pop.closed = true
			ch <- struct{}{}
			return

		}
	}
}

// Iter returns an iterator ranging over the values of the population.
func (pop *population) Iter() evo.Iterator {
	return iterate(pop)
}

// Stats returns statistics on the fitness of genomes in the population.
func (pop *population) Stats() (s evo.Stats) {
	for i := pop.Iter(); i.Value() != nil; i.Next() {
		s = s.Insert(i.Value().Fitness())
	}
	return s
}

// Close terminates the evolutionary algorithm.
func (pop *population) Close() {
	ch := make(chan struct{})
	pop.closec <- ch
	<-ch
	return
}

// Fitness returns the maximum fitness within the population.
func (pop *population) Fitness() float64 {
	return pop.Stats().Max()
}

// Evolve performs a random migration between this population and a random suiter.
func (pop *population) Evolve(suiters ...evo.Genome) evo.Genome {
	panic("Evolve not yet implemented on generational populations")
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

// New returns a generational population initially containing the given members.
func New(members []evo.Genome) evo.Population {
	pop := new(population)
	pop.init(members)
	go pop.run()
	return pop
}

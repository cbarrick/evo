// Package gen implements a generational population.
//
// Generational populations are used to implement OG genetic algorithms. Each
// successive generation is created in its entirety before starting the next
// generation.
package gen

import (
	"math/rand"
	"sync"

	"github.com/cbarrick/evo"
)

type population struct {
	members []evo.Genome

	com struct {
		members chan []evo.Genome // read members safely
		close   chan bool         // stop the generational loop
		inject  chan evo.Genome   // inject a genome into a random position
	}
}

// loop implements the generational loop
func (p *population) loop() {
	var inject evo.Genome

	next := make([]evo.Genome, len(p.members))
	ready := make(chan bool)

	// mate performs one iteration of the loop,
	// and sends on the ready channel when done.
	mate := func(members []evo.Genome) {
		var wg sync.WaitGroup
		wg.Add(len(members))
		for i := range members {
			go func(i int) {
				suiters := make([]evo.Genome, len(members)-1)
				copy(suiters[:i], members[:i])
				copy(suiters[i:], members[i+1:])
				next[i] = members[i].Cross(suiters...)
				wg.Done()
			}(i)
		}
		wg.Wait()
		ready <- true
	}

	// turnover swaps in the next generation
	// and performs an injection if needed
	turnover := func() {
		if inject != nil {
			i := rand.Intn(len(next))
			if inject != next[i] {
				next[i].Close()
				next[i] = inject
			}
			inject = nil
		}
		for i := range p.members {
			if next[i] != p.members[i] {
				p.members[i].Close()
			}
			p.members[i] = nil
		}
		p.members, next = next, p.members
	}

	// start the initial iteration
	go mate(p.members)

	for {
		select {
		// inject a genome into a random position
		// the injection will occur during the next turnover
		case inject = <-p.com.inject:
			break

		// read p.members safely
		case p.com.members <- p.members:
			dup := make([]evo.Genome, len(p.members))
			copy(dup, p.members)
			p.members = dup

		// the current iteration is done
		// turnover and start the next iteration
		case <-ready:
			turnover()
			go mate(p.members)

		// close
		// wait for the final iteration, turnover, then exit
		case x := <-p.com.close:
			if x == true {
				<-ready
				turnover()
				p.com.close <- true
				close(p.com.members)
				close(p.com.close)
				return
			}
		}
	}
}

// Close stops the main loop.
func (p *population) Close() {
	p.com.close <- true
	<-p.com.close
	return
}

// View constructs a view of genomes in the population.
func (p *population) View() evo.View {
	members := <-p.com.members
	if members == nil {
		members = make([]evo.Genome, len(p.members))
		copy(members, p.members)
	}
	return evo.NewView(members...)
}

// Fitness returns the maximum fitness within the population.
func (p *population) Fitness() (f float64) {
	v := p.View()
	f = v.Max().Fitness()
	v.Recycle()
	return f
}

// Cross injects the best genome of a random suiter into a random slot in the
// population.
func (p *population) Cross(suiters ...evo.Genome) evo.Genome {
	q := suiters[rand.Intn(len(suiters))].(*population)
	v := q.View()
	p.com.inject <- v.Max()
	v.Recycle()
	return p
}

// New starts a new generational genetic algorithm with the given members.
func New(members ...evo.Genome) evo.Population {
	var p population
	p.members = members
	p.com.members = make(chan []evo.Genome)
	p.com.close = make(chan bool)
	p.com.inject = make(chan evo.Genome)
	go p.loop()
	return &p
}

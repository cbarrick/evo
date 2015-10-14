package sel

import (
	"github.com/cbarrick/evo"
)

// Pool selectors allow many goroutines to contribute competitors to a selection
// process. Pool selectors reset after each competition and must be closed when
// they are no longer needed.
type Pool struct {
	in    chan evo.Genome
	out   chan evo.Genome
	close chan chan struct{}
}

// Put adds a competitor to the pool.
// Put blocks until all winners of the previous competition have been retrieved.
func (p Pool) Put(val evo.Genome) {
	p.in <- val
}

// Get retrieves a winner from the most current competition.
// Get blocks until all competitors have been added.
func (p Pool) Get() (val evo.Genome) {
	val = <-p.out
	return val
}

// Close stops the pool selector.
func (p Pool) Close() {
	ch := make(chan struct{})
	p.close <- ch
	<-ch
}

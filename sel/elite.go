package sel

import (
	"sort"

	"github.com/cbarrick/evo"
)

// An elcomp competes in an elite tournament.
type elcomp struct {
	evo.Genome
	fit float64
}

// Elcomps implements the sort interface in descending order by fitness.
type elcomps []elcomp

func (h elcomps) Len() int           { return len(h) }
func (h elcomps) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h elcomps) Less(i, j int) bool { return h[i].fit > h[j].fit }

// Sort (re)evaluates the fitness of each competitor and sorts them.
// Fitness evaluation is probably expensive, so we do it in parallel.
func (pool elcomps) sort() {
	done := make(chan struct{})
	for i := range pool {
		go func(i int) {
			pool[i].fit = pool[i].Genome.Fitness()
			done <- struct{}{}
		}(i)
	}
	for _ = range pool {
		<-done
	}
	sort.Sort(pool)
}

// Elite returns the µ genomes with the best fitness.
func Elite(µ int, genomes ...evo.Genome) (winners []evo.Genome) {
	winners = make([]evo.Genome, µ)
	pool := make(elcomps, len(genomes))
	for i := range genomes {
		pool[i] = elcomp{genomes[i], 0}
	}
	pool.sort()
	for i := range winners {
		winners[i] = pool[i].Genome
	}
	return winners
}

// ElitePool creates an elite pool selector. Once λ competitors have been put
// into the pool, they are sorted by fitness. The best µ competitors must then
// be retrieved from the pool. Once the winners are retrieved, the pool starts
// accepting competitors for another tournamnent.
func ElitePool(µ, λ int) Pool {
	var p Pool
	p.in = make(chan evo.Genome)
	p.out = make(chan evo.Genome, µ)
	p.close = make(chan chan struct{})

	go func() {
		// the competitors, memory shared accross iterations
		pool := make(elcomps, 0, λ)

		for {
			// wait to receive all competitors
			for len(pool) < λ {
				select {
				case ch := <-p.close:
					ch <- struct{}{}
					return

				case val := <-p.in:
					// we only add the competitor to the pool
					// we do _not_ compute the fitness yet
					pool = append(pool, elcomp{val, 0})
				}
			}

			// we sort the pool by fitness
			// this evaluates the fitness of each member for the first time
			pool.sort()

			// send out the most fit µ genomes
			pool = pool[:µ]
			for i := range pool {
				select {
				case ch := <-p.close:
					ch <- struct{}{}
					return

				case p.out <- pool[i].Genome:
				}
			}
			pool = pool[:0]
		}
	}()

	return p
}

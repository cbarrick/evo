package sel

import (
	"math"
	"math/rand"
	"sort"

	"github.com/cbarrick/evo"
)

// A dummy is used as the opponent in the bye of an odd tournament
type dummy struct{}

func (d dummy) Evolve(_ ...evo.Genome) evo.Genome { return dummy{} }
func (d dummy) Fitness() float64                  { return math.Inf(-1) }
func (d dummy) Close()                            {}

// An rrcomp competes in a round-robin tournament.
type rrcomp struct {
	evo.Genome
	wins int
}

// rrcomps implements the sort interface in descending order by wins
type rrcomps []rrcomp

func (h rrcomps) Len() int           { return len(h) }
func (h rrcomps) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h rrcomps) Less(i, j int) bool { return h[i].wins > h[j].wins }

// tourney performs a concurrent round-robin tournament.
// pool becomes sorted by score.
func (pool rrcomps) tourney(rounds int) {
	var (
		size    = len(pool)            // the size of the tournament
		half    = size / 2             // half that
		tcount  = rounds * half        // number of tournaments
		sched   = rand.Perm(len(pool)) // the tournamnent schedule
		winners = make(chan int)       // communicates the winners
	)

	if size % 2 != 0 {
		panic("odd size round-robin")
	}

	// run returns the best competitor among indices i and j
	// sends the index of the winner over winners
	run := func(i, j int) {
		if pool[i].Fitness() < pool[j].Fitness() {
			winners <- j
		} else {
			winners <- i
		}
	}

	// for each round, start the tournament according to the schedule
	// then rotate the schedule, keeping one element in place, and repeat
	for round := 0; round < rounds; round++ {
		for i := 0; i < half; i++ {
			go run(sched[i], sched[size-1-i])
		}
		carry := sched[0]
		for i := range sched[:size-1] {
			sched[i] = sched[i+1]
		}
		sched[size-2] = carry
	}

	// wait for all competitions to end and keep score
	for i := 0; i < tcount; i++ {
		j := <-winners
		pool[j].wins++
	}

	// finally, sort by score
	sort.Sort(pool)
}

// RoundRobin returns the µ best genomes after some rounds of a tournament.
func RoundRobin(µ, rounds int, genomes ...evo.Genome) (winners []evo.Genome) {
	pool := make(rrcomps, 0, len(genomes)+1)
	for i := range genomes {
		pool = append(pool, rrcomp{genomes[i], 0})
	}
	if len(pool) % 2 != 0 {
		pool = append(pool, rrcomp{dummy{}, -1})
	}
	pool.tourney(rounds)
	winners = make([]evo.Genome, µ)
	for i := range winners {
		winners[i] = pool[i].Genome
	}
	return winners
}

// RoundRobinPool creates a round-robin pool selector. Once λ competitors
// have been put into the pool, the tournamnet is performed. The best µ
// competitors must then be retrieved from the pool. Once the winners are
// retrieved, the pool starts accepting competitors for another tournamnent.
func RoundRobinPool(µ, λ, rounds int) Pool {
	var p Pool
	p.in = make(chan evo.Genome)
	p.out = make(chan evo.Genome)
	p.close = make(chan chan struct{})

	go func() {
		// the competitors, memory shared accross iterations
		pool := make(rrcomps, 0, λ+(λ%2))

		for {
			// wait to receive all competitors
			for len(pool) < λ {
				select {
				case ch := <-p.close:
					ch <- struct{}{}
					return

				case val := <-p.in:
					pool = append(pool, rrcomp{val, 0})
				}
			}

			// do the tournament
			if λ % 2 != 0 {
				pool = append(pool, rrcomp{dummy{}, -1})
			}
			pool.tourney(rounds)

			// send out the µ genomes that won the most
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

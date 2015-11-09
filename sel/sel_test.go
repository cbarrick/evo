package sel_test

import (
	"testing"

	"github.com/cbarrick/evo"
	"github.com/cbarrick/evo/sel"
)

type dummy float64

func (d dummy) Evolve(_ ...evo.Genome) evo.Genome { return d }
func (d dummy) Fitness() float64                  { return float64(d) }
func (d dummy) Close()                            {}

func dummies() []evo.Genome {
	return []evo.Genome{
		dummy(1),
		dummy(2),
		dummy(3),
		dummy(4),
		dummy(5),
		dummy(6),
		dummy(7),
		dummy(8),
		dummy(9),
		dummy(0),
	}
}

func search(ds []evo.Genome, d float64) bool {
	for i := range ds {
		if ds[i].(dummy) == dummy(d) {
			return true
		}
	}
	return false
}

// elite.go
// -------------------------

func TestElite(t *testing.T) {
	pop := dummies()
	elite := sel.Elite(5, pop...)
	ok := search(elite, 9) &&
		search(elite, 8) &&
		search(elite, 7) &&
		search(elite, 6) &&
		search(elite, 5)
	if !ok {
		t.Fail()
	}
}

func TestElitePool(t *testing.T) {
	pop := dummies()
	pool := sel.ElitePool(5, 10)
	for i := range pop {
		pool.Put(pop[i])
	}
	for i := dummy(9); 4 < i; i-- {
		if pool.Get().(dummy) != i {
			t.Fail()
			return
		}
	}
}

// round_robin.go
// -------------------------

func TestRoundRobin(t *testing.T) {
	pop := dummies()
	elite := sel.RoundRobin(5, 10, pop...)
	ok := search(elite, 9) &&
		search(elite, 8) &&
		search(elite, 7) &&
		search(elite, 6) &&
		search(elite, 5)
	if !ok {
		t.Fail()
	}
}

func TestRoundRobinPool(t *testing.T) {
	pop := dummies()
	pool := sel.RoundRobinPool(5, 10, 10)
	for i := range pop {
		pool.Put(pop[i])
	}
	for i := dummy(9); 4 < i; i-- {
		if pool.Get().(dummy) != i {
			t.Fail()
			return
		}
	}
}

// tournament.go
// -------------------------

func TestTournament(t *testing.T) {
	pop := dummies()
	winner := sel.Tournament(pop...)
	if winner != dummy(9) {
		t.Fail()
	}
}

func TestBinaryTournament(t *testing.T) {
	var stats evo.Stats
	pop := dummies()
	for i := 0; i < 1e6; i++ {
		winner := sel.BinaryTournament(pop...).(dummy)
		stats = stats.Put(float64(winner))
	}
	if stats.Mean() < 5.5 || 6.5 < stats.Mean() {
		t.Fail()
	}
}

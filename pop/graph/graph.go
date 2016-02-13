// Package graph provides a spatial population for diffusion and island models.
//
// Graph populations map genomes to nodes in a graph. Each node is evolved in
// parallel, and only sees neighboring nodes as suitors. When used as a
// meta-population, this technique is known as the island model. When used as a
// regular population, it is known as the diffusion model.
package graph

import (
	"time"

	"github.com/cbarrick/evo"
)

type Graph []node

type node struct {
	val    *evo.Genome
	peers  []*node
	getc   chan chan evo.Genome
	setc   chan chan evo.Genome
	closec chan chan struct{}
	done   chan struct{}
}

// Grid creates a new graph population arranged as a 2D grid.
func Grid(size int) Graph {
	width := size << 1
	layout := make([][]int, size)
	for i := range layout {
		layout[i] = make([]int, 4)
		layout[i][0] = (i + 1 + size) % size
		layout[i][1] = (i - 1 + size) % size
		layout[i][2] = (i + width + size) % size
		layout[i][3] = (i - width + size) % size
	}
	return Custom(layout)
}

// Hypercube creates a new graph population arranged as a hypercube.
func Hypercube(size int) Graph {
	var dim uint
	for size > (1 << dim) {
		dim++
	}
	layout := make([][]int, size)
	for i := 0; i < size; i++ {
		layout[i] = make([]int, dim)
		for j := range layout[i] {
			layout[i][j] = (i ^ (1 << uint(j))) % size
		}
	}
	return Custom(layout)
}

// Ring creates a new graph population arranged as a ring.
func Ring(size int) Graph {
	layout := make([][]int, size)
	for i := 0; i < size; i++ {
		layout[i] = make([]int, 2)
		layout[i][0] = (i - 1 + size) % size
		layout[i][0] = (i + 1) % size
	}
	return Custom(layout)
}

// Custom creates a new graph population with a custom layout.
// The layout is specified as an adjacency list.
func Custom(layout [][]int) Graph {
	g := make([]node, len(layout))
	for i := range g {
		peers := make([]*node, len(layout[i]))
		for j := range layout[i] {
			peers[j] = &g[j]
		}
		g[i].peers = peers
	}

	return g
}

// Stats returns statistics on the fitness of genomes in the population.
func (g Graph) Stats() (s evo.Stats) {
	for i := range g {
		s = s.Put(g[i].get().Fitness())
	}
	return s
}

// Fitness returns the maximum fitness within the population.
func (g Graph) Fitness() float64 {
	return g.Stats().Max()
}

// Evolve starts the optimization in a separate goroutine.
func (g Graph) Evolve(members []evo.Genome, body evo.EvolveFn) {
	for i := range g {
		g[i].val = &members[i]
		g[i].getc = make(chan chan evo.Genome)
		g[i].setc = make(chan chan evo.Genome)
		g[i].closec = make(chan chan struct{}, 1)
	}
	for i := range g {
		i := i
		go g[i].run(body)
	}
}

// Stop terminates the optimization.
func (g Graph) Stop() {
	ch := make(chan struct{})
	for i := range g {
		g[i].closec <- ch
		<-ch
		close(g[i].getc)
		close(g[i].setc)
	}
}

// Poll executes a function at some frequency for the duration of the
// current optimization. If the function returns true, the current optimization
// is halted.
func (g Graph) Poll(freq time.Duration, cond evo.ConditionFn) {
	done := g[0].closec
	go func() {
		for {
			select {
			case <-time.After(freq):
				if cond() {
					g.Stop()
					return
				}
			case ch := <-done:
				done <- ch
				return
			}
		}
	}()
}

// Wait blocks until the evolution terminates.
func (g Graph) Wait() {
	for i := range g {
		g[i].wait()
	}
}

func (n *node) wait() {
	n.closec <- <-n.closec
}

// get returns the genome underlying the node.
func (n *node) get() evo.Genome {
	getter := <-n.getc
	if getter == nil {
		return *n.val
	}
	return <-getter
}

// The main goroutine.
func (n *node) run(body evo.EvolveFn) {
	var (
		// drives the main loop
		loop = make(chan struct{}, 1)

		// used to access/mutate the value
		getter = make(chan evo.Genome)
		setter = make(chan evo.Genome)
	)

	loop <- struct{}{}

	for {
		select {
		case <-loop:
			go func() {
				suiters := make([]evo.Genome, len(n.peers))
				for i := range n.peers {
					suiters[i] = n.peers[i].get()
				}
				setter <- body(*n.val, suiters)
				loop <- struct{}{}
			}()

		case n.getc <- getter:
			getter <- *n.val

		case *n.val = <-setter:

		case ch := <-n.closec:
			if subpop, ok := (*n.val).(evo.Population); ok {
				subpop.Stop()
			}
			ch <- struct{}{}
			n.closec <- ch

			return
		}
	}
}

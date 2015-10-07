// Package graph needs documentation
package graph

import (
	"math/rand"
	"runtime"
	"time"

	"github.com/cbarrick/evo"
)

// Nodes
// -------------------------

// Nodes wrap a genome and are aggregated to form a graph of genomes.
// A node manages the lifecycle of one slot in a population concurrently with
// all other nodes in the graph. The underlying genome is only allowed to mate
// with genomes from adjacent nodes.
type node struct {
	val    evo.Genome
	peers  []*node
	valc   chan evo.Genome
	delayc chan time.Duration
	closec chan chan struct{}
}

func (n *node) init(val evo.Genome, peers []*node) {
	n.val = val
	n.peers = peers
	n.valc = make(chan evo.Genome)
	n.delayc = make(chan time.Duration)
	n.closec = make(chan chan struct{})
}

func (n *node) run() {
	var (
		delay   time.Duration
		mate    = time.After(delay)
		suiters = make([]evo.Genome, len(n.peers))
		done    = make(chan evo.Genome)
		nextval = n.val
	)

	runtime.SetFinalizer(n.val, nil)
	runtime.SetFinalizer(n.val, func(val evo.Genome) {
		val.Close()
	})

	for {
		select {

		case delay = <-n.delayc:

		case n.valc <- n.val:
		case nextval = <-n.valc:

		case <-mate:
			go func(oldval evo.Genome) {
				var ok bool
				for i := range n.peers {
					suiters[i], ok = <-n.peers[i].valc
					if !ok {
						return
					}
				}
				newval := oldval.Evolve(suiters...)
				done <- newval
			}(n.val)

		case val := <-done:
			if nextval == n.val {
				nextval = val
			} else if val != n.val && val != nextval {
				val.Close()
			} else {
				n.val = nextval
				runtime.SetFinalizer(n.val, nil)
				runtime.SetFinalizer(n.val, func(val evo.Genome) {
					val.Close()
				})
			}
			mate = time.After(delay)

		case ch := <-n.closec:
			ch <- struct{}{}
			return
		}
	}
}

// Close stops the main goroutine
func (n *node) Close() {
	ch := make(chan struct{})
	n.closec <- ch
	<-ch
	close(n.valc)
	close(n.delayc)
	close(n.closec)
}

// Value returns the value underlying the node.
func (n *node) value() (val evo.Genome) {
	val, ok := <-n.valc
	if !ok {
		val = n.val
	}
	return val
}

// SetDelay sets a delay between each iteration
func (n *node) setDelay(d time.Duration) {
	n.delayc <- d
}

// Graphs
// -------------------------

// Graphs aggregate nodes into a population.
type Graph struct {
	nodes []node
}

// Iter returns an iterator ranging over the values of the population.
func (g *Graph) Iter() evo.Iterator {
	return iterate(g)
}

// Stats returns statistics on the fitness of genomes in the population.
func (g *Graph) Stats() (s evo.Stats) {
	for i := g.Iter(); i.Value() != nil; i.Next() {
		s = s.Insert(i.Value().Fitness())
	}
	return s
}

// Close stops the goroutines of all nodes.
func (g *Graph) Close() {
	for i := range g.nodes {
		g.nodes[i].Close()
	}
}

// Fitness returns the maximum fitness within the graph.
func (g *Graph) Fitness() float64 {
	return g.Stats().Max()
}

// Evolve performs a random migration between this graph and a random suiter.
func (g *Graph) Evolve(suiters ...evo.Genome) evo.Genome {
	h := suiters[rand.Intn(len(suiters))].(*Graph)
	i := rand.Intn(len(g.nodes))
	j := rand.Intn(len(h.nodes))
	x := g.nodes[i].value()
	y := h.nodes[j].value()
	g.nodes[i].valc <- y
	h.nodes[j].valc <- x
	return g
}

// SetDelay sets a delay between each iteration of each node
func (g *Graph) SetDelay(d time.Duration) *Graph {
	for i := range g.nodes {
		g.nodes[i].setDelay(d)
	}
	return g
}

// Iterator
// -------------------------

type iter struct {
	sub evo.Iterator
	idx int
	g   *Graph
	val evo.Genome
}

func iterate(g *Graph) evo.Iterator {
	var it iter
	it.idx = 0
	it.g = g
	it.val = g.nodes[it.idx].value()
	if pop, ok := it.val.(evo.Population); ok {
		it.sub = pop.Iter()
	}
	return &it
}

func (it *iter) Value() evo.Genome {
	if it.sub != nil {
		return it.sub.Value()
	}
	return it.val
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
		if it.idx >= len(it.g.nodes) {
			it.g = nil
			it.val = nil
		} else {
			it.val = it.g.nodes[it.idx].value()
			if pop, ok := it.val.(evo.Population); ok {
				it.sub = pop.Iter()
			}
		}
	}
}

// Functions
// -------------------------

// New creates a new diffusion population with a layout chosen by the system.
// Currently, the hypercube layout is always used.
func New(values []evo.Genome) *Graph {
	return Hypercube(values)
}

// Grid creates a new diffusion population arranged in a 2D grid.
func Grid(values []evo.Genome) *Graph {
	offset := len(values) / 2
	layout := make([][]int, len(values))
	for i := range values {
		layout[i] = make([]int, 4)
		layout[i][0] = ((i + 1) + len(values)) % len(values)
		layout[i][1] = ((i - 1) + len(values)) % len(values)
		layout[i][2] = ((i + offset) + len(values)) % len(values)
		layout[i][3] = ((i - offset) + len(values)) % len(values)
	}
	return Custom(layout, values)
}

// Hypercube creates a new diffusion population arranged in a hypercube graph.
func Hypercube(values []evo.Genome) *Graph {
	var dimension uint
	for dimension = 0; len(values) > (1 << dimension); dimension++ {
	}
	layout := make([][]int, len(values))
	for i := range values {
		layout[i] = make([]int, dimension)
		for j := range layout[i] {
			layout[i][j] = (i ^ (1 << uint(j))) % len(values)
		}
	}
	return Custom(layout, values)
}

// Ring creates a new diffusion population arranged in a ring.
func Ring(values []evo.Genome) *Graph {
	layout := make([][]int, len(values))
	for i := range values {
		layout[i] = make([]int, 2)
		layout[i][0] = (i - 1 + len(values)) % len(values)
		layout[i][0] = (i + 1) % len(values)
	}
	return Custom(layout, values)
}

// Custom creates a new diffusion population with a custom layout.
// The layout is specified as an adjacency list in terms of position, e.g. if
// `layout[0] == [1,2,3]` then the 0th node will have three peers, namely the
// 1st, 2nd, and 3rd nodes.
func Custom(layout [][]int, values []evo.Genome) *Graph {
	g := new(Graph)
	g.nodes = make([]node, len(values))
	for i := range g.nodes {
		val := values[i]
		peers := make([]*node, len(layout[i]))
		for j := range layout[i] {
			peers[j] = &g.nodes[j]
		}
		g.nodes[i].init(val, peers)
	}
	for i := range g.nodes {
		go g.nodes[i].run()
	}

	return g
}

// Package graph provides a spatial population for diffusion and island models.
//
// Graph populations map genomes to nodes in a graph. Each node is evolved in
// parallel, and only sees neighboring nodes as suitors. When used as a
// meta-population, this technique is known as the island model. When used as a
// regular population, it is known as the diffusion model.
package graph

import (
	"math/rand"
	"runtime"
	"time"

	"github.com/cbarrick/evo"
)

// Nodes
// -------------------------

// Node wraps a genome and manages the lifecycle of one slot in a population.
// Nodes evolve their underlying genome concurrently with all other nodes in a
// graph. The underlying genome takes suitors from adjacent nodes.
type node struct {
	val     evo.Genome
	peers   []*node
	delay   time.Duration
	valc    chan evo.Genome
	delayc  chan time.Duration
	closec  chan chan struct{}
	running bool
}

func (n *node) run() {
	var (
		mate    = time.After(n.delay)
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

		case n.delay = <-n.delayc:

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
			mate = time.After(n.delay)

		case ch := <-n.closec:
			ch <- struct{}{}
			return
		}
	}
}

// Start starts the node evolving it's genome in a separate goroutine.
func (n *node) Start() {
	if !n.running {
		n.valc = make(chan evo.Genome)
		n.delayc = make(chan time.Duration)
		n.closec = make(chan chan struct{})
		n.running = true
		go n.run()
	}
}

// Close stops the node from evolving it's genome.
func (n *node) Close() {
	if n.running {
		ch := make(chan struct{})
		n.closec <- ch
		<-ch
		close(n.valc)
		close(n.delayc)
		close(n.closec)
		n.running = false
	}
}

// Value returns the genome underlying the node.
func (n *node) Value() (val evo.Genome) {
	if n.running {
		val = <-n.valc
	} else {
		val = n.val
	}
	return val
}

// SetDelay sets a delay between each evolution.
func (n *node) SetDelay(d time.Duration) {
	if n.running {
		n.delayc <- d
	} else {
		n.delay = d
	}
}

// Graphs
// -------------------------

type graph struct {
	nodes []node
}

func (g *graph) Start() {
	for i := range g.nodes {
		g.nodes[i].Start()
	}
}

func (g *graph) Close() {
	for i := range g.nodes {
		g.nodes[i].Close()
	}
}

func (g *graph) Iter() evo.Iterator {
	return iterate(g)
}

func (g *graph) Stats() (s evo.Stats) {
	for i := g.Iter(); i.Value() != nil; i.Next() {
		s = s.Put(i.Value().Fitness())
	}
	return s
}

func (g *graph) Fitness() float64 {
	return g.Stats().Max()
}

func (g *graph) Evolve(suiters ...evo.Genome) evo.Genome {
	h := suiters[rand.Intn(len(suiters))].(*graph)
	i := rand.Intn(len(g.nodes))
	j := rand.Intn(len(h.nodes))
	x := g.nodes[i].Value()
	y := h.nodes[j].Value()
	g.nodes[i].valc <- y
	h.nodes[j].valc <- x
	return g
}

func (g *graph) SetDelay(d time.Duration) {
	for i := range g.nodes {
		g.nodes[i].SetDelay(d)
	}
}

// Iterator
// -------------------------

type iter struct {
	sub evo.Iterator
	idx int
	g   *graph
	val evo.Genome
}

func iterate(g *graph) evo.Iterator {
	var it iter
	it.idx = 0
	it.g = g
	it.val = g.nodes[it.idx].Value()
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
			it.val = it.g.nodes[it.idx].Value()
			if pop, ok := it.val.(evo.Population); ok {
				it.sub = pop.Iter()
			}
		}
	}
}

// Functions
// -------------------------

// New creates a new graph population. No particular layout is guarenteed.
func New(values []evo.Genome) *graph {
	return Hypercube(values)
}

// Grid creates a new graph population arranged as a 2D grid.
func Grid(values []evo.Genome) *graph {
	width := len(values) << 1
	layout := make([][]int, len(values))
	for i := range layout {
		layout[i] = make([]int, 4)
		layout[i][0] = (i + 1 + len(values)) % len(values)
		layout[i][1] = (i - 1 + len(values)) % len(values)
		layout[i][2] = (i + width + len(values)) % len(values)
		layout[i][3] = (i - width + len(values)) % len(values)
	}
	return Custom(layout, values)
}

// Hypercube creates a new graph population arranged as a hypercube.
func Hypercube(values []evo.Genome) *graph {
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

// Ring creates a new graph population arranged as a ring.
func Ring(values []evo.Genome) *graph {
	layout := make([][]int, len(values))
	for i := range values {
		layout[i] = make([]int, 2)
		layout[i][0] = (i - 1 + len(values)) % len(values)
		layout[i][0] = (i + 1) % len(values)
	}
	return Custom(layout, values)
}

// Custom creates a new graph population with a custom layout.
// The layout is specified as an adjacency list in terms of position, e.g. if
// layout[0] == [1,2,3] then the 0th node will have three peers, namely the
// 1st, 2nd, and 3rd nodes.
func Custom(layout [][]int, values []evo.Genome) *graph {
	g := new(graph)
	g.nodes = make([]node, len(values))
	for i := range g.nodes {
		val := values[i]
		peers := make([]*node, len(layout[i]))
		for j := range layout[i] {
			peers[j] = &g.nodes[j]
		}
		g.nodes[i].val = val
		g.nodes[i].peers = peers
	}

	return g
}

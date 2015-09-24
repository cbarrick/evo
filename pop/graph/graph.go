// Package graph needs documentation
package graph

import (
	"math/rand"
	"strconv"
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
	delayc chan time.Duration
	setval chan evo.Genome
	getval chan evo.Genome
	closec chan int
}

func (n *node) init() {
	n.delayc = make(chan time.Duration)
	n.setval = make(chan evo.Genome, 1)
	n.getval = make(chan evo.Genome)
	n.closec = make(chan int)
}

func (n *node) run() {
	var (
		delay   time.Duration
		mate    = time.After(delay)
		suiters = make([]evo.Genome, len(n.peers))
	)

	for {
		select {

		// gat and set the delay
		case n.delayc <- delay:
		case delay = <-n.delayc:

		// get underlying genome
		case n.getval <- n.val:

		// set the underlying genome
		case newval := <-n.setval:
			if n.val != newval {
				n.val.Close()
				n.val = newval
			}
			n.setval = make(chan evo.Genome, 1)

		// close channels and underlying genome
		case x := <-n.closec:
			close(n.getval)
			n.val.Close()
			n.closec <- x
			return

		// do one iteration
		case <-mate:
			mate = nil
			go func(setval chan evo.Genome) {
				for i := range n.peers {
					suiters[i] = n.peers[i].value()
				}
				newval := n.val.Cross(suiters...)
				setval <- newval
				mate = time.After(<-n.delayc)
			}(n.setval)
		}
	}
}

// Close stops the main goroutine
func (n *node) Close() {
	n.closec <- 1
	<-n.closec
	close(n.closec)
}

// Value returns the current underlying value.
func (n *node) value() (val evo.Genome) {
	val = <-n.getval
	if val == nil {
		return n.val
	}
	return val
}

// SetValue sets the value of the node
func (n *node) setValue(val evo.Genome) {
	c := n.setval
	c <- val
	<-c
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

// View constructs a view of genomes in the graph.
func (g *Graph) View() evo.View {
	members := make([]evo.Genome, len(g.nodes))
	for i := range members {
		members[i] = g.nodes[i].value()
	}
	return evo.NewView(members...)
}

// Close stops the goroutines of all nodes.
func (g *Graph) Close() {
	for i := range g.nodes {
		g.nodes[i].Close()
	}
}

// Fitness returns the maximum fitness within the graph.
func (g *Graph) Fitness() (f float64) {
	v := g.View()
	f = v.Max().Fitness()
	v.Recycle()
	return f
}

// Cross injects the best genome of the suiter into a random node in the graph.
func (g *Graph) Cross(suiters ...evo.Genome) evo.Genome {
	// mate by replacing a random node from g
	// with the best node from a random suiter
	i := rand.Intn(len(suiters))
	h := suiters[i].(*Graph)
	g.nodes[rand.Intn(len(g.nodes))].setValue(h.Max())
	return g
}

// Max returns the best genome in the population.
func (g *Graph) Max() (max evo.Genome) {
	v := g.View()
	max = v.Max()
	v.Recycle()
	return max
}

// SetDelay sets a delay between each iteration of each node
func (g *Graph) SetDelay(d time.Duration) *Graph {
	for i := range g.nodes {
		g.nodes[i].setDelay(d)
	}
	return g
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

	// validate the layout
	size := len(values)
	if len(layout) != size {
		panic("invalid layout, len(layout) != len(values)")
	}
	for i := range layout {
		for j := range layout[i] {
			if layout[i][j] >= size {
				panic("invalid layout, no such node: " + strconv.Itoa(layout[i][j]))
			}
		}
	}

	// make the graph
	g := new(Graph)
	g.nodes = make([]node, len(values))

	// for each node, assign its initial value and peers
	// and initialize its other members
	for i := range g.nodes {
		n := &g.nodes[i]
		n.val = values[i]
		n.peers = make([]*node, len(layout[i]))
		for j := range layout[i] {
			n.peers[j] = &g.nodes[j]
		}
		n.init()
	}

	// start each node's main loop
	for i := range g.nodes {
		go g.nodes[i].run()
	}

	return g
}

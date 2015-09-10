// Package diffusion implements a fine-grained parallel genetic algorithm.
//
// A diffusion population maps each genome to a node in a connected graph. Each
// node manages one slot of the population. For each iteration, a node calls the
// `Cross` method of its underlying genome, passing the adjacent genomes as the
// mating pool. The underlying genome is replaced by the result of that call.
// In this way, good genes diffuse through the population over many iterations.
// Each node manages its lifecycle in parallel.
package diffusion

import (
	"math/rand"
	"strconv"

	"github.com/cbarrick/evo"
)

// Nodes
// -------------------------

// Nodes wrap a genome and are aggregated to form a graph of genomes.
// A node manages the lifecycle of one slot in a population concurrently with
// all other nodes in the graph. The underlying genome is only allowed to mate
// with genomes from adjacent nodes.
type node struct {
	value  evo.Genome
	peers  []*node
	valuec chan evo.Genome
	closec chan int
}

// Init must be called on each node before any node is started.
func (n *node) init() {
	n.valuec = make(chan evo.Genome)
	n.closec = make(chan int)
}

// Start spins up the main goroutine.
func (n *node) Start() {
	go n.loop()
}

func (n *node) loop() {
	var (
		// Used as temporary storage by the mating routine
		suiters = make([]evo.Genome, len(n.peers))

		// Channel on which the mating routine communicates
		updates = make(chan evo.Genome)

		// This flag is set when the value within the node has been recently
		// overridden. It causes the current mating routine (which will be using
		// the old value) to be ignored.
		manualOverride bool
	)

	// set sets the underlying value of the node.
	// If it is different from the existing value, the old value is closed.
	set := func(newval evo.Genome) {
		if n.value != newval {
			n.value.Close()
			n.value = newval
		}
	}

	// mate calls the Cross method of the underlying genome of the node,
	// presenting the adjacent genomes as suiters. The result of the cross is
	// returned over the updates channel. Exactly one instance of this function
	// is running as a goroutine as long as the node is alive.
	mate := func(value evo.Genome) {
		for i := range suiters {
			suiters[i] = n.peers[i].Value()
		}
		updates <- value.Cross(suiters...)
	}
	go mate(n.value)

	for {
		select {

		// get and set the underlying genome
		case n.valuec <- n.value:
		case val := <-n.valuec:
			set(val)
			manualOverride = true

		// update the underlying genome whenever the mating routine returns
		case child := <-updates:
			if !manualOverride {
				set(child)
			}
			manualOverride = false
			go mate(n.value)

			// cleanup by closing channels and waiting on the last mating routine
		case x := <-n.closec:
			close(n.valuec)
			if updates != nil {
				n.value = <-updates
			}
			n.value.Close()
			n.closec <- x
			return
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
func (n *node) Value() (value evo.Genome) {
	value = <-n.valuec
	if value == nil {
		return n.value
	}
	return value
}

// Graphs
// -------------------------

// Graphs aggregate nodes into a population.
type graph struct {
	nodes []node
}

// View constructs a view of genomes in the graph..
func (g *graph) View() evo.View {
	members := make([]evo.Genome, len(g.nodes))
	for i := range members {
		members[i] = g.nodes[i].Value()
	}
	return evo.NewView(members...)
}

// Close stops the goroutines of all nodes.
func (g *graph) Close() {
	for i := range g.nodes {
		g.nodes[i].Close()
	}
}

// Fitness returns the maximum fitness within the graph.
func (g *graph) Fitness() (f float64) {
	v := g.View()
	f = v.Max().Fitness()
	v.Close()
	return f
}

// Cross injects the best genome of the suiter into a random node in the graph.
func (g *graph) Cross(suiters ...evo.Genome) evo.Genome {
	// pick a mate, try not to mate with self
	i := rand.Intn(len(suiters))
	h := suiters[i].(*graph)
	if h == g {
		if len(suiters) > 1 {
			// remove conflict and try again
			newsuiters := make([]evo.Genome, len(suiters)-1)
			copy(newsuiters[0:], suiters[:i])
			copy(newsuiters[i:], suiters[i+1:])
			return g.Cross(newsuiters...)
		}
		return g // our only option is to mate with self
	}

	// mate by setting a random node in g with the best node in h
	g.nodes[rand.Intn(len(g.nodes))].valuec <- h.Max()
	return g
}

// Max returns the best genome in the population.
func (g *graph) Max() (max evo.Genome) {
	v := g.View()
	max = v.Max()
	v.Close()
	return max
}

// Functions
// -------------------------

// New creates a new diffusion population with a layout chosen by the system.
// Currently, the hypercube layout is always used.
func New(values []evo.Genome) evo.Population {
	return Hypercube(values)
}

// Grid creates a new diffusion population arranged in a 2D grid.
func Grid(values []evo.Genome) evo.Population {
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
func Hypercube(values []evo.Genome) evo.Population {
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
func Ring(values []evo.Genome) evo.Population {
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
func Custom(layout [][]int, values []evo.Genome) evo.Population {

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
	g := new(graph)
	g.nodes = make([]node, len(values))

	// for each node, assign its initial value and peers
	// and initialize its other members
	for i := range g.nodes {
		n := &g.nodes[i]
		n.value = values[i]
		n.peers = make([]*node, len(layout[i]))
		for j := range layout[i] {
			n.peers[j] = &g.nodes[j]
		}
		n.init()
	}

	// start each node's main loop
	for i := range g.nodes {
		g.nodes[i].Start()
	}

	return g
}

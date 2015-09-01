// Package diffusion implements a fine-grained parallel genetic algorithm.
//
// A diffusion population maps each genome to a node in a connected graph. Each
// node manages the lifecycle of exactly one genome at a time. For each
// iteration, a node calls the `Cross` method of its underlying genome, passing
// a subset of the adjacent genomes as arguments. The underlying genome is
// replaced by the result of that call. All nodes run concurrently on separate
// goroutines.
package diffusion

import (
	"math"
	"math/rand"
	"strconv"

	"github.com/cbarrick/evo"
)

// Nodes
// -------------------------

type node struct {
	// these need to be initially set
	value evo.Genome
	peers []*node

	// communication channels for the main loop
	valuec chan evo.Genome
	closec chan chan error
}

func (n *node) init() {
	n.valuec = make(chan evo.Genome)
	n.closec = make(chan chan error)
}

func (n *node) loop() {
	var (
		globalErr error
		suiters   = make([]evo.Genome, len(n.peers)>>1)
		updates   = make(chan evo.Genome)
	)

	setValue := func(val evo.Genome) {
		if n.value != val {
			err := n.value.Close()
			if err != nil {
				globalErr = err
			}
			n.value = val
		}
	}

	update := func() {
		perm := rand.Perm(len(suiters))
		for i := range suiters {
			suiters[i] = n.peers[perm[i]].Value()
		}
		updates <- n.value.Cross(suiters...)
	}
	go update()

	for {
		select {
		case n.valuec <- n.value:
			break

		case val := <-n.valuec:
			setValue(val)

		case child := <-updates:
			setValue(child)
			if globalErr == nil {
				go update()
			}

		case ch := <-n.closec:
			close(n.valuec)
			close(n.closec)

			// if there is an update running in another goroutine
			// then we must wait for it so that it closes cleanly
			if updates != nil {
				<-updates
			}

			ch <- globalErr
			return
		}
	}
}

func (n *node) Close() error {
	errc := make(chan error)
	n.closec <- errc
	return <-errc
}

func (n *node) Value() (value evo.Genome) {
	value = <-n.valuec
	if value == nil {
		return n.value
	}
	return value
}

func (n *node) Swap(m *node) {
	nval := <-n.valuec
	mval := <-m.valuec
	switch {
	case nval == nil && mval == nil:
		n.value, m.value = m.value, n.value
	case nval == nil:
		m.valuec <- n.value
		n.value = mval
	case mval == nil:
		n.valuec <- m.value
		m.value = nval
	default:
		n.valuec <- mval
		m.valuec <- nval
	}
}

// Graphs
// -------------------------

type graph struct {
	nodes []node
}

func (g *graph) Members() (values []evo.Genome) {
	values = make([]evo.Genome, len(g.nodes))
	for i := range values {
		values[i] = g.nodes[i].Value()
	}
	return values
}

func (g *graph) Close() (err error) {
	for i := range g.nodes {
		err_i := g.nodes[i].Close()
		if err_i != nil {
			err = err_i
		}
	}
	return err
}

func (g *graph) Fitness() float64 {
	return evo.Max(g.Members()...).Fitness()
}

func (g *graph) Cross(suiters ...evo.Genome) evo.Genome {
	i := rand.Intn(len(suiters))
	n := g.MaxNode()
	m := suiters[i].(*graph).MaxNode()
	n.Swap(m)
	return g
}

func (g *graph) MaxNode() (best *node) {
	fitness := math.Inf(-1)
	for i := range g.nodes {
		newfitness := g.nodes[i].Value().Fitness()
		if newfitness > fitness {
			fitness = newfitness
			best = &g.nodes[i]
		}
	}
	return best
}

// Functions
// -------------------------

// New creates a new diffusion population with the default topology.
func New(values []evo.Genome) evo.Population {
	return Hypercube(values)
}

// Grid creates a new diffusion population arranged in a 2D grid.
func Grid(values []evo.Genome) evo.Population {
	offset := len(values) >> 1
	topology := make([][]int, len(values))
	for i := range values {
		topology[i] = make([]int, 4)
		topology[i][0] = ((i + 1) + len(values)) % len(values)
		topology[i][1] = ((i - 1) + len(values)) % len(values)
		topology[i][2] = ((i + offset) + len(values)) % len(values)
		topology[i][3] = ((i - offset) + len(values)) % len(values)
	}
	return Custom(topology, values)
}

// Hypercube creates a new diffusion population arranged in a hypercube graph.
func Hypercube(values []evo.Genome) evo.Population {
	var dimension uint
	for dimension = 0; len(values) > (1 << dimension); dimension++ {
	}
	topology := make([][]int, len(values))
	for i := range values {
		topology[i] = make([]int, dimension)
		for j := range topology[i] {
			topology[i][j] = (i ^ (1 << uint(j))) % len(values)
		}
	}
	return Custom(topology, values)
}

// Ring creates a new diffusion population arranged in a ring.
func Ring(values []evo.Genome) evo.Population {
	topology := make([][]int, len(values))
	for i := range values {
		topology[i] = make([]int, 2)
		topology[i][0] = (i - 1 + len(values)) % len(values)
		topology[i][0] = (i + 1) % len(values)
	}
	return Custom(topology, values)
}

// Custom creates a new diffusion population with a custom topology.
// The topology maps each node to the list of its peers, e.g. if
// `topology[0] == [1,2,3]` then the 0th node will have three peers,
// namely the 1st, 2nd, and 3rd nodes.
func Custom(topology [][]int, values []evo.Genome) evo.Population {

	// validate the topology
	size := len(values)
	if len(topology) != size {
		panic("invalid topology, len(topology) != len(values)")
	}
	for i := range topology {
		for j := range topology[i] {
			if topology[i][j] >= size {
				panic("invalid topology, no such node: " + strconv.Itoa(topology[i][j]))
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
		n.peers = make([]*node, len(topology[i]))
		for j := range topology[i] {
			n.peers[j] = &g.nodes[j]
		}
		n.init()
	}

	// start each node's main loop
	for i := range g.nodes {
		go g.nodes[i].loop()
	}

	return g
}

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
		// Used as temporary storage by the mating routine
		suiters = make([]evo.Genome, len(n.peers)>>1)

		// Channel on which the mating routine communicates
		updates = make(chan evo.Genome)

		// This flag is set when the value within the node has been recently
		// overridden. It causes the current mating routine (which will be using
		// the old value) to be ignored.
		manualOverride bool

		// Set whenever an error is encountered
		err error
	)

	// set sets the underlying value of the node. If it is different from the
	// existing value, the old value is closed. If an error occurs, err is set
	// to the error.
	set := func(newval evo.Genome) {
		var err_ error
		if n.value != newval {
			err_ = n.value.Close()
			n.value = newval
		}
		if err_ != nil {
			err = err_
		}
	}

	// mate calls the Cross method of the underlying value of the node,
	// presenting a subset of the adjacent values as suiters. The size of the
	// subset is determined by the length of the suiters slice. The result of
	// the cross is returned over the updates channel. This function should be
	// called as a goroutine.
	mate := func(value evo.Genome) {
		perm := rand.Perm(len(suiters))
		for i := range suiters {
			suiters[i] = n.peers[perm[i]].Value()
		}
		updates <- value.Cross(suiters...)
	}
	go mate(n.value)

	for {
		select {
		case n.valuec <- n.value:
			break

		case val := <-n.valuec:
			set(val)
			manualOverride = true

		case child := <-updates:
			if !manualOverride {
				set(child)
			}
			manualOverride = false
			if err == nil {
				go mate(n.value)
			} else {
				updates = nil
			}

		case ch := <-n.closec:
			close(n.valuec)
			close(n.closec)
			if updates != nil {
				n.value = <-updates
			}
			err_ := n.value.Close()
			if err_ != nil {
				err = err_
			}
			ch <- err
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

// Graphs
// -------------------------

type graph struct {
	nodes []node
}

func (g *graph) View() (values evo.View) {
	values = make(evo.View, len(g.nodes))
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
	return g.Max().Fitness()
}

func (g *graph) Cross(suiters ...evo.Genome) evo.Genome {
	h := suiters[rand.Intn(len(suiters))].(*graph)
	gmax := g.Max()
	hmax := h.Max()
	g.nodes[rand.Intn(len(g.nodes))].valuec <- hmax
	h.nodes[rand.Intn(len(h.nodes))].valuec <- gmax
	return g
}

func (g *graph) Max() evo.Genome {
	return g.View().Max()
}

func (g *graph) Min() evo.Genome {
	return g.View().Min()
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

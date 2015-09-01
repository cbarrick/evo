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

// type vfpair combines a value (genome) and a fitness
// so that the fitness need not be recomputed from the value
type vfpair struct {
	value   evo.Genome
	fitness float64
}

type node struct {
	// these need to be initially set
	value evo.Genome
	peers []*node

	// the fitness of the current value is cached
	// to avoid possibly expensive recomputation
	fitness float64

	// communication channels for the main loop
	valuec   chan vfpair
	closec   chan chan error
	pausec   chan bool
}

func (n *node) init() {
	n.valuec = make(chan vfpair)
	n.closec = make(chan chan error)
	n.pausec = make(chan bool)
}

func (n *node) loop() {
	n.fitness = n.value.Fitness()

	suiters := make([]evo.Genome, len(n.peers)>>1)
	updates := make(chan evo.Genome)
	update := func() {
		perm := rand.Perm(len(suiters))
		for i := range suiters {
			suiters[i], _ = n.peers[perm[i]].Value()
		}
		updates <- n.value.Cross(suiters...)
	}
	go update()

	var err error
	for {
		select {
		case n.valuec <- vfpair{n.value, n.fitness}:
			break

		case ch := <-n.closec:
			close(n.valuec)
			close(n.closec)
			close(n.pausec)

			// if there is an update running in another goroutine
			// then we must wait for it so that it closes cleanly
			if updates != nil {
				<-updates
			}

			ch <- err
			return

		case x := <-n.pausec:
			n.pausec <- x
			// the value may change while we are paused
			// so we must update the fitness
			n.fitness = n.value.Fitness()

		case child := <-updates:
			if n.value != child {
				err = n.value.Close()
				if err != nil {
					updates = nil
					break
				}
				n.value = child
				n.fitness = child.Fitness()
			}
			go update()
		}
	}
}

func (n *node) Close() error {
	errc := make(chan error)
	n.closec <- errc
	return <-errc
}

func (n *node) Value() (value evo.Genome, fitness float64) {
	vf := <-n.valuec
	if vf.value == nil {
		return n.value, n.fitness
	}
	return vf.value, vf.fitness
}

func (n *node) pause() {
	n.pausec <- true
}

func (n *node) resume() {
	<-n.pausec
}

// Graphs
// -------------------------

type graph struct {
	nodes []node
}

func (g *graph) Stats() (s evo.Stats) {
	var (
		value     evo.Genome
		fitness   float64
		diversity float64
		maxfit    = math.Inf(-1)
		minfit    = math.Inf(+1)
		count     int
	)

	s.Members = make([]evo.Genome, 0, len(g.nodes))
	for i := range g.nodes {
		value, fitness = g.nodes[i].Value()
		if fitness > maxfit {
			s.Max = value
			maxfit = fitness
		}
		if fitness < minfit {
			s.Min = value
			minfit = fitness
		}
		for j := range s.Members {
			diversity *= float64(count)
			diversity += value.Difference(s.Members[j])
			diversity /= float64(count + 1)
			count++
		}
		s.Members = append(s.Members, value)
	}

	s.N = make(map[string]float64, 4)
	s.N["maxfit"] = maxfit
	s.N["minfit"] = minfit
	s.N["diversity"] = diversity
	s.N["convergence"] = maxfit - minfit

	return s
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
	return g.Stats().N["maxfit"]
}

func (g *graph) Cross(suiters ...evo.Genome) evo.Genome {
	i := rand.Intn(len(suiters))
	us := g.MaxNode()
	them := suiters[i].(*graph).MaxNode()
	us.pause()
	them.pause()
	us.value, them.value = them.value, us.value
	us.resume()
	them.resume()
	return g
}

func (g *graph) Difference(other evo.Genome) float64 {
	us := g.Stats().Max
	them := other.(evo.Population).Stats().Max
	return us.Difference(them)
}

func (g *graph) MaxNode() (best *node) {
	fitness := math.Inf(-1)
	for i := range g.nodes {
		_, newfitness := g.nodes[i].Value()
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

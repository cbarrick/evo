// The snake-in-the-box problem in graph theory deals with finding a certain
// kind of path along the edges of a hypercube. This path starts at one node
// and travels along the edges to as many nodes as it can reach. After it gets
// to a new node, the previous node and all of its neighbors must be marked as
// unusable. The path should never travel to a node after it has been marked
// unusable.
//
// In other words, a snake is a connected open path in the hypercube where each
// node in the path, with the exception of the head (start) and the tail
// (finish), has exactly two neighbors that are also in the snake. The head and
// the tail each have only one neighbor in the snake. The rule for generating a
// snake is that a node in the hypercube may be visited if it is connected to
// the current node and it is not a neighbor of any previously visited node in
// the snake, other than the current node.
//
// In graph theory terminology, this is called finding the longest possible
// induced path in a hypercube; it can be viewed as a special case of the
// induced subgraph isomorphism problem. There is a similar problem of finding
// long induced cycles in hypercubes, called the coil-in-the-box problem.
//
// The snake-in-the-box problem was first described by Kautz (1958), motivated
// by the theory of error-correcting codes. The vertices of a solution to the
// snake or coil in the box problems can be used as a Gray code that can detect
// single-bit errors. Such codes have applications in electrical engineering,
// coding theory, and computer network topologies. In these applications, it is
// important to devise as long a code as is possible for a given dimension of
// hypercube. The longer the code, the more effective are its capabilities.
//
// Finding the longest snake or coil becomes notoriously difficult as the
// dimension number increases and the search space suffers a serious
// combinatorial explosion.
//
// See https://en.wikipedia.org/wiki/Snake-in-the-box
package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/cbarrick/evo"
	"github.com/cbarrick/evo/diffusion"
)

// The type snake describes the genome of our problem.
// An important property of an n-dimensional hypercube is that each node can be
// identified with n-bits, where adjacent nodes only differ by one bit. Our
// genome is encoded as a list of edges. Each edge identifies the bit that is
// flipped between the nodes it connects. A random list of edges may not encode
// a valid snake, so we ignore any invalid edges when calculating fitness.
// We also keep track of the dimension of the hypercube.
type snake struct {
	dim   uint
	edges []byte
}

// nodes converts our edge-list representation to a node-list representation
// starting at node 0. Each edge that would create an invalid snake is skipped,
// Thus the returned node-list is of variable length and always represents a
// valid snake.
func (s *snake) nodes() (nodes []byte) {
	contains := func(list []byte, elem byte) bool {
		for i := range list {
			if list[i] == elem {
				return true
			}
		}
		return false
	}

	free := func(list []byte, elem byte) bool {
		for i := uint(0); i < s.dim; i++ {
			test := elem ^ (1 << i)
			if contains(list[:len(list)-1], test) {
				return false
			}
		}
		return true
	}

	nodes = make([]byte, 1, len(s.edges)+1)
	current := nodes[0]
	for i := range s.edges {
		next := current ^ (1 << s.edges[i])
		if uint(s.edges[i]) < s.dim && free(nodes, next) {
			nodes = append(nodes, next)
			current = next
		}
	}
	return nodes
}

// Fitness returns the number of edges in the snake
// i.e. 1 less than the number of nodes in the snake
func (s *snake) Fitness() float64 {
	return float64(len(s.nodes()) - 1)
}

// Cross performs a point-wise crossover between two snakes.
// A random edge of the child is mutated by a normally distributed random value.
func (mom *snake) Cross(suiters ...evo.Genome) evo.Genome {
	perm := rand.Perm(len(suiters))
	dad := evo.Tournament(suiters[perm[0]], suiters[perm[1]]).(*snake)

	split := rand.Intn(len(mom.edges))
	child := new(snake)
	child.dim = mom.dim
	child.edges = make([]byte, 0, len(mom.edges))
	child.edges = append(child.edges, dad.edges[:split]...)
	child.edges = append(child.edges, mom.edges[split:]...)

	i := rand.Intn(len(child.edges))
	mutation := child.edges[i] + byte(rand.NormFloat64())
	if !(mutation < 0 || uint(mutation) > child.dim) {
		child.edges[i] = mutation
	}

	if child.Fitness() >= mom.Fitness() {
		return child
	} else {
		return mom
	}
}

// Difference returns the percent difference between the edge-lists of two snakes.
func (s *snake) Difference(other evo.Genome) (score float64) {
	edges := other.(*snake).edges
	for i := range edges {
		if edges[i] != s.edges[i] {
			score++
		}
	}
	score /= float64(len(edges))
	return score
}

// Close simply returns nil. Snakes do not control any closeable resources.
func (_ *snake) Close() error {
	return nil
}

// Random creates a random snake in a given dimension.
func Random(dim uint) evo.Genome {
	if dim > 8 {
		panic(fmt.Sprint("cannot create a snake of dimension", dim))
	}

	s := new(snake)
	s.dim = dim
	s.edges = make([]byte, 1<<uint(dim-1))
	for i := range s.edges {
		s.edges[i] = byte(rand.Intn(int(dim + 1)))
	}
	return s
}

// main runs a diffusion-population genetic algorithm to search for large snakes
// of dimension 6. The largest snake is known to be 26 edges long.
func main() {
	size := 128    // controls the size of the population
	dim := uint(6) // sets the dimension of the cube in which we search
	fmt.Printf("snake-in-the-box: dimension=%d population-size=%d\n", dim, size)

	var stats evo.Stats

	// Create an initial random population
	snakes := make([]evo.Genome, size)
	for i := range snakes {
		snakes[i] = Random(dim)
	}
	population := diffusion.Hypercube(snakes)

	// update prints a status line to the terminal
	// the string "\x1b[2K" is the escape code to clear the line
	update := func() {
		stats = population.Stats()
		fmt.Printf("\x1b[2K\rMax: %f | Min: %f | Diversity: %f",
			stats.N["maxfit"],
			stats.N["minfit"],
			stats.N["diversity"])
	}

	// Stop after 5 seconds
	stop := time.After(5 * time.Second)
	done := false
	for !done {
		select {
		case _ = <-stop:
			done = true
		default:
			update()
		}
	}
	population.Close()
	update()
	fmt.Println()

	// Print the final population
	fmt.Println("Solution:")
	fmt.Println(stats.Max)
	fmt.Println()
}

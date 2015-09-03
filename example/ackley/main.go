// http://www.sfu.ca/~ssurjano/ackley.html
package main

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/cbarrick/evo"
	"github.com/cbarrick/evo/diffusion"
)

type ackley struct {
	gene []float64
}

func (ack *ackley) Fitness() (f float64) {
	var a, b float64
	a = 20
	b = 0.2

	var sum1, sum2, n float64
	n = float64(len(ack.gene))
	for _, x := range ack.gene {
		sum1 += x * x
		sum2 += math.Cos(2 * math.Pi * x)
	}

	f -= a
	f *= math.Exp(-b * math.Sqrt(sum1/n))
	f -= math.Exp(sum2 / n)
	f += a
	f += math.E
	f *= -1
	return f
}

func (mom *ackley) Cross(suiters ...evo.Genome) evo.Genome {
	perm := rand.Perm(len(suiters))
	dad := evo.Tournament(suiters[perm[0]], suiters[perm[1]]).(*ackley)

	split := rand.Intn(len(mom.gene))
	child := new(ackley)
	child.gene = make([]float64, 0, len(mom.gene))
	child.gene = append(child.gene, mom.gene[:split]...)
	child.gene = append(child.gene, dad.gene[split:]...)

	i := rand.Intn(len(child.gene))
	child.gene[i] += rand.NormFloat64()

	if child.Fitness() >= mom.Fitness() {
		return child
	} else {
		return mom
	}
}

func (_ *ackley) Close() error {
	return nil
}

func Random(dim int) (ack *ackley) {
	ack = new(ackley)
	ack.gene = make([]float64, dim)
	for i := 0; i < dim; i++ {
		ack.gene[i] = rand.Float64()*2*32.768 - 32.768
	}
	return ack
}

func main() {
	var (
		size     = 32
		dim      = 16
		accuracy = 0.001
		convergence float64
	)


	fmt.Printf("ackley: dimension=%d population=%d accuracy=%g\n", dim, size, accuracy)

	// random initial population
	// each gene is in the range [-32.768, +32.768).
	acks := make([]evo.Genome, size)
	for i := range acks {
		acks[i] = Random(dim)
	}
	population := diffusion.Hypercube(acks)

	// update sets the convergence variable
	// and prints a status line to the terminal
	// the string "\x1b[2K" is the escape code to clear the line
	update := func() {
		max := population.Max().Fitness()
		min := population.Min().Fitness()
		convergence = max - min
		fmt.Printf("\x1b[2K\rMax: %f | Min: %f | Conv: %f", max, min, convergence)
	}

	// the global maximum fitness is known to be 0 when all variables are 0
	// run the GA until the population converges to the given degree of accuracy
	update()
	for convergence > accuracy {
		update()
	}
	population.Close()
	update()
	fmt.Println()

	// print the final population
	fmt.Println("Solution:")
	fmt.Println(population.Max())
	fmt.Println()
}

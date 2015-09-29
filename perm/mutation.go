package perm

import (
	"math/rand"
)

func RandInvert(gene []int) {
	i := rand.Intn(len(gene) - 1)
	j := i + 1 + rand.Intn(len(gene) - i - 1)
	reverse(gene[i:j])
}

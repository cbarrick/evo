package integer_test

import (
	"testing"

	"github.com/cbarrick/evo/integer"
)

// cross.go
// -------------------------

func TestUniformX(t *testing.T) {
	mom := make([]int, 8)
	dad := make([]int, 8)
	for i := range mom {
		mom[i] = 1
	}
	for i := range dad {
		dad[i] = 2
	}
	child := make([]int, 8)
	integer.UniformX(child, mom, dad)
	for i := range child {
		if child[i] != mom[i] && child[i] != dad[i] {
			t.Fail()
		}
	}
}

func TestPointX(t *testing.T) {
	mom := make([]int, 8)
	dad := make([]int, 8)
	for i := range mom {
		mom[i] = 1
	}
	for i := range dad {
		dad[i] = 2
	}
	child := make([]int, 8)
	integer.PointX(7, child, mom, dad)
	for i := range child {
		if child[i] != mom[i] && child[i] != dad[i] {
			t.Fail()
		}
	}
}

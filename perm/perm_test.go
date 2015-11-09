package perm_test

import (
	"math/rand"
	"testing"

	"github.com/cbarrick/evo/perm"
)

// validate fails the test if perm is not a permutation
func validate(t *testing.T, perm []int) {
	n := len(perm)
	for i := 0; i < n; i++ {
		found := false
		for j := range perm {
			if perm[j] == i {
				found = true
				break
			}
		}
		if !found {
			t.Fail()
		}
	}
}

// cross.go
// -------------------------

func TestOrderX(t *testing.T) {
	mom := rand.Perm(8)
	dad := rand.Perm(8)
	child := make([]int, 8)
	perm.OrderX(child, mom, dad)
	validate(t, child)
}

func TestPMX(t *testing.T) {
	mom := rand.Perm(8)
	dad := rand.Perm(8)
	child := make([]int, 8)
	perm.PMX(child, mom, dad)
	validate(t, child)
}

func TestCycleX(t *testing.T) {
	mom := rand.Perm(8)
	dad := rand.Perm(8)
	child := make([]int, 8)
	perm.CycleX(child, mom, dad)
	validate(t, child)
}

func TestEdgeX(t *testing.T) {
	mom := rand.Perm(8)
	dad := rand.Perm(8)
	child := make([]int, 8)
	perm.EdgeX(child, mom, dad)
	validate(t, child)
}

// mutation.go
// -------------------------

func TestRandInvert(t *testing.T) {
	a := rand.Perm(8)
	b := make([]int, 8)
	copy(b, a)
	perm.RandInvert(b)
	flipped := false
	i, j := 0, 7
	for {
		if j <= i {
			if !flipped {
				t.Fail()
			}
			return
		} else if a[i] == b[i] {
			i++
		} else if a[j] == b[j] {
			j--
		} else {
			if flipped {
				t.Fail()
				return
			}
			perm.Reverse(b[i : j+1])
			flipped = true
		}
	}
}

func TestRandSwap(t *testing.T) {
	a := rand.Perm(8)
	b := make([]int, 8)
	copy(b, a)
	perm.RandSwap(b)
	swapped := false
	i, j := 0, 7
	for {
		if j <= i {
			if !swapped {
				t.Fail()
			}
			return
		} else if a[i] == b[i] {
			i++
		} else if a[j] == b[j] {
			j--
		} else {
			if swapped {
				t.Fail()
				return
			}
			b[i], b[j] = b[j], b[i]
			swapped = true
		}
	}
}

// util.go
// -------------------------

func TestRandSlice(t *testing.T) {
	slice := make([]int, 8)
	sub, left, right := perm.RandSlice(slice)
	sub[0] = 1
	sub[len(sub)-1] = 1
	if slice[left] != 1 || slice[right-1] != 1 {
		t.Fail()
	}
}

func TestSearch(t *testing.T) {
	slice := []int{0, 1, 2, 3, 4, 5, 6, 7}
	if perm.Search(slice, 7) != 7 {
		t.Fail()
	}
	if perm.Search(slice, 8) != -1 {
		t.Fail()
	}
}

func TestReverse(t *testing.T) {
	slice := rand.Perm(8)
	rev := make([]int, 8)
	copy(rev, slice)
	perm.Reverse(rev)
	for i, j := 0, 7; i < j; i, j = i+1, j-1 {
		if slice[i] != rev[j] {
			t.Fail()
		}
	}
}

func TestValidate(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fail()
		}
		perm.Validate([]int{0, 1, 2, 3})
	}()
	perm.Validate([]int{0, 0, 1, 2})
}

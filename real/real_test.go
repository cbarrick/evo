package real_test

import (
	"math"
	"testing"

	"github.com/cbarrick/evo"
	"github.com/cbarrick/evo/real"
)

// cross.go
// -------------------------

func TestUniformX(t *testing.T) {
	mom := real.Random(8, 1)
	dad := real.Random(8, 1)
	child := make([]float64, 8)
	real.UniformX(child, mom, dad)
	for i := range child {
		if child[i] != mom[i] && child[i] != dad[i] {
			t.Fail()
		}
	}
}

func TestArithX(t *testing.T) {
	mom := []float64{0, 0}
	dad := []float64{1, -1}
	child := []float64{0, 0}
	real.ArithX(1, child, mom, dad)
	a := 0 < child[0] && child[0] < 1
	b := -1 < child[1] && child[1] < 0
	c := child[0] == -child[1]
	if !a || !b || !c {
		t.Fail()
	}
}

// distributions.go
// -------------------------

func TestNormal(t *testing.T) {
	var s evo.Stats
	for i := 0; i < 65536; i++ {
		x := real.Normal(1e-3)
		s = s.Insert(x)
	}
	mean := s.Mean()
	if mean < -1e-3 || 1e-3 < mean || math.IsNaN(mean) {
		t.Fail()
	}
}

func TestLognormal(t *testing.T) {
	var s evo.Stats
	for i := 0; i < 65536; i++ {
		x := math.Log(real.Lognormal(1e-3))
		s = s.Insert(x)
	}
	mean := s.Mean()
	if mean < -1e-3 || 1e-3 < mean || math.IsNaN(mean) {
		t.Fail()
	}
}

// evostrat.go
// -------------------------

func TestAdapt(t *testing.T) {
	x := real.Random(8, 1)
	y := x.Copy()
	y.Adapt()
	for i := range x {
		if x[i] == y[i] {
			t.Fail()
		}
	}
}

func TestStep(t *testing.T) {
	x := make(real.Vector, 8)
	x.Step(real.Vector{1,1,1,1,1,1,1,1})
	for i := range x {
		if x[i] < -3 || 3 < x[i] {
			t.Fail()
		}
	}
}

// vector.go
// -------------------------

func TestRandom(t *testing.T) {
	x := real.Random(8, 1)
	if len(x) != 8 {
		t.Fail()
		return
	}
	for i := range x {
		if x[i] < 0 || 1 < x[i] {
			t.Fail()
		}
	}
}

func TestCopy(t *testing.T) {
	x := real.Random(8, 1)
	y := x.Copy()
	for i := range x {
		if x[i] != y[i] {
			t.Fail()
		}
	}
	x[0] = 0
	if x[0] == y[0] {
		t.Fail()
	}
}

func TestAdd(t *testing.T) {
	x := real.Random(8, 1)
	y := real.Random(8, 1)
	z := x.Copy()
	z.Add(y)
	for i := range z {
		if z[i] != x[i] + y[i] {
			t.Fail()
		}
	}
}

func TestSubtract(t *testing.T) {
	x := real.Random(8, 1)
	y := real.Random(8, 1)
	z := x.Copy()
	z.Subtract(y)
	for i := range z {
		if z[i] != x[i] - y[i] {
			t.Fail()
		}
	}
}

func TestScale(t *testing.T) {
	x := real.Random(8, 1)
	y := x.Copy()
	y.Scale(3)
	for i := range y {
		if y[i] != x[i] * 3 {
			t.Fail()
		}
	}
}

func TestHighBound(t *testing.T) {
	x := real.Vector{1,3}
	x.HighBound(2)
	if x[0] != 1 || x[1] != 2 {
		t.Fail()
	}
}

func TestLowBound(t *testing.T) {
	x := real.Vector{1,3}
	x.LowBound(2)
	if x[0] != 2 || x[1] != 3 {
		t.Fail()
	}
}

func TestBound(t *testing.T) {
	x := real.Vector{1,4}
	x.Bound(real.Vector{2,2}, real.Vector{3,3})
	if x[0] != 2 || x[1] != 3 {
		t.Fail()
	}
}

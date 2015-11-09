package evo_test

import (
	"testing"

	"github.com/cbarrick/evo"
)

func TestMerge(t *testing.T) {
	var a, b evo.Stats
	for i := float64(0); i < 5; i++ {
		a = a.Put(i)
	}
	for i := float64(5); i < 10; i++ {
		b = b.Put(i)
	}
	stats := a.Merge(b)
	if stats.Mean() != 4.5 {
		t.Fail()
	}
	if stats.Var() != 8.25 {
		t.Fail()
	}
}

func TestMax(t *testing.T) {
	stats := data()
	if stats.Max() != 855 {
		t.Fail()
	}
}

func TestMin(t *testing.T) {
	stats := data()
	if stats.Min() != 760 {
		t.Fail()
	}
}

func TestRange(t *testing.T) {
	stats := data()
	if stats.Range() != 95 {
		t.Fail()
	}
}

func TestMean(t *testing.T) {
	stats := data()
	if stats.Mean() < 810.1388888 || 810.1388890 < stats.Mean() {
		t.Fail()
	}
}

func TestVar(t *testing.T) {
	stats := data()
	if stats.Var() < 829.841820 || 829.841822 < stats.Var() {
		t.Fail()
	}
}

func TestSD(t *testing.T) {
	stats := data()
	if stats.SD() < 28.80697520 || 28.80697522 < stats.SD() {
		t.Fail()
	}
}

func TestRSD(t *testing.T) {
	stats := data()
	if stats.RSD() < 0.03555806986 || 0.03555806988 < stats.RSD() {
		t.Fail()
	}
}

func TestCount(t *testing.T) {
	stats := data()
	if stats.Count() != 36 {
		t.Fail()
	}
}

func data() (s evo.Stats) {
	s = s.Put(810)
	s = s.Put(820)
	s = s.Put(820)
	s = s.Put(840)
	s = s.Put(840)
	s = s.Put(845)
	s = s.Put(785)
	s = s.Put(790)
	s = s.Put(785)
	s = s.Put(835)
	s = s.Put(835)
	s = s.Put(835)
	s = s.Put(845)
	s = s.Put(855)
	s = s.Put(850)
	s = s.Put(760)
	s = s.Put(760)
	s = s.Put(770)
	s = s.Put(820)
	s = s.Put(820)
	s = s.Put(820)
	s = s.Put(820)
	s = s.Put(820)
	s = s.Put(825)
	s = s.Put(775)
	s = s.Put(775)
	s = s.Put(775)
	s = s.Put(825)
	s = s.Put(825)
	s = s.Put(825)
	s = s.Put(815)
	s = s.Put(825)
	s = s.Put(825)
	s = s.Put(770)
	s = s.Put(760)
	s = s.Put(765)
	return s
}

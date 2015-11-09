package evo_test

import (
	"testing"

	"github.com/cbarrick/evo"
)

func TestMerge(t *testing.T) {
	var a, b evo.Stats
	for i := float64(0); i < 5; i++ {
		a = a.Insert(i)
	}
	for i := float64(5); i < 10; i++ {
		b = b.Insert(i)
	}
	stats := a.Merge(b)
	if stats.Mean() != 4.5 {
		t.Fail()
	}
	if stats.Variance() != 8.25 {
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

func TestVariance(t *testing.T) {
	stats := data()
	if stats.Variance() < 829.841820 || 829.841822 < stats.Variance() {
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

func TestLen(t *testing.T) {
	stats := data()
	if stats.Len() != 36 {
		t.Fail()
	}
}

func data() (s evo.Stats) {
	s = s.Insert(810)
	s = s.Insert(820)
	s = s.Insert(820)
	s = s.Insert(840)
	s = s.Insert(840)
	s = s.Insert(845)
	s = s.Insert(785)
	s = s.Insert(790)
	s = s.Insert(785)
	s = s.Insert(835)
	s = s.Insert(835)
	s = s.Insert(835)
	s = s.Insert(845)
	s = s.Insert(855)
	s = s.Insert(850)
	s = s.Insert(760)
	s = s.Insert(760)
	s = s.Insert(770)
	s = s.Insert(820)
	s = s.Insert(820)
	s = s.Insert(820)
	s = s.Insert(820)
	s = s.Insert(820)
	s = s.Insert(825)
	s = s.Insert(775)
	s = s.Insert(775)
	s = s.Insert(775)
	s = s.Insert(825)
	s = s.Insert(825)
	s = s.Insert(825)
	s = s.Insert(815)
	s = s.Insert(825)
	s = s.Insert(825)
	s = s.Insert(770)
	s = s.Insert(760)
	s = s.Insert(765)
	return s
}

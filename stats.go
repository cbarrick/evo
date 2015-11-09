package evo

import (
	"fmt"
	"math"
)

// A Stats object is a statistics collector. A common source of Stats objects is
// the return value of Population.Stats() which gives statistics about the
// fitness of genomes in the population.
type Stats struct {
	max, min float64
	mean     float64
	sumsq    float64 // sum of squares of deviation from the mean
	len      float64
}

// Insert inserts a new value into the data.
func (s Stats) Insert(x float64) Stats {
	if s.len == 0 {
		s.max = math.Inf(-1)
		s.min = math.Inf(+1)
	}

	delta := x - s.mean
	newlen := s.len + 1

	// max & min
	s.max = math.Max(s.max, x)
	s.min = math.Min(s.min, x)

	// mean
	s.mean += delta / newlen

	// sum of squares
	s.sumsq += delta * delta * (s.len / newlen)

	// len
	s.len = newlen

	return s
}

// Merge merges the data of two Stats objects.
func (s Stats) Merge(t Stats) Stats {
	if s.len == 0 {
		s.max = math.Inf(-1)
		s.min = math.Inf(+1)
	}

	delta := t.mean - s.mean
	newlen := t.len + s.len

	// max & min
	s.max = math.Max(s.max, t.max)
	s.min = math.Min(s.min, t.min)

	// mean
	s.mean += delta * (t.len / newlen)

	// sum of squares
	s.sumsq += t.sumsq
	s.sumsq += delta * delta * (t.len * s.len / newlen)

	// len
	s.len = newlen

	return s
}

// Max returns the maximum data point.
func (s Stats) Max() float64 {
	return s.max
}

// Min returns the minimum data point.
func (s Stats) Min() float64 {
	return s.min
}

// Range returns the difference in the maximum and minimum data points.
func (s Stats) Range() float64 {
	return s.max - s.min
}

// Mean returns the average of the data.
func (s Stats) Mean() float64 {
	return s.mean
}

// Variance returns the population variance of the data.
func (s Stats) Variance() float64 {
	return s.sumsq / s.len
}

// SD returns the standard deviation of the data.
func (s Stats) SD() float64 {
	return math.Sqrt(s.sumsq / s.len)
}

// RSD returns the relative standard deviation of the data, also known as the
// coefficient of variation.
func (s Stats) RSD() float64 {
	return s.SD() / s.Mean()
}

// Len returns the size of the data.
func (s Stats) Len() int {
	return int(s.len)
}

// String returns a string listing a summary of the statistics.
func (s Stats) String() string {
	return fmt.Sprintf("Max: %f | Min: %f | SD: %f",
		s.Max(),
		s.Min(),
		s.SD())
}

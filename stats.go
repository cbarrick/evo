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
	sum      float64
	sumsq    float64
	count    float64
}

// Put inserts a new value into the data.
func (s *Stats) Put(x float64) Stats {
	if x > s.max || s.count == 0 {
		s.max = x
	}
	if x < s.min || s.count == 0 {
		s.min = x
	}

	s.sum += x
	s.sumsq += x * x
	s.count++

	return *s
}

// Merge merges the data of two Stats objects.
func (s *Stats) Merge(t Stats) Stats {

	// Do not merge if t is empty.
	if t.count == 0 {
		return *s
	}

	s.count += t.count
	s.sum += t.sum
	s.sumsq += t.sumsq

	// The OR clause is used in case s is empty (max and min will be 0)
	// and t.max < 0 or t.min > 0.
	if t.max > s.max || s.count == 0 {
		s.max = t.max
	}
	if t.min > s.min || s.count == 0 {
		s.min = t.min
	}

	return *s
}

// Max returns the maximum data point.
func (s *Stats) Max() float64 {
	return s.max
}

// Min returns the minimum data point.
func (s *Stats) Min() float64 {
	return s.min
}

// Range returns the difference in the maximum and minimum data points.
func (s *Stats) Range() float64 {
	return s.max - s.min
}

// Mean returns the average of the data.
func (s *Stats) Mean() float64 {
	if s.count == 0 {
		return 0
	}
	return s.sum / s.count
}

// Var returns the population variance of the data calculated in a single-
// pass method.
func (s *Stats) Var() float64 {
	if s.count < 2 {
		return 0 // no variation
	}
	return (s.count*s.sumsq - s.sum*s.sum) / (s.count * s.count)
}

// SD returns the population standard deviation of the data.
func (s *Stats) SD() float64 {
	if s.count < 2 {
		return 0 // no variation
	}
	return math.Sqrt(s.Var())
}

// RSD returns the population relative standard deviation of the data, also
// known as the coefficient of variation.
func (s *Stats) RSD() float64 {
	if s.count == 0 {
		return 0
	}
	return s.SD() / s.Mean()
}

// Count returns the size of the data.
func (s *Stats) Count() int {
	return int(s.count)
}

// String returns a string listing a summary of the statistics.
func (s *Stats) String() string {
	return fmt.Sprintf("Max: %f | Min: %f | SD: %f",
		s.Max(),
		s.Min(),
		s.SD())
}

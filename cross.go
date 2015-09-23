package evo

import (
	"math/rand"
)

func PMX(p1, p2 []int) (c1, c2 []int) {
	c1 = make([]int, len(p1))
	c2 = make([]int, len(p1))

	search := func(slice []int, val int) (idx int) {
		for idx = range slice {
			if slice[idx] == val {
				return idx
			}
		}
		return -1
	}

	// pick a random range
	left := rand.Intn(len(p1)-1)
	right := left
	for right == left {
		right = rand.Intn(len(p1))
	}
	if right < left {
		left, right = right, left
	}

	// produce c1
	for i := range c1 {
		c1[i] = -1
	}
	copy(c1[left:right], p1[left:right])
	for i := left; i < right; i++ {
		if search(c1, p2[i]) == -1 {
			j := i
			for left <= j && j < right {
				j = search(p2, p1[j])
			}
			c1[j] = p2[i]
		}
	}
	for i := range c1 {
		if c1[i] == -1 {
			c1[i] = p2[i]
		}
	}

	// produce c2
	for i := range c2 {
		c2[i] = -1
	}
	copy(c2[left:right], p2[left:right])
	for i := left; i < right; i++ {
		if search(c2, p1[i]) == -1 {
			j := i
			for left <= j && j < right {
				j = search(p1, p2[j])
			}
			c2[j] = p1[i]
		}
	}
	for i := range c2 {
		if c2[i] == -1 {
			c2[i] = p1[i]
		}
	}

	return c1, c2
}

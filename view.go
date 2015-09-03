package evo

// The purpose of a View is to inspect the contents of a Population in a
// thread-safe way. Views are typically encountered when calling
// Population.View(). Views are simply a named type for []Genome that
// implement Population.
type View []Genome

// Max returns the member of the view with the highest fitness.
func (v View) Max() (max Genome) {
	max = v[0]
	for i := range v {
		if v[i].Fitness() > max.Fitness() {
			max = v[i]
		}
	}
	return max
}

// Min returns the member of the view with the lowest fitness.
func (v View) Min() (min Genome) {
	min = v[0]
	for i := range v {
		if v[i].Fitness() < min.Fitness() {
			min = v[i]
		}
	}
	return min
}

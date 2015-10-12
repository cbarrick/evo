package sel

import (
	"github.com/cbarrick/evo"
)

type Elite struct {
	in    chan evo.Genome
	out   chan evo.Genome
	close chan chan struct{}
}

func NewElite(popsize, poolsize uint) Elite {
	var e Elite
	e.in = make(chan evo.Genome)
	e.out = make(chan evo.Genome, popsize)
	e.close = make(chan chan struct{})
	go e.run(popsize, poolsize)
	return e
}

func (e Elite) run(popsize, poolsize uint) {
	type item struct {
		val  evo.Genome
		fit  float64
		len  uint
		next *item
	}

	var (
		best   *item
		window = poolsize - popsize
	)

	for {
		select {
		case ch := <-e.close:
			ch <- struct{}{}
			return

		case val := <-e.in:
			newitem := &item{
				val: val,
				fit: val.Fitness(),
			}

			if best == nil {
				best = newitem
				best.len = 1
			} else if best.fit < newitem.fit {
				newitem.len = best.len + 1
				newitem.next = best
				best = newitem
			} else {
				for i := best; true; i = i.next {
					i.len++
					if i.next == nil {
						newitem.len = 1
						i.next = newitem
						break
					} else if i.next.fit < newitem.fit {
						newitem.next = i.next
						newitem.len = i.len - 1
						i.next = newitem
						break
					}
				}
			}

		default:
			if best != nil && best.len >= window {
				e.out <- best.val
				best = best.next
				poolsize--

				// Once the poolsize reaches the window,
				// we reset for the next generation.
				if poolsize == window {
					best = nil
					poolsize = window + popsize
				}
			}
		}
	}
}

func (e Elite) Put(val evo.Genome) {
	e.in <- val
}

func (e Elite) Get() (val evo.Genome) {
	val = <-e.out
	return val
}

func (e Elite) Close() {
	ch := make(chan struct{})
	e.close <- ch
	<-ch
}

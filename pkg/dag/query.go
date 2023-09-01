package dag

import (
	"math/rand"

	"github.com/cockroachdb/errors"
)

type Result[T INode] []T

func (q Result[T]) All() []T {
	return q
}

// Pick a random selection of exactly n elements. Panics if there are less than
// n total elements.
func (q Result[T]) Pick(n int) []T {
	if len(q) < n {
		panic(errors.Newf("Pick(%d) called on Result with %d elements", n, len(q)))
	}
	if len(q) == n {
		return q.All()
	}
	picked := map[int]bool{}
	for len(picked) < n {
		picked[rand.Intn(n)] = true
	}

	var out []T
	for i := range q {
		if picked[i] {
			out = append(out, q[i])
		}
	}
	return out
}

func (q Result[T]) Filter(predicates ...func(T) bool) []T {
	pred := func(n INode) bool {
		t, ok := n.(T)
		if !ok {
			return false
		}

		for _, p := range predicates {
			if !p(t) {
				return false
			}
		}
		return true
	}

	var out []T
	for _, n := range q {
		if pred(n) {
			out = append(out, n)
		}
	}
	return out
}

func (q Result[T]) PickUpTo(n int) []T {
	return q.PickBetween(1, n)
}

func (q Result[T]) PickBetween(min, max int) []T {
	if len(q) < max {
		max = len(q)
	}
	return q.Pick(min + rand.Intn(max-min))
}

func (q Result[T]) One() T {
	if len(q) != 1 {
		panic("TODO")
	}
	return q[0]
}

func (q Result[T]) Any() T {
	return q.Pick(1)[0]
}

func ByID[T INode](g *Graph, id string) T {
	return g.nodesByID[id].(T)
}

func Nodes[T INode](g *Graph, predicates ...Filter[T]) Result[T] {
	return filter[T](g.nodes, predicates...)
}

func Incoming[T INode](n INode, predicates ...Filter[T]) Result[T] {
	return filter[T](n.graph().incoming[n])
}

func Outgoing[T INode](n INode, predicates ...Filter[T]) Result[T] {
	return filter(
		n.graph().outgoing[n],
		predicates...,
	)
}

// func All[T INode](g *Graph, predicates ...func(T) bool) []T {
// 	pred := func(n INode) bool {
// 		t, ok := n.(T)
// 		if !ok {
// 			return false
// 		}
// 		for _, p := range predicates {
// 			if !p(t) {
// 				return false
// 			}
// 		}
// 		return true
// 	}
//
// 	var out []T
// 	for _, n := range g.nodes {
// 		if pred(n) {
// 			out = append(out, n.(T))
// 		}
// 	}
// 	return out
// }

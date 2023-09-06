package dag

type Filter[T INode] func(T) bool

func Any[T INode](g *Graph, predicates ...Filter[T]) T {
	return Nodes[T](g, predicates...).Any()
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

func filter[T INode](in []INode, predicates ...Filter[T]) []T {
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
	for _, n := range in {
		if pred(n) {
			out = append(out, n.(T))
		}
	}
	return out
}

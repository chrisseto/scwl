package dag

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

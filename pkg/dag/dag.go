package dag

type INode interface {
	setGraph(*Graph)
	Graph() *Graph
}

type Node struct {
	g *Graph
}

func (n *Node) setGraph(g *Graph) {
	n.g = g
}

func (n *Node) Graph() *Graph {
	return n.g
}

func New() *Graph {
	return &Graph{
		nodesByID: make(map[string]INode),
		outgoing:  make(map[INode][]INode),
		incoming:  make(map[INode][]INode),
	}
}

type Graph struct {
	nodes     []INode
	nodesByID map[string]INode
	outgoing  map[INode][]INode
	incoming  map[INode][]INode
}

func (g *Graph) AddNode(id string, n INode) {
	n.setGraph(g)
	g.nodes = append(g.nodes, n)
	g.nodesByID[id] = n
}

func (g *Graph) AddEdge(from, to string) {
	fromNode := g.nodesByID[from]
	toNode := g.nodesByID[to]
	g.outgoing[fromNode] = append(g.outgoing[fromNode], toNode)
	g.incoming[toNode] = append(g.incoming[toNode], fromNode)
}

type Result[T INode] []T

func (q Result[T]) All() []T {
	return q
}

func (q Result[T]) One() T {
	if len(q) != 1 {
		panic("TODO")
	}
	return q[0]
}

func Any[T INode](g *Graph, predicates ...func(T) bool) T {
	options := All(g, predicates...)
	return options[0]
}

func Incoming[T INode](n INode, predicates ...Filter[T]) Result[T] {
	return filter[T](n.Graph().incoming[n])
}

func Outgoing[T INode](n INode, predicates ...Filter[T]) Result[T] {
	return filter(
		n.Graph().outgoing[n],
		predicates...,
	)
}

func All[T INode](g *Graph, predicates ...func(T) bool) []T {
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
	for _, n := range g.nodes {
		if pred(n) {
			out = append(out, n.(T))
		}
	}
	return out
}

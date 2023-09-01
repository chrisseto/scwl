package dag

import (
	// "bytes"
	"bytes"
	"fmt"
	"reflect"
	// "strings"
	// "golang.org/x/exp/maps"
)

type INode interface {
	graph() *Graph
	setGraph(*Graph)
}

type Node struct {
	g *Graph `json:"-"`
}

func (n *Node) graph() *Graph {
	return n.g
}

func (n *Node) setGraph(g *Graph) {
	n.g = g
}

func New(clone func(INode) INode) *Graph {
	return &Graph{
		clone: func(i INode) INode {
			c := clone(i)
			c.setGraph(nil)
			return c
		},
		nodesByID: make(map[string]INode),
		outgoing:  make(map[INode][]INode),
		incoming:  make(map[INode][]INode),
	}
}

type Graph struct {
	clone func(INode) INode

	// TODO add RWMutext
	nodes     []INode
	nodesByID map[string]INode
	outgoing  map[INode][]INode
	incoming  map[INode][]INode
}

func (g *Graph) ByID(id string) INode {
	return g.nodesByID[id]
}

func (g *Graph) AddNode(id string, n INode) {
	n.setGraph(g)
	g.nodes = append(g.nodes, n)
	g.nodesByID[id] = n
}

func (g *Graph) AddEdge(from, to INode) {
	// TODO assert same graph
	g.incoming[to] = append(g.incoming[to], from)
	g.outgoing[from] = append(g.outgoing[from], to)
}

func (g *Graph) String() string {
	var b bytes.Buffer

	nodes := make(map[INode]int, len(g.nodes))
	// nodes := make([]string, len(g.nodes))
	for i, n := range g.nodes {
		nodes[n] = i
		v := reflect.ValueOf(n).Elem()
		t := v.Type()
		fmt.Fprintf(&b, "%d: %s{", i, t.Name())
		for j := 0; j < t.NumField(); j++ {
			f := t.Field(j)
			if f.Anonymous {
				continue
			}
			if j > 1 {
				fmt.Fprint(&b, ", ")
			}
			fmt.Fprintf(&b, "%s: %v", f.Name, v.Field(j))
		}
		fmt.Fprint(&b, "}\n")
	}

	for _, from := range g.nodes {
		for _, to := range g.outgoing[from] {
			// fromT := reflect.TypeOf()
			fmt.Fprintf(&b, "%T(%d) -> %T(%d)\n", from, nodes[from], to, nodes[to])
		}
	}
	return b.String()
	// return strings.Join(nodes, "\n")
}

// func (g *Graph) Clone() *Graph {
//
// }

type Edge struct {
	From INode
	To   INode
}

type Comparable struct {
	Nodes []INode
	Edges [][2]int
}

type CNode struct {
	Node     INode
	Outgoing []INode
	Incoming []INode
}

func (g *Graph) Comparable() []CNode {
	nodes := make([]INode, len(g.nodes))
	nodeToIndex := make(map[INode]int, len(g.nodes))
	cloneToNode := make(map[INode]INode, len(g.nodes))
	nodeToClone := make(map[INode]INode, len(g.nodes))

	for i := range g.nodes {
		nodes[i] = g.clone(g.nodes[i])
		nodeToIndex[nodes[i]] = i
		cloneToNode[nodes[i]] = g.nodes[i]
		nodeToClone[g.nodes[i]] = nodes[i]
	}

	cnodes := make([]CNode, len(g.nodes))
	for i := range nodes {
		n := nodes[i]

		outgoing := make([]INode, len(g.outgoing[n]))
		for j, out := range g.outgoing[n] {
			outgoing[j] = out
		}

		incoming := make([]INode, len(g.incoming[n]))
		for j, in := range g.incoming[n] {
			incoming[j] = in
		}

		cnodes[i] = CNode{
			Node:     n,
			Outgoing: outgoing,
			Incoming: incoming,
		}

		// for _, to := range g.outgoing[cloneToNode[n]] {
		// 	edges = append(edges, [2]int{
		// 		i,
		// 		nodeToIndex[nodeToClone[to]],
		// 	})
		// }
	}

	return cnodes
}

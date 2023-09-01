package dag_test

import (
	"testing"

	"github.com/chrisseto/scwl/pkg/dag"
	"github.com/stretchr/testify/require"
)

type Person struct {
	dag.Node
	Name string
}

type Cat struct {
	dag.Node
	Name string
}

type Dog struct {
	dag.Node
	Name string
}

func TestDag(t *testing.T) {
	g := dag.New()

	bob := &Person{Name: "bob"}
	june := &Cat{Name: "alice"}
	alice := &Person{Name: "alice"}

	g.AddNode("bob", bob)
	g.AddNode("june", june)
	g.AddNode("alice", alice)
	g.AddEdge(bob, alice)
	g.AddEdge(alice, june)

	require.Equal(t, []dag.INode{bob, june, alice}, dag.Nodes[dag.INode](g).All())
	require.Equal(t, []*Person{bob, alice}, dag.Nodes[*Person](g).All())
	require.Equal(t, []*Cat{june}, dag.Nodes[*Cat](g).All())
	require.Equal(t, []*Dog(nil), dag.Nodes[*Dog](g).All())

	require.Equal(t, []dag.INode(nil), dag.Incoming[dag.INode](bob).All())
	require.Equal(t, []dag.INode{bob}, dag.Incoming[dag.INode](alice).All())
	require.Equal(t, []dag.INode{alice}, dag.Incoming[dag.INode](june).All())

	require.Equal(t, []dag.INode{alice}, dag.Outgoing[dag.INode](bob).All())
	require.Equal(t, []dag.INode{june}, dag.Outgoing[dag.INode](alice).All())
	require.Equal(t, []dag.INode(nil), dag.Outgoing[dag.INode](june).All())

	require.Equal(t, dag.Comparable{
		Nodes: []dag.INode{bob, june, alice},
		Edges: [][2]int{
			{0, 2}, // bob -> alice
			{2, 1}, // alice -> june
		},
	}, g.Comparable())
}

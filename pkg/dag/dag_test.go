package dag_test

import (
	"testing"

	"github.com/chrisseto/scwl/pkg/dag"
	"github.com/cockroachdb/errors"
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
	g := dag.New(nil)

	bob := &Person{Name: "bob"}
	june := &Cat{Name: "june"}
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
}

// func TestString(t *testing.T) {
// 	g := dag.New(nil)
//
// 	bob := &Person{Name: "bob"}
// 	june := &Cat{Name: "june"}
// 	alice := &Person{Name: "alice"}
//
// 	g.AddNode("bob", bob)
// 	g.AddNode("june", june)
// 	g.AddNode("alice", alice)
// 	g.AddEdge(bob, alice)
// 	g.AddEdge(alice, june)
//
// 	require.Equal(t, `n_0 := g.AddNode("", &Person{Name: "bob"})
// 	n_1 := g.AddNode(&Person{Name: "alice"})
//
// 	g.AddEdge(g.ByID("1"), g.ByID("2"))
// 	`, g.String())
// }

func clone(in dag.INode) dag.INode {
	// Go is weird. We can't clone a struct that implements an interface
	// without doing some reflection magic. Instead, opted for a repetitive
	// type switch. Could probably code gen it.
	switch n := in.(type) {
	case *Person:
		o := *n
		return &o
	case *Cat:
		o := *n
		return &o
	case *Dog:
		o := *n
		return &o
	default:
		panic(errors.Newf("unhandled type %T", in))
	}
}

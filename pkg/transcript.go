package pkg

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/chrisseto/scwl/pkg/dag"
	"github.com/cockroachdb/errors"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

type Transcript struct {
	Steps []Step
}

func (t *Transcript) Run(ctx context.Context, sys System) error {
	for _, step := range t.Steps {
		if err := sys.Execute(ctx, step.Command); err != nil {
			return err // TODO Annotate
		}

		state, err := sys.State(ctx)
		if err != nil {
			return err // TODO Annotate
		}

		opts := []cmp.Option{
			cmpopts.IgnoreTypes(dag.Node{}),
			cmp.Transformer("Comparable", func(g *dag.Graph) []dag.CNode {
				return g.Comparable()
			}),
		}

		if diff := cmp.Diff(step.Expected, state, opts...); diff != "" {
			return errors.Newf("state mismatch: %s", diff)
		}
	}

	return nil
}

type Step struct {
	Command  Command
	Expected *dag.Graph
}

func WithFullQualifiedName[T dag.INode](name string) dag.Filter[T] {
	return func(i T) bool {
		return FullyQualifiedName(i) == name
	}
}

func FullyQualifiedName(el dag.INode) string {
	switch n := el.(type) {
	case *Database:
		return fmt.Sprintf("%s", n.Name)
	case *Schema:
		return FullyQualifiedName(n.Database()) + fmt.Sprintf(".%s", n.Name)
	case *Table:
		return FullyQualifiedName(n.Schema()) + fmt.Sprintf(".%s", n.Name)
	case *Column:
		return FullyQualifiedName(n.Table()) + fmt.Sprintf(".cols.%s", n.Name)
	case *Index:
		return FullyQualifiedName(n.Table()) + fmt.Sprintf(".idxs.%s", n.Name)
	default:
		panic(errors.Newf("unhandled type: %T", el))
	}
}

func CommandToString(cmd Command) string {
	var b strings.Builder
	val := reflect.ValueOf(cmd)
	t := val.Type()

	fmt.Fprintf(&b, "pkg.%s{", val.Type().Name())
	for i := 0; i < val.NumField(); i++ {
		field := t.Field(i)

		if i > 0 {
			fmt.Fprintf(&b, ", ")
		}

		switch v := val.Field(i).Interface().(type) {
		case dag.Node:
			continue

		case dag.INode:
			fmt.Fprintf(&b, "%s: ByFQN(g, %q)", field.Name, FullyQualifiedName(v))

		case string:
			fmt.Fprintf(&b, "%s: %q", field.Name, v)

		default:
			fmt.Fprintf(&b, "%s: %v", field.Name, v)
		}
	}
	b.WriteRune('}')
	return b.String()
}

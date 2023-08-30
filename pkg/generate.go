package pkg

import (
	"fmt"
	"math/rand"
	"reflect"
)

var Weights = map[reflect.Type]int{
	reflect.TypeOf(CreateIndex{}):  3,
	reflect.TypeOf(AddColumn{}):    3,
	reflect.TypeOf(CreateTable{}):  2,
	reflect.TypeOf(DropDatabase{}): 0,
	reflect.TypeOf(DropColumn{}):   0,
	reflect.TypeOf(DropTable{}):    0,
	reflect.TypeOf(DropSchema{}):   0,
}

func GenerateCommand(s *Root) Command {
	var cmds []Command

	// TODO remove parents, all reference-able nodes can be reached via
	// properties.
	Walk(s, func(sn StateNode, parents []StateNode) {
		switch n := sn.(type) {
		case *Root:
			cmds = append(cmds, CreateDatabase{
				Name: RandomString(),
			})

		case *Database:
			cmds = append(
				cmds,
				DropDatabase{Database: n},
				// RenameDatabase{Database: n.Name, Name: RandomString()},
				CreateSchema{Database: n, Name: RandomString()},
			)

		case *Schema:
			if n.Name != "public" {
				cmds = append(
					cmds,
					DropSchema{Schema: n},
				)
			}

			cmds = append(
				cmds,
				CreateTable{Schema: n, Name: RandomString()},
				// RenameSchema{Schema: n, Name: RandomString()},
			)

		case *Table:
			cmds = append(
				cmds,
				DropTable{Table: n},
				AddColumn{Table: n, Name: RandomString()},
				// RenameTable{Table: n, Name: RandomString()},
			)

			if len(n.Columns) > 1 {
				cmds = append(
					cmds,
					CreateIndex{
						Table:   n,
						Columns: []*Column{n.Columns[0]},
						Name:    RandomString(),
						Unique:  true, // Hard coded for now.
					},
				)

				// This is where a DAG structure really starts to shine
				// through. In theory, we could write a graph query to find any
				// two tables with at least 1 column.
				for _, other := range n.Schema.Tables {
					// Maybe this is fine though?
					if other == n {
						continue
					}
					for _, index := range other.Indexes {
						if !index.Unique {
							continue
						}

						cmds = append(cmds, CreateForeignKeyConstraint{
							From: n.Columns[0],
							To:   index.Columns[0],
							Name: RandomString(),
						})
					}
				}
			}

		case *Column:
			cmds = append(
				cmds,
				DropColumn{Column: n},
			)

		case *ForeignKeyConstraint:
			cmds = append(cmds, DropForeignKeyConstraint{ForeignKeyConstraint: n})

		case *Index:
			cmds = append(cmds, DropIndex{Index: n})

		default:
			panic(fmt.Sprintf("Unhandled Type: %T", sn))

		}
	})

	// TODO this is pretty memory hungry, there's certainly a better way to do
	// this. I'm just a bit lazy.
	// https://en.wikipedia.org/wiki/Alias_method
	var weighted []Command
	for _, cmd := range cmds {
		weight, ok := Weights[reflect.TypeOf(cmd)]
		if !ok {
			weight = 1
		}
		for i := 0; i < weight; i++ {
			weighted = append(weighted, cmd)
		}
	}

	return weighted[rand.Intn(len(weighted))]
}

func Walk(n StateNode, cb func(StateNode, []StateNode)) {
	var doWalk func(StateNode, []StateNode)
	doWalk = func(n StateNode, stack []StateNode) {
		cb(n, stack)
		for _, c := range n.Children() {
			doWalk(c, append([]StateNode{n}, stack...))
		}
	}
	doWalk(n, nil)
}

package pkg

import (
	"fmt"
	"math/rand"
	"reflect"
)

var Weights = map[reflect.Type]int{
	reflect.TypeOf(AddColumn{}):    5,
	reflect.TypeOf(CreateTable{}):  2,
	reflect.TypeOf(DropDatabase{}): 0,
	reflect.TypeOf(DropColumn{}):   0,
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
				// RenameSchema{
				// 	Database: parents[0].(*Database).Name,
				// 	Schema:   n.Name,
				// 	Name:     RandomString(),
				// },
				CreateTable{Schema: n, Name: RandomString()},
			)

		case *Table:
			cmds = append(
				cmds,
				DropTable{Table: n},
				AddColumn{Table: n, Name: RandomString()},
				// RenameTable{
				// 	Database: parents[1].(*Database).Name,
				// 	Schema:   parents[0].(*Schema).Name,
				// 	Table:    n.Name,
				// 	Name:     RandomString(),
				// },
			)

		case *Column:
			cmds = append(
				cmds,
				DropColumn{Column: n},
			)

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

package pkg

import (
	"fmt"
	"math/rand"
)

func Filter[T any](in []T, filter func(T) bool) []T {
	out := make([]T, 0, len(in))
	for i := range in {
		if filter(in[i]) {
			out = append(out, in[i])
		}
	}
	return out
}

func GenerateCommand(s *Root) Command {
	var cmds []Command

	Walk(s, func(sn StateNode, parents []StateNode) {
		switch n := sn.(type) {
		case *Root:
			cmds = append(cmds, CreateDatabase{
				Name: RandomString(),
			})

		case *Database:
			cmds = append(
				cmds,
				DropDatabase{Name: n.Name},
				// RenameDatabase{Database: n.Name, Name: RandomString()},
				CreateSchema{Database: n.Name, Name: RandomString()},
			)

		case *Schema:
			cmds = append(
				cmds,
				DropSchema{
					Database: parents[0].(*Database).Name,
					Name:     n.Name,
				},
				// RenameSchema{
				// 	Database: parents[0].(*Database).Name,
				// 	Schema:   n.Name,
				// 	Name:     RandomString(),
				// },
				CreateTable{
					Database: parents[0].(*Database).Name,
					Schema:   n.Name,
					Name:     RandomString(),
				},
			)

		case *Table:
			cmds = append(
				cmds,
				DropTable{
					Database: parents[1].(*Database).Name,
					Schema:   parents[0].(*Schema).Name,
					Name:     n.Name,
				},
				// RenameTable{
				// 	Database: parents[1].(*Database).Name,
				// 	Schema:   parents[0].(*Schema).Name,
				// 	Table:    n.Name,
				// 	Name:     RandomString(),
				// },
			)

		default:
			panic(fmt.Sprintf("Unhandled Type: %T", sn))

		}
	})

	cmds = Filter(cmds, func(sn Command) bool {
		switch sn.(type) {
		case DropSchema, DropDatabase:
			return false
		}
		return true
	})

	return cmds[rand.Intn(len(cmds))]
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

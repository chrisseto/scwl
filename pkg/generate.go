package pkg

import (
	"log"
	"math/rand"
	"reflect"

	"github.com/chrisseto/scwl/pkg/dag"
	"github.com/cockroachdb/errors"
)

func NotPublic(s *Schema) bool {
	return s.Name != "public"
}

func init() {
	// Linting, assert that there's a generator for every Command.
	for _, cmd := range AllCommands {
		if _, ok := Generators[reflect.TypeOf(cmd)]; !ok {
			panic(errors.Newf("missing Generator for %T", cmd))
		}
	}
}

var Weights = map[reflect.Type]int{
	reflect.TypeOf(CreateIndex{}):  3,
	reflect.TypeOf(AddColumn{}):    3,
	reflect.TypeOf(CreateTable{}):  2,
	reflect.TypeOf(DropDatabase{}): 0,
	reflect.TypeOf(DropColumn{}):   0,
	reflect.TypeOf(DropTable{}):    0,
	reflect.TypeOf(DropSchema{}):   0,
}

// TODO: There's probably no reason to use reflect.TypeOf here. Command is
// fine.
var Generators = map[reflect.Type]func(*dag.Graph) Command{
	// DROP ... https://github.com/cockroachdb/cockroach/blob/master/pkg/sql/parser/sql.y#L5447-L5455
	reflect.TypeOf(DropDatabase{}): func(g *dag.Graph) Command { return DropDatabase{dag.Nodes[*Database](g).Any()} },
	reflect.TypeOf(DropIndex{}):    func(g *dag.Graph) Command { return DropIndex{dag.Nodes[*Index](g).Any()} },
	reflect.TypeOf(DropSchema{}):   func(g *dag.Graph) Command { return DropSchema{dag.Nodes[*Schema](g).Any()} },
	reflect.TypeOf(DropTable{}):    func(g *dag.Graph) Command { return DropTable{dag.Nodes[*Table](g).Any()} },

	// CREATE ... https://github.com/cockroachdb/cockroach/blob/master/pkg/sql/parser/sql.y#L5058-L5070
	reflect.TypeOf(CreateDatabase{}): func(g *dag.Graph) Command {
		// Limit total number of databases to 5.
		if len(dag.Nodes[*Database](g)) > 5 {
			return nil
		}
		return CreateDatabase{Name: RandomString()}
	},
	reflect.TypeOf(CreateSchema{}): func(g *dag.Graph) Command {
		// Limit total number of schemas to 5.
		if len(dag.Nodes[*Schema](g)) > 2 {
			return nil
		}
		return CreateSchema{Database: dag.Nodes[*Database](g).Any(), Name: RandomString()}
	},
	reflect.TypeOf(CreateTable{}): func(g *dag.Graph) Command {
		return CreateTable{
			Schema: dag.Nodes[*Schema](g).Any(),
			Name:   RandomString(),
		}
	},
	reflect.TypeOf(CreateIndex{}): func(g *dag.Graph) Command {
		table := dag.Nodes[*Table](g, func(t *Table) bool {
			return len(t.Columns()) > 1
		}).Any()
		return CreateIndex{
			Table:   table,
			Name:    RandomString(),
			Columns: table.Columns().PickUpTo(3),
			Unique:  FlipCoin(),
		}
	},
	// CreateTableAs
	// CreateType
	// CreateView
	// CreateSequence
	// CreateFunc
	// CreateProc

	// ALTER ... https://github.com/cockroachdb/cockroach/blob/master/pkg/sql/parser/sql.y#L1816-L1830
	reflect.TypeOf(RenameTable{}): func(g *dag.Graph) Command {
		return RenameTable{
			Table: dag.Any[*Table](g),
			Name:  RandomString(),
		}
	},
	reflect.TypeOf(RenameSchema{}): func(g *dag.Graph) Command {
		return RenameSchema{
			Schema: dag.Any[*Schema](g, NotPublic),
			Name:   RandomString(),
		}
	},
	reflect.TypeOf(RenameDatabase{}): func(g *dag.Graph) Command {
		return RenameDatabase{
			Database: dag.Any[*Database](g),
			Name:     RandomString(),
		}
	},

	// ALTER TABLE ... https://github.com/cockroachdb/cockroach/blob/master/pkg/sql/parser/sql.y#L1878-L1888
	reflect.TypeOf(DropColumn{}):               func(g *dag.Graph) Command { return DropColumn{dag.Nodes[*Column](g).Any()} },
	reflect.TypeOf(DropForeignKeyConstraint{}): func(g *dag.Graph) Command { panic("not implemented") },
	reflect.TypeOf(AddColumn{}): func(g *dag.Graph) Command {
		return AddColumn{
			Table:    dag.Nodes[*Table](g).Any(),
			Name:     RandomString(),
			Nullable: false,
		}
	},
	reflect.TypeOf(CreateForeignKeyConstraint{}): func(g *dag.Graph) Command {
		// TODO this is pretty constrainted.

		// Find any column that has a unique index.
		to := dag.Nodes[*Index](g, func(i *Index) bool {
			return i.Unique && len(i.Columns()) == 1
		}).Any().Columns()[0]

		// Find any other column that isn't from the same table (could be
		// literally any other column though).
		from := dag.Nodes[*Column](g, func(c *Column) bool {
			return c.Table().Schema().Database() == to.Table().Schema().Database() && c.Table() != to.Table()
		}).Any()

		return CreateForeignKeyConstraint{
			Name: RandomString(),
			From: from,
			To:   to,
		}
	},
}

func GenerateCommand(g *dag.Graph) Command {
	// TODO this is pretty memory hungry, there's certainly a better way to do
	// this. I'm just a bit lazy.
	// https://en.wikipedia.org/wiki/Alias_method
	var weighted []reflect.Type
	for _, cmd := range AllCommands {
		t := reflect.TypeOf(cmd)

		weight, ok := Weights[t]
		if !ok {
			weight = 1
		}

		for i := 0; i < weight; i++ {
			weighted = append(weighted, t)
		}
	}

	rand.Shuffle(len(weighted), func(i, j int) {
		weighted[i], weighted[j] = weighted[j], weighted[i]
	})

	generate := func(t reflect.Type) (cmd Command, ok bool) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("failed to generate %s: %v", t, err)
			}
		}()
		cmd = Generators[t](g)
		return cmd, cmd != nil
	}

	// Try to generate up to 10 steps before giving up. This should probably be
	// a weighted selection that eliminates options as it goes but this is
	// close enough for now.
	for _, t := range weighted[:10] {
		if cmd, ok := generate(t); ok {
			// A bit of sanity checking as we can't statically verify this too
			// well.
			if reflect.TypeOf(cmd) != t {
				panic(errors.Newf("generator for %v returned %T", t, cmd))
			}
			return cmd
		}
	}

	panic("Failed to generate any valid steps after 10 attempts")
}

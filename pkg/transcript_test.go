package pkg_test

import (
	"testing"

	"github.com/chrisseto/scwl/pkg"
	"github.com/chrisseto/scwl/pkg/dag"
	"github.com/stretchr/testify/require"
)

func TestToString(t *testing.T) {
	g := dag.New(nil)

	defaultdb := g.AddNode("1", &pkg.Database{Name: "defaultdb"}).(*pkg.Database)
	public := g.AddNode("2", &pkg.Schema{Name: "public"}).(*pkg.Schema)
	users := g.AddNode("3", &pkg.Table{Name: "users"}).(*pkg.Table)

	g.AddEdge(defaultdb, public)
	g.AddEdge(public, users)

	testCases := []struct {
		In  pkg.Command
		Out string
	}{
		{
			In:  pkg.RenameDatabase{Database: defaultdb, Name: "postgres"},
			Out: `pkg.RenameDatabase{Database: ByFQN(g, "defaultdb"), Name: "postgres"}`,
		},
		{
			In:  pkg.RenameTable{Table: users, Name: "people"},
			Out: `pkg.RenameTable{Table: ByFQN(g, "defaultdb.public.users"), Name: "people"}`,
		},
	}

	for _, tc := range testCases {
		require.Equal(t, tc.Out, pkg.CommandToString(tc.In))
	}
}

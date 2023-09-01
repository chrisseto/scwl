package pkg

import (
	"context"

	"github.com/chrisseto/scwl/pkg/dag"
	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
)

func clone(in dag.INode) dag.INode {
	// Go is weird. We can't clone a struct that implements an interface
	// without doing some reflection magic. Instead, opted for a repetitive
	// type switch. Could probably code gen it.
	switch n := in.(type) {
	case *Schema:
		o := *n
		return &o
	case *Database:
		o := *n
		return &o
	case *Table:
		o := *n
		return &o
	case *Column:
		o := *n
		return &o
	case *Index:
		o := *n
		return &o
	case *ForeignKeyConstraint:
		o := *n
		return &o
	default:
		panic(errors.Newf("unhandled type %T", in))
	}
}

type Database struct {
	dag.Node
	Name string `db:"name"`
}

func (d *Database) Schemas() []*Schema { return dag.Outgoing[*Schema](d) }

type Schema struct {
	dag.Node
	Name string `db:"name"`
}

func (s *Schema) Database() *Database { return dag.Incoming[*Database](s).One() }
func (s *Schema) Tables() []*Table    { return dag.Outgoing[*Table](s) }

type Table struct {
	dag.Node
	Name string `db:"name"`
	// ForeignKeyConstraints []*ForeignKeyConstraint `db:"-"`
}

func (t *Table) Schema() *Schema              { return dag.Incoming[*Schema](t).One() }
func (t *Table) Columns() dag.Result[*Column] { return dag.Outgoing[*Column](t) }
func (t *Table) Indexes() dag.Result[*Index]  { return dag.Outgoing[*Index](t) }

type ForeignKeyConstraint struct {
	dag.Node
	Name string `db:"name"`
}

// TODO relying on order here feels SUPER sketchy. We may need a way to
// annotate edges...
func (c *ForeignKeyConstraint) To() *Column   { return dag.Outgoing[*Column](c)[0] }
func (c *ForeignKeyConstraint) From() *Column { return dag.Outgoing[*Column](c)[1] }

type Index struct {
	dag.Node
	Unique bool   `db:"unique"`
	Name   string `db:"name"`
}

func (i *Index) Table() *Table      { return dag.Incoming[*Table](i).One() }
func (i *Index) Columns() []*Column { return dag.Outgoing[*Column](i) }

type Column struct {
	dag.Node
	Name string `db:"name"`
}

func (c *Column) Table() *Table { return dag.Incoming[*Table](c).One() }

type Queries struct {
	Databases             string
	Schemas               string
	Tables                string
	Columns               string
	Indexes               string
	ColumnsToIndexes      string
	ForeignKeyConstraints string
	// ColumnsToForeignKeyConstraints string
}

func loadState(ctx context.Context, conn *sqlx.DB, queries Queries) (*dag.Graph, error) {
	var databases []struct {
		ID string `db:"id"`
		Database
	}

	var schemas []struct {
		ID         string `db:"id"`
		DatabaseID string `db:"database_id"`
		Schema
	}

	var tables []struct {
		ID       string `db:"id"`
		SchemaID string `db:"schema_id"`
		Table
	}

	var columns []struct {
		ID      string `db:"id"`
		TableID string `db:"table_id"`
		Column
	}

	var columnIndexes []struct {
		ColumnID string `db:"column_id"`
		IndexID  string `db:"index_id"`
	}

	var indexes []struct {
		ID      string `db:"id"`
		TableID string `db:"table_id"`
		Index
	}

	var foreignKeyConstraints []struct {
		FromID string `db:"from_id"`
		ToID   string `db:"to_id"`
		ForeignKeyConstraint
	}

	if err := sqlx.SelectContext(ctx, conn, &databases, queries.Databases); err != nil {
		return nil, errors.WithStack(err)
	}

	if err := sqlx.SelectContext(ctx, conn, &schemas, queries.Schemas); err != nil {
		return nil, errors.WithStack(err)
	}

	if err := sqlx.SelectContext(ctx, conn, &tables, queries.Tables); err != nil {
		return nil, errors.WithStack(err)
	}

	if err := sqlx.SelectContext(ctx, conn, &columns, queries.Columns); err != nil {
		return nil, errors.WithStack(err)
	}

	if err := sqlx.SelectContext(ctx, conn, &indexes, queries.Indexes); err != nil {
		return nil, errors.WithStack(err)
	}

	if err := sqlx.SelectContext(ctx, conn, &columnIndexes, queries.ColumnsToIndexes); err != nil {
		return nil, errors.WithStack(err)
	}

	if err := sqlx.SelectContext(ctx, conn, &foreignKeyConstraints, queries.ForeignKeyConstraints); err != nil {
		return nil, errors.WithStack(err)
	}

	g := dag.New(clone)

	for i := range databases {
		db := &databases[i]
		g.AddNode(db.ID, &db.Database)
	}

	for i := range schemas {
		schema := &schemas[i]
		g.AddNode(schema.ID, &schema.Schema)
		g.AddEdge(dag.ByID[dag.INode](g, schema.DatabaseID), &schema.Schema)
	}

	for i := range tables {
		table := &tables[i]
		g.AddNode(table.ID, &table.Table)
		g.AddEdge(dag.ByID[dag.INode](g, table.SchemaID), &table.Table)
	}

	for i := range columns {
		column := &columns[i]
		g.AddNode(column.ID, &column.Column)
		g.AddEdge(dag.ByID[dag.INode](g, column.TableID), &column.Column)
	}

	for i := range indexes {
		index := &indexes[i]
		g.AddNode(index.ID, &index.Index)
		g.AddEdge(g.ByID(index.TableID), &index.Index)
	}

	for _, colIndex := range columnIndexes {
		g.AddEdge(dag.ByID[dag.INode](g, colIndex.IndexID), dag.ByID[dag.INode](g, colIndex.ColumnID))
	}

	// TODO
	for i := range foreignKeyConstraints {
		fk := &foreignKeyConstraints[i]
		id := RandomString()
		g.AddNode(id, &fk.ForeignKeyConstraint)

		g.AddEdge(&fk.ForeignKeyConstraint, g.ByID(fk.ToID))
		g.AddEdge(&fk.ForeignKeyConstraint, g.ByID(fk.FromID))
	}

	return g, nil
}

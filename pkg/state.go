package pkg

import (
	"context"

	"github.com/chrisseto/scwl/pkg/dag"
	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
)

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

func (t *Table) Schema() *Schema    { return dag.Incoming[*Schema](t).One() }
func (t *Table) Columns() []*Column { return dag.Outgoing[*Column](t) }
func (t *Table) Indexes() []*Index  { return dag.Outgoing[*Index](t) }

type ForeignKeyConstraint struct {
	Name string `db:"name"`

	From *Column `db:"-" json:"-"`
	To   *Column `db:"-" json:"-"`
}

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

	g := dag.New()

	for i := range databases {
		db := &databases[i]
		g.AddNode(db.ID, &db.Database)
	}

	for i := range schemas {
		schema := &schemas[i]
		g.AddNode(schema.ID, &schema.Schema)
		g.AddEdge(schema.DatabaseID, schema.ID)
	}

	for i := range tables {
		table := &tables[i]
		g.AddNode(table.ID, &table.Table)
		g.AddEdge(table.SchemaID, table.ID)
	}

	for i := range columns {
		column := &columns[i]
		g.AddNode(column.ID, &column.Column)
		g.AddEdge(column.TableID, column.ID)
	}

	for i := range indexes {
		index := &indexes[i]
		g.AddNode(index.ID, &index.Index)
		g.AddEdge(index.TableID, index.ID)
	}

	for _, colIndex := range columnIndexes {
		g.AddEdge(colIndex.IndexID, colIndex.ColumnID)
	}

	// TODO
	// for i := range foreignKeyConstraints {
	// 	fk := &foreignKeyConstraints[i]
	//
	// 	fk.To = columnsByID[fk.ToID]
	// 	fk.From = columnsByID[fk.FromID]
	// 	fk.From.Table.ForeignKeyConstraints = append(fk.From.Table.ForeignKeyConstraints, &fk.ForeignKeyConstraint)
	// }

	return g, nil
}

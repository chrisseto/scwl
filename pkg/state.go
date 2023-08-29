package pkg

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
)

func asStateNodes[V StateNode](in []V) []StateNode {
	out := make([]StateNode, len(in))
	for i := range in {
		out[i] = in[i]
	}
	return out
}

type Root struct {
	Databases []*Database
}

type Database struct {
	Name    string    `db:"name"`
	Schemas []*Schema `db:"-"`
}

type Schema struct {
	Name   string   `db:"name"`
	Tables []*Table `db:"-"`
}

type Table struct {
	Name    string    `db:"name"`
	Columns []*Column `db:"-"`
}

type Index struct {
	// Name    string    `db:"name"`
	// Table   *Table    `db:"-"`
	// Columns []*Column `db:"-"`
}

type Column struct {
	Name string `db:"name"`
}

func (n *Root) Children() []StateNode     { return asStateNodes(n.Databases) }
func (n *Database) Children() []StateNode { return asStateNodes(n.Schemas) }
func (n *Schema) Children() []StateNode   { return asStateNodes(n.Tables) }
func (n *Table) Children() []StateNode    { return asStateNodes(n.Columns) }
func (n *Column) Children() []StateNode   { return nil }

type Queries struct {
	Databases string
	Schemas   string
	Tables    string
	Columns   string
	Indexes   string
	// ColumnsToIndexes string
}

func loadState(ctx context.Context, conn *sqlx.DB, queries Queries) (*Root, error) {
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

	// var indexes []struct {
	// 	ID      string `db:"id"`
	// 	TableID string `db:"table_id"`
	// 	Index
	// }

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

	// if err := sqlx.SelectContext(ctx, conn, &indexes, queries.Indexes); err != nil {
	// 	return nil, errors.WithStack(err)
	// }

	databasesByID := make(map[string]*Database, len(databases))
	for i := range databases {
		db := &databases[i]
		databasesByID[db.ID] = &db.Database
	}

	schemasByID := make(map[string]*Schema, len(schemas))
	for i := range schemas {
		schema := &schemas[i]
		schemasByID[schema.ID] = &schema.Schema

		databasesByID[schema.DatabaseID].Schemas = append(databasesByID[schema.DatabaseID].Schemas, &schema.Schema)
	}

	tablesByID := make(map[string]*Table, len(tables))
	for i := range tables {
		table := &tables[i]
		tablesByID[table.ID] = &table.Table

		schemasByID[table.SchemaID].Tables = append(schemasByID[table.SchemaID].Tables, &table.Table)
	}

	columnsByID := make(map[string]*Column, len(columns))
	for i := range columns {
		column := &columns[i]
		columnsByID[column.ID] = &column.Column

		tablesByID[column.TableID].Columns = append(tablesByID[column.TableID].Columns, &column.Column)
	}

	rawDBs := make([]*Database, len(databases))
	for i := range databases {
		rawDBs[i] = &databases[i].Database
	}

	return &Root{Databases: rawDBs}, nil
}

package pkg

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
)

func flatten[T any](in ...[]T) []T {
	var out []T
	for i := range in {
		out = append(out, in[i]...)
	}
	return out
}

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
	Database *Database `db:"-" json:"-"`
	Name     string    `db:"name"`
	Tables   []*Table  `db:"-"`
}

type Table struct {
	Schema                *Schema                 `db:"-" json:"-"`
	Name                  string                  `db:"name"`
	Columns               []*Column               `db:"-"`
	Indexes               []*Index                `db:"-"`
	ForeignKeyConstraints []*ForeignKeyConstraint `db:"-"`
}

type ForeignKeyConstraint struct {
	Name string `db:"name"`

	From *Column `db:"-" json:"-"`
	To   *Column `db:"-" json:"-"`
}

type Index struct {
	Unique bool   `db:"unique"`
	Name   string `db:"name"`

	Table   *Table    `db:"-" json:"-"`
	Columns []*Column `db:"-"`
}

type Column struct {
	Name string `db:"name"`

	Table *Table `db:"-" json:"-"`
}

func (n *Column) Children() []StateNode               { return nil }
func (n *ForeignKeyConstraint) Children() []StateNode { return nil }
func (n *Index) Children() []StateNode                { return nil }
func (n *Database) Children() []StateNode             { return asStateNodes(n.Schemas) }
func (n *Root) Children() []StateNode                 { return asStateNodes(n.Databases) }
func (n *Schema) Children() []StateNode               { return asStateNodes(n.Tables) }
func (n *Table) Children() []StateNode {
	return flatten(
		asStateNodes(n.Columns),
		asStateNodes(n.Indexes),
		asStateNodes(n.ForeignKeyConstraints),
	)
}

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

	databasesByID := make(map[string]*Database, len(databases))
	for i := range databases {
		db := &databases[i]
		databasesByID[db.ID] = &db.Database
	}

	schemasByID := make(map[string]*Schema, len(schemas))
	for i := range schemas {
		schema := &schemas[i]
		schemasByID[schema.ID] = &schema.Schema

		schema.Database = databasesByID[schema.DatabaseID]
		databasesByID[schema.DatabaseID].Schemas = append(databasesByID[schema.DatabaseID].Schemas, &schema.Schema)
	}

	tablesByID := make(map[string]*Table, len(tables))
	for i := range tables {
		table := &tables[i]
		tablesByID[table.ID] = &table.Table

		table.Schema = schemasByID[table.SchemaID]
		schemasByID[table.SchemaID].Tables = append(schemasByID[table.SchemaID].Tables, &table.Table)
	}

	columnsByID := make(map[string]*Column, len(columns))
	for i := range columns {
		column := &columns[i]
		columnsByID[column.ID] = &column.Column

		column.Table = tablesByID[column.TableID]
		tablesByID[column.TableID].Columns = append(tablesByID[column.TableID].Columns, &column.Column)
	}

	indexesByID := make(map[string]*Index, len(indexes))
	for i := range indexes {
		index := &indexes[i]
		indexesByID[index.ID] = &index.Index

		index.Table = tablesByID[index.TableID]
		tablesByID[index.TableID].Indexes = append(tablesByID[index.TableID].Indexes, &index.Index)
	}

	for _, colIndex := range columnIndexes {
		// TODO, fix me?
		if _, ok := indexesByID[colIndex.IndexID]; !ok {
			continue
		}
		indexesByID[colIndex.IndexID].Columns = append(indexesByID[colIndex.IndexID].Columns, columnsByID[colIndex.ColumnID])
	}

	for i := range foreignKeyConstraints {
		fk := &foreignKeyConstraints[i]

		fk.To = columnsByID[fk.ToID]
		fk.From = columnsByID[fk.FromID]
		fk.From.Table.ForeignKeyConstraints = append(fk.From.Table.ForeignKeyConstraints, &fk.ForeignKeyConstraint)
	}

	rawDBs := make([]*Database, len(databases))
	for i := range databases {
		rawDBs[i] = &databases[i].Database
	}

	// TODO switch to using a DAG and order in comparisons or serializations.
	return &Root{Databases: rawDBs}, nil
}

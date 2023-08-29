package pkg

import (
	"context"
	"log"

	"github.com/jmoiron/sqlx"
)

type sut struct {
	conn *sqlx.DB
	log  *log.Logger
}

func NewSUT(conn *sqlx.DB, log *log.Logger) *sut {
	return &sut{conn: conn, log: log}
}

func (o *sut) Execute(ctx context.Context, cmd Command) error {
	stmt := AsDDL(cmd)
	o.log.Printf("Running: %q", stmt)
	_, err := o.conn.ExecContext(ctx, stmt)
	return err
}

func (o *sut) State(ctx context.Context) (StateNode, error) {
	const databasesQuery = `SELECT id, name FROM crdb_internal.databases WHERE name NOT IN ('system') ORDER BY name DESC`

	const schemasQuery = `SELECT id, "parentID" as database_id, name FROM system.namespace WHERE "parentSchemaID" = 0 AND "parentID" > 1 ORDER BY name DESC`

	const tablesQuery = `SELECT
		parent_schema_id as schema_id,
		table_id as id,
		name
	FROM crdb_internal.tables
	WHERE drop_time IS NULL AND parent_schema_id IN (
		SELECT id FROM system.namespace WHERE "parentSchemaID" = 0 AND "parentID" > 1
	)
	ORDER BY name DESC
	`

	const columnQuery = `SELECT 
		descriptor_id::string || column_id::string as id,
		descriptor_id as table_id,
		column_name as name
	FROM "".crdb_internal.table_columns
	WHERE NOT hidden AND descriptor_id IN (
		SELECT id FROM system.namespace WHERE "parentSchemaID" > 99
	)
	ORDER BY column_name DESC
	`

	return loadState(ctx, o.conn, Queries{
		Databases: databasesQuery,
		Schemas:   schemasQuery,
		Tables:    tablesQuery,
		Columns:   columnQuery,
	})
}

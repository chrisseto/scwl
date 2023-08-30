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
		descriptor_id as table_id,
		descriptor_id::string || column_id::string as id,
		column_name as name
	FROM "".crdb_internal.table_columns
	WHERE NOT hidden AND descriptor_id IN (
		SELECT id FROM system.namespace WHERE "parentSchemaID" > 99
	)
	ORDER BY column_name DESC
	`

	const indexQuery = `SELECT
		descriptor_id as table_id,
		descriptor_id::string || index_id::string as id,
		is_unique as "unique",
		index_name as name
	FROM "".crdb_internal.table_indexes
	WHERE index_type != 'primary' AND created_at IS NOT NULL
	ORDER BY index_name DESC
	`

	const columnIndexQuery = `SELECT
		ic.descriptor_id::string || ic.index_id::string as index_id,
		ic.descriptor_id::string || ic.column_id::string as column_id
	FROM "".crdb_internal.index_columns ic
	JOIN "".crdb_internal.table_indexes ti ON (ic.descriptor_id = ti.descriptor_id AND ic.index_id = ti.index_id)
	WHERE ti.index_type != 'primary' AND ti.created_at IS NOT NULL AND ic.column_type = 'key'
	ORDER BY ti.index_name DESC
	`
	// TODO support multi-column FKs
	// TODO convert most queries to protobuf queries??
	const fkQuery = `SELECT
		(fk->>'referencedTableId') || (fk#>>'{referencedColumnIds,0}')  AS to_id,
		(fk->>'originTableId') || (fk#>>'{originColumnIds,0}')  AS from_id,
		fk->>'name' as name
	FROM (
		SELECT jsonb_array_elements(descriptor->'table'->'outboundFks') as fk FROM (
			SELECT
				id,
				crdb_internal.pb_to_json('cockroach.sql.sqlbase.Descriptor', descriptor) as descriptor
			FROM system.descriptor
		)
	) ORDER BY name DESC
	`

	return loadState(ctx, o.conn, Queries{
		Databases:             databasesQuery,
		Schemas:               schemasQuery,
		Tables:                tablesQuery,
		Columns:               columnQuery,
		Indexes:               indexQuery,
		ColumnsToIndexes:      columnIndexQuery,
		ForeignKeyConstraints: fkQuery,
	})
}

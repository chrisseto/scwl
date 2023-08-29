package pkg

import (
	"context"
	"log"

	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
)

const oracleSchema = `
CREATE TABLE databases (
	id TEXT PRIMARY KEY AS (name) STORED,
	name TEXT UNIQUE NOT NULL
);

CREATE TABLE schemas (
	id TEXT PRIMARY KEY AS (database_id || '.' || name) STORED,
	database_id TEXT NOT NULL references databases(id) ON DELETE CASCADE ON UPDATE CASCADE,
	name TEXT NOT NULL
);

CREATE TABLE tables (
	id TEXT PRIMARY KEY AS (schema_id || '.' || name) STORED,
	schema_id TEXT NOT NULL REFERENCES schemas(id) ON DELETE CASCADE ON UPDATE CASCADE,
	name TEXT NOT NULL
);

CREATE TABLE columns (
	id TEXT PRIMARY KEY AS (table_id || '.' || name) STORED,
	table_id TEXT NOT NULL REFERENCES tables(id) ON DELETE CASCADE ON UPDATE CASCADE,
	name TEXT NOT NULL,
	nullable BOOL NOT NULL
);

CREATE TABLE indexes (
	id TEXT PRIMARY KEY AS (table_id || '.' || name) STORED,
	table_id TEXT NOT NULL REFERENCES tables(id) ON DELETE CASCADE ON UPDATE CASCADE,
	name TEXT NOT NULL
);

CREATE TABLE index_columns (
	index_id TEXT NOT NULL REFERENCES indexes(id) ON DELETE CASCADE ON UPDATE CASCADE,
	column_id TEXT NOT NULL REFERENCES columns(id) ON DELETE CASCADE ON UPDATE CASCADE,
	PRIMARY KEY (index_id, column_id)
);
`

type oracle struct {
	conn *sqlx.DB
	log  *log.Logger
}

func NewOracle(db *sqlx.DB, log *log.Logger) (*oracle, error) {
	ctx := context.Background()

	if _, err := db.ExecContext(ctx, oracleSchema); err != nil {
		return nil, errors.WithStack(err)
	}

	o := &oracle{conn: db, log: log}

	for _, cmd := range []Command{
		CreateDatabase{Name: "defaultdb"},
		CreateDatabase{Name: "postgres"},
	} {
		if err := o.Execute(ctx, cmd); err != nil {
			return nil, errors.WithStack(err)
		}
	}

	return o, nil
}

func (o *oracle) Execute(ctx context.Context, cmd Command) error {
	stmt := AsDML(cmd)
	o.log.Printf("Running: %q", stmt)
	_, err := o.conn.ExecContext(ctx, stmt)
	return err
}

func (o *oracle) State(ctx context.Context) (StateNode, error) {
	return loadState(ctx, o.conn, Queries{
		Databases: `SELECT id, name FROM databases ORDER BY name DESC`,
		Schemas:   `SELECT database_id, id, name FROM schemas ORDER BY name DESC`,
		Tables:    `SELECT schema_id, id, name FROM tables ORDER BY name DESC`,
		Columns:   `SELECT table_id, id, name FROM columns ORDER BY name DESC`,
		// Indexes:
	})
}

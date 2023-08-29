package pkg

import (
	"context"
	"log"

	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
)

const oracleSchema = `
CREATE TABLE databases (
	name TEXT PRIMARY KEY
);

CREATE TABLE schemas (
	database TEXT references databases(name) ON DELETE CASCADE ON UPDATE CASCADE NOT NULL,
	name text not null,
	PRIMARY KEY (database, name)
);

CREATE TABLE tables (
	database TEXT references databases(name) ON DELETE CASCADE ON UPDATE CASCADE NOT NULL,
	schema TEXT NOT NULL,
	name text not null,
	PRIMARY KEY (database, schema, name),
	CONSTRAINT fk_schema FOREIGN KEY (database, schema) REFERENCES schemas ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE columns (
	database TEXT references databases(name) ON DELETE CASCADE ON UPDATE CASCADE NOT NULL,
	schema TEXT NOT NULL,
	"table" TEXT NOT NULL,
	name TEXT NOT NULL,
	nullable BOOL NOT NULL,
	PRIMARY KEY (database, schema, "table", name),
	CONSTRAINT fk_table FOREIGN KEY (database, schema, "table") REFERENCES tables ON DELETE CASCADE ON UPDATE CASCADE
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
		Databases: `SELECT name as id, name FROM databases ORDER BY name DESC`,
		Schemas:   `SELECT database as database_id, database || name as id, name FROM schemas ORDER BY name DESC`,
		Tables:    `SELECT database || schema as schema_id, database || schema || name as id, name FROM tables ORDER BY name DESC`,
		Columns:   `SELECT database || schema || "table" as table_id, database || schema || "table" || name as id, name FROM columns ORDER BY name DESC`,
		// Indexes:
	})
}

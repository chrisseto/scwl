package pkg

import (
	"bytes"
	"fmt"
	"reflect"
	"text/template"

	"github.com/chrisseto/scwl/pkg/dag"
	"github.com/cockroachdb/errors"
)

func init() {
	// Linting, assert that there's a translation for every Command.
	for _, cmd := range AllCommands {
		if _, ok := translations[reflect.TypeOf(cmd)]; !ok {
			panic(errors.Newf("missing DDL/DML translation for %T", cmd))
		}
	}
}

func AsDDL(cmd Command) string {
	tpl, ok := translations[reflect.TypeOf(cmd)]
	if !ok {
		panic(fmt.Sprintf("unimplemented: %v", cmd))
	}
	return Tpl(tpl.DDL, cmd)
}

func AsDML(cmd Command) string {
	tpl, ok := translations[reflect.TypeOf(cmd)]
	if !ok {
		panic(fmt.Sprintf("unimplemented: %T", cmd))
	}
	return Tpl(tpl.DML, cmd)
}

type Translation struct {
	DML string
	DDL string
}

var translations = map[reflect.Type]Translation{
	reflect.TypeOf(CreateDatabase{}): {
		DDL: `CREATE DATABASE "{{ .Name }}"`,
		DML: `
			INSERT INTO databases(name) VALUES ('{{ .Name }}');
			INSERT INTO schemas(database_id, name) VALUES ('{{ .Name }}', 'public');
		`,
	},
	reflect.TypeOf(DropDatabase{}): {
		DDL: `DROP DATABASE "{{ .Database.Name }}"`,
		DML: `DELETE FROM databases WHERE name = '{{ .Name }}'`,
	},
	reflect.TypeOf(CreateSchema{}): {
		DDL: `CREATE SCHEMA "{{.Database.Name }}"."{{.Name}}"`,
		DML: `INSERT INTO schemas(database_id, name) VALUES ('{{.Database.Name}}', '{{.Name}}')`,
	},
	reflect.TypeOf(DropSchema{}): {
		DDL: `DROP SCHEMA {{.Database | fqnq}} CASCADE`,
		DML: `DELETE FROM schemas WHERE id = '{{ .Database | fqn }}'`,
	},
	reflect.TypeOf(CreateTable{}): {
		DDL: `CREATE TABLE {{ .Schema | fqnq }}."{{.Name}}" ()`,
		DML: `INSERT INTO tables(schema_id, name) VALUES ('{{ .Schema | fqn }}', '{{.Name}}')`,
	},
	reflect.TypeOf(DropTable{}): {
		DDL: `DROP TABLE {{ .Table | fqnq }}`,
		DML: `DELETE FROM tables WHERE id = '{{ .Table | fqn}}'`,
	},
	reflect.TypeOf(AddColumn{}): {
		DDL: `ALTER TABLE {{ .Table | fqnq }} ADD COLUMN "{{ .Name }}" TEXT NOT NULL`,
		DML: `INSERT INTO columns(table_id, name, nullable) VALUES ('{{ .Table | fqn }}', '{{ .Name }}', false)`,
	},
	reflect.TypeOf(DropColumn{}): {
		DDL: `ALTER TABLE {{ .Table | fqnq }} DROP COLUMN "{{ .Column.Name }}"`,
		DML: `DELETE FROM columns WHERE id = '{{ .Column | fqn }}'`,
	},
	reflect.TypeOf(CreateIndex{}): {
		DDL: `CREATE {{if .Unique}}UNIQUE{{ end }} INDEX "{{ .Name }}"  ON {{ .Table | fqnq }} (
			{{range $i, $column := .Columns}}
				{{if gt $i 0 }},{{ end }}
				"{{ $column.Name }}"
			{{end}}
		)`,
		DML: `
			INSERT INTO indexes(table_id, name, "unique") VALUES ('{{ .Table | fqn }}', '{{ .Name }}', {{ .Unique }});
			{{range $i, $column := .Columns}}
				INSERT INTO index_columns(index_id, column_id) VALUES ('{{ .Table | fqn }}.idxs.{{ $.Name }}', '{{ $column | fqn }}');
			{{end}}
		`,
	},
	reflect.TypeOf(CreateForeignKeyConstraint{}): {
		DDL: `ALTER TABLE {{ .From.Table | fqnq }} ADD CONSTRAINT "{{ .Name }}" FOREIGN KEY ("{{ .From.Name }}") REFERENCES {{ .To.Table | fqnq }} ({{ .To.Name }})`,
		DML: ` INSERT INTO fk_constraints(to_id, from_id, name) VALUES (
			'{{ .To | fqn }}',
			'{{ .From | fqn }}',
			'{{ .Name }}'
		)`,
	},
	reflect.TypeOf(DropForeignKeyConstraint{}): {
		DDL: `ALTER TABLE {{ .ForeignKeyConstraint.From.Table | fqnq }} DROP CONSTRAINT "{{ .ForeignKeyConstraint.Name }}"`,
		// TODO This is probably buggy
		DML: `DELETE FROM fk_constraints WHERE name = '{{ .ForeignKeyConstraint.Name }}'`,
	},
	reflect.TypeOf(DropIndex{}): {
		DDL: `DROP INDEX {{ .Index.Table | fqnq }}@"{{ .Index.Name }}" CASCADE`,
		DML: `DELETE FROM indexes WHERE id = '{{ .Index | fqn }}'`,
	},
	reflect.TypeOf(RenameDatabase{}): {
		DDL: `ALTER DATABASE {{ .Database | fqnq }} RENAME TO "{{ .Name }}"`,
		DML: `UPDATE databases SET name = '{{ .Name }}' WHERE id = '{{ .Database | fqn }}'`,
	},
	reflect.TypeOf(RenameSchema{}): {
		DDL: `ALTER SCHEMA {{ .Schema | fqnq }} RENAME TO "{{ .Name }}"`,
		DML: `UPDATE schemas SET name = '{{ .Name }}' WHERE id = '{{ .Schema | fqn }}'`,
	},
	reflect.TypeOf(RenameTable{}): {
		DDL: `ALTER TABLE {{ .Table | fqnq }} RENAME TO "{{ .Name }}"`,
		DML: `UPDATE tables SET name = '{{ .Name }}' WHERE id = '{{ .Table | fqn }}'`,
	},
}

// TODO just use text/template
func Tpl(body string, vars any) string {
	tmpl, err := template.New("").Funcs(template.FuncMap{
		"fqn": FullyQualifiedName,
		"fqnq": func(sn dag.INode) string {
			switch n := sn.(type) {
			case *Database:
				return fmt.Sprintf("%q", n.Name)
			case *Schema:
				return fmt.Sprintf("%q.%q", n.Database().Name, n.Name)
			case *Table:
				return fmt.Sprintf("%q.%q.%q", n.Schema().Database().Name, n.Schema().Name, n.Name)
			default:
				panic(fmt.Sprintf("unhandled type: %T", sn))
			}
		},
	}).Parse(body)
	if err != nil {
		panic(err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, vars); err != nil {
		panic(err)
	}

	return buf.String()
}

package pkg

import (
	"fmt"
	"reflect"
)

type Translation struct {
	DML string
	DDL string
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
}

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
		DDL: `CREATE DATABASE "{{.Name}}"`,
		DML: `
			INSERT INTO databases(name) VALUES ('{{.Name}}');
			INSERT INTO schemas(database, name) VALUES ('{{.Name}}', 'public');
		`,
	},
	reflect.TypeOf(DropDatabase{}): {
		DDL: `DROP DATABASE "{{.Name}}"`,
		DML: `DELETE FROM databases WHERE name = '{{.Name}}'`,
	},
	reflect.TypeOf(CreateSchema{}): {
		DDL: `CREATE SCHEMA "{{.Database}}"."{{.Name}}"`,
		DML: `INSERT INTO schemas(database, name) VALUES ('{{.Database}}', '{{.Name}}')`,
	},
	reflect.TypeOf(DropSchema{}): {
		DDL: `DROP SCHEMA "{{.Database}}"."{{.Name}}" CASCADE`,
		DML: `DELETE FROM schemas WHERE name = '{{.Name}}' and database = '{{.Database}}'`,
	},
	reflect.TypeOf(CreateTable{}): {
		DDL: `CREATE TABLE "{{.Database}}"."{{.Schema}}"."{{.Name}}" ()`,
		DML: `INSERT INTO tables(database, schema, name) VALUES ('{{.Database}}', '{{.Schema}}', '{{.Name}}')`,
	},
	reflect.TypeOf(DropTable{}): {
		DDL: `DROP TABLE "{{.Database}}"."{{.Schema}}"."{{.Name}}"`,
		DML: `DELETE FROM tables WHERE  database = '{{.Database}}' AND schema = '{{.Schema}}' and name = '{{.Name}}'`,
	},
	reflect.TypeOf(AddColumn{}): {
		DDL: `ALTER TABLE "{{ .Table.Schema.Database.Name }}"."{{ .Table.Schema.Name }}"."{{ .Table.Name }}" ADD COLUMN "{{ .Name }}" TEXT NOT NULL`,
		DML: `INSERT INTO columns(database, schema, "table", name, nullable) VALUES ('{{.Table.Schema.Database.Name}}', '{{.Table.Schema.Name}}', '{{.Table.Name}}', '{{.Name}}', false)`,
	},
	reflect.TypeOf(DropColumn{}): {
		DDL: `ALTER TABLE "{{ .Column.Table.Schema.Database.Name }}"."{{ .Column.Table.Schema.Name }}"."{{ .Column.Table.Name }}" DROP COLUMN "{{ .Column.Name }}"`,
		DML: `DELETE FROM columns WHERE database || schema || "table" || name = '{{ .Column.Table.Schema.Database.Name }}{{ .Column.Table.Schema.Name }}{{ .Column.Table.Name }}{{ .Column.Name }}'`,
	},
}

package pkg

import (
	"bytes"
	"fmt"
	"text/template"
)

// TODO just use text/template
func Tpl(body string, vars any) string {
	tmpl, err := template.New("").Funcs(template.FuncMap{
		"fqn": func(sn StateNode) string {
			switch n := sn.(type) {
			case *Database:
				return n.Name
			case *Schema:
				return fmt.Sprintf("%s.%s", n.Database.Name, n.Name)
			case *Table:
				return fmt.Sprintf("%s.%s.%s", n.Schema.Database.Name, n.Schema.Name, n.Name)
			case *Column:
				return fmt.Sprintf("%s.%s.%s.%s", n.Table.Schema.Database.Name, n.Table.Schema.Name, n.Table.Name, n.Name)
			default:
				panic(fmt.Sprintf("unhandled type: %T", sn))
			}
		},
		"fqnq": func(sn StateNode) string {
			switch n := sn.(type) {
			case *Database:
				return fmt.Sprintf("%q", n.Name)
			case *Schema:
				return fmt.Sprintf("%q.%q", n.Database.Name, n.Name)
			case *Table:
				return fmt.Sprintf("%q.%q.%q", n.Schema.Database.Name, n.Schema.Name, n.Name)
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

type CreateDatabase struct {
	Name string
}

type RenameDatabase struct {
	Database *Database
	Name     string
}

type DropDatabase struct {
	Database *Database
}

type CreateSchema struct {
	Database *Database
	Name     string
}

type RenameSchema struct {
	Schema *Schema
	Name   string
}

type DropSchema struct {
	Schema *Schema
}

type CreateTable struct {
	Schema *Schema
	Name   string
}

type RenameTable struct {
	Table *Table
	Name  string
}

type DropTable struct {
	Table *Table
}

type AddColumn struct {
	Table    *Table
	Name     string
	Nullable bool
}

type DropColumn struct {
	Column *Column
}

package pkg

import (
	"bytes"
	"text/template"
)

// TODO just use text/template
func Tpl(body string, vars any) string {
	tmpl, err := template.New("").Parse(body)
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
	Database string
	Name     string
}

type DropDatabase struct {
	Name string
}

type CreateSchema struct {
	Database string
	Name     string
}

type RenameSchema struct {
	Database string
	Schema   string
	Name     string
}

type DropSchema struct {
	Database string
	Name     string
}

type CreateTable struct {
	Database string
	Schema   string
	Name     string
}

type RenameTable struct {
	Database string
	Schema   string
	Table    string
	Name     string
}

type DropTable struct {
	Database string
	Schema   string
	Name     string
}

type AddColumn struct {
	Table    *Table
	Name     string
	Nullable bool
}

type DropColumn struct {
	Database string
	Schema   string
	Table    string
	Name     string
	Nullable bool
}

package pkg

import (
	"fmt"
	"reflect"
	"strings"
)

// TODO just use text/template
func Tpl(template string, vars any) string {
	val := reflect.ValueOf(vars)
	typ := val.Type()
	bindings := make([]string, val.NumField()*2)
	for i := 0; i < val.NumField(); i++ {
		bindings[i*2] = fmt.Sprintf("{.%s}", typ.Field(i).Name)
		bindings[(i*2)+1] = val.Field(i).String()
	}
	return strings.NewReplacer(bindings...).Replace(template)
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
	Database string
	Schema   string
	Table    string
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

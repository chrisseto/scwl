package pkg

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

type CreateIndex struct {
	Table   *Table
	Columns []*Column
	Name    string
	Unique  bool
}

type DropIndex struct {
	Index *Index
}

type CreateForeignKeyConstraint struct {
	From *Column
	To   *Column
	Name string
}

type DropForeignKeyConstraint struct {
	ForeignKeyConstraint *ForeignKeyConstraint
}

package sqlparser

import (
	"strings"
)

type TableConstraint struct {
	ColumnPrimaryKey []string
	ColumnUniqueKey  []string // col names, not index name
	ColumnForeignKey []string
}

type Table struct {
	Name        string
	Columns     []*Column
	Constraints []*TableConstraint
}

func (t *Table) String() string {
	str := strings.Builder{}
	str.WriteString("table name: ")
	str.WriteString(t.Name)
	str.WriteString("\n")
	for _, col := range t.Columns {
		str.WriteString("col ")
		str.WriteString(col.Name)
		str.WriteString("\t\t")
		str.WriteString(col.DataType.Source)
		str.WriteString("\t\t")
		if col.Constraint.Comment != "" {
			str.WriteString("comment ")
			str.WriteString(col.Constraint.Comment)
		}
		str.WriteString("\n")
	}
	for _, cons := range t.Constraints {
		if cons.ColumnPrimaryKey != nil {
			str.WriteString("primary key: ")
			str.WriteString(strings.Join(cons.ColumnPrimaryKey, ", "))
			str.WriteString("\n")
		}
		if cons.ColumnUniqueKey != nil {
			str.WriteString("primary key: ")
			str.WriteString(strings.Join(cons.ColumnUniqueKey, ", "))
			str.WriteString("\n")
		}
		if cons.ColumnForeignKey != nil {
			str.WriteString("primary key: ")
			str.WriteString(strings.Join(cons.ColumnForeignKey, ", "))
			str.WriteString("\n")
		}
	}
	return str.String()
}

type Column struct {
	Name       string
	DataType   *DataType
	Constraint *ColumnConstraint
}

type CreateTable struct {
	Name        string
	Columns     []*ColumnDeclaration
	Constraints []*TableConstraint
}

// Convert from CreateTable to Table
func (c *CreateTable) Convert() *Table {
	var res Table
	res.Name = onlyTableName(c.Name)
	for _, col := range c.Columns {
		def := col.ColumnDefinition
		var data Column
		data.Name = col.Name
		if def != nil {
			data.DataType = def.DataType
			data.Constraint = def.ColumnConstraint
		}
		res.Columns = append(res.Columns, &data)
	}
	res.Constraints = c.Constraints
	return &res
}

type CreateDefinitions struct {
	ColumnDeclarations []*ColumnDeclaration
	TableConstraints   []*TableConstraint
}

type ColumnDeclaration struct {
	Name             string
	ColumnDefinition *ColumnDefinition
}

type ColumnDefinition struct {
	DataType         *DataType
	ColumnConstraint *ColumnConstraint
}

type DataType struct {
	Name   string
	Number int
	Source string

	HasLength  bool
	Length     int
	IsNational bool
	IsBinary   bool
	IsNChar    bool
	IsVarying  bool

	IsSigned   bool
	IsUnsigned bool
	IsZeroFill bool

	// LengthTwoDimension | Optional
	HasTwoLength bool
	Len1         int
	Len2         int

	// Long
	IsVarchar   bool
	IsChar      bool
	IsCharset   bool
	IsCharacter bool

	CharsetName string

	CollectionOptions []string // for collectionDataType (enum, set)
}

type ColumnConstraint struct {
	NotNull       bool
	DefaultValue  *DefaultValue
	AutoIncrement bool
	Primary       bool
	Key           bool
	Unique        bool
	Comment       string
}

type DefaultValue struct {
	Value string
	Is    bool
}

type key bool
type primary bool

func onlyTableName(name string) string {
	// antlr4 parse table name `db_name`.`tbl_name` as string
	// "db_name`.`tbl_name"
	ss := strings.Split(name, "`.`")
	return ss[len(ss)-1]
}

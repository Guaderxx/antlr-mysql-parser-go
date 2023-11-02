package sqlparser

import (
	"log/slog"
	"strconv"
	"strings"

	"github.com/antlr4-go/antlr/v4"
)

type Visitor struct {
	*BaseMySqlParserVisitor
}

var _ MySqlParserVisitor = (*Visitor)(nil)

func (v *Visitor) VisitUid(ctx *UidContext) interface{} {
	str := ctx.GetText()
	str = WithTrimQuote(str)
	str = WithReplacer(str, "\r", "", "\n", "")
	return str
}

func (v *Visitor) VisitIndexColumnName(ctx *IndexColumnNameContext) interface{} {
	var col string
	if ctx.Uid() != nil {
		uidCtx := ctx.Uid().(*UidContext)
		tmp := v.VisitUid(uidCtx)
		if val, ok := tmp.(string); ok {
			col = val
		}
	} else {
		col = ctx.STRING_LITERAL().GetText()
		col = WithTrimQuote(col)
		col = WithReplacer(col, "\t", "", "\n", "", "\r", "")
	}
	slog.Debug("VisitIndexColumnName", "index column name", col)
	return col
}

func (v *Visitor) VisitIndexColumnNames(ctx *IndexColumnNamesContext) interface{} {
	var cols []string
	for _, col := range ctx.AllIndexColumnName() {
		if indexCtx, ok := col.(*IndexColumnNameContext); ok {
			indexColName := v.VisitIndexColumnName(indexCtx)
			if val, ok := indexColName.(string); ok {
				cols = append(cols, val)
			}
		}
	}
	slog.Debug("VisitIndexColumnNames", "index column names", cols)
	return cols
}

// sqlstatement start

func (v *Visitor) VisitRoot(ctx *RootContext) interface{} {
	if val, ok := ctx.SqlStatements().(*SqlStatementsContext); ok && val != nil {
		return v.VisitSqlStatements(val)
	}
	return nil
}

func (v *Visitor) VisitSqlStatements(ctx *SqlStatementsContext) interface{} {
	var createTables []*CreateTable
	for _, val := range ctx.AllSqlStatement() {
		if sqlStatementCtx, ok := val.(*SqlStatementContext); ok && sqlStatementCtx != nil {
			res := v.VisitSqlStatement(sqlStatementCtx)
			if res != nil {
				if data, ok := res.(*CreateTable); ok {
					createTables = append(createTables, data)
				}
			}
		}
	}
	return createTables
}

func (v *Visitor) VisitSqlStatement(ctx *SqlStatementContext) interface{} {
	if tmp := ctx.DdlStatement(); tmp != nil {
		return v.VisitDdlStatement(tmp.(*DdlStatementContext))
	}
	if tmp := ctx.DmlStatement(); tmp != nil {
		slog.Warn("unsupport VisitDmlStatement")
	}
	if tmp := ctx.TransactionStatement(); tmp != nil {
		slog.Warn("unsupport VisitTransactionStatement")
	}
	if tmp := ctx.ReplicationStatement(); tmp != nil {
		slog.Warn("unsupport VisitReplicationStatement")
	}
	if tmp := ctx.PreparedStatement(); tmp != nil {
		slog.Warn("unsupport VisitPreparedStatement")
	}
	if tmp := ctx.AdministrationStatement(); tmp != nil {
		slog.Warn("unsupport AdministrationStatement")
	}
	if tmp := ctx.UtilityStatement(); tmp != nil {
		slog.Warn("unsupport UtilityStatement")
	}
	return nil
}

func (v *Visitor) VisitDdlStatement(ctx *DdlStatementContext) interface{} {
	if ctx.CreateTable() != nil {
		return v.VisitCreateTable(ctx.CreateTable())
	}
	return nil
}

// sqlstatement end

// --- createTable start

func (v *Visitor) VisitCreateTable(ctx ICreateTableContext) interface{} {

	switch tx := ctx.(type) {
	case *CopyCreateTableContext:
		slog.Warn("unsupported creating a table by copying from another table")
		return nil
	case *QueryCreateTableContext:
		slog.Warn("unsupported creating a table by querying from another table")
		return nil
	case *ColumnCreateTableContext:
		slog.Debug("CreateTable  ColumnCreateTable")
		return v.VisitColumnCreateTable(tx)
	default:
		slog.Warn("unknown CreateTableContext", tx)
		return nil
	}
}

func (v *Visitor) VisitColumnCreateTable(ctx *ColumnCreateTableContext) interface{} {

	var res CreateTable
	tblName := ctx.TableName().GetText()
	tblName = WithTrimQuote(tblName)
	tblName = WithReplacer(tblName, "\t", "", "\r", "", "\n", "")
	res.Name = tblName

	slog.Debug("VisitColumnCreateTable", "tableName", tblName)

	if ctx.CreateDefinitions() != nil {
		if createDefCtx, ok := ctx.CreateDefinitions().(*CreateDefinitionsContext); ok {
			slog.Debug("ColumnCreateTable CreateDefinition Exist")

			definitions := v.VisitCreateDefinitions(createDefCtx)
			if val, ok := definitions.(CreateDefinitions); ok {
				res.Columns = val.ColumnDeclarations
				res.Constraints = val.TableConstraints
			}
			return &res
		}
	}
	slog.Debug("ColumnCreateTable CreateDefinition Not Exist")
	return &res
}

func (v *Visitor) VisitCreateDefinitions(ctx *CreateDefinitionsContext) interface{} {

	var res CreateDefinitions
	res.ColumnDeclarations = make([]*ColumnDeclaration, 0)
	res.TableConstraints = make([]*TableConstraint, 0)

	for _, def := range ctx.AllCreateDefinition() {
		data := v.VisitCreateDefinition(def)
		if data != nil {
			switch r := data.(type) {
			case *ColumnDeclaration:
				tmp := append(res.ColumnDeclarations[:], r)
				res.ColumnDeclarations = tmp
			case *TableConstraint:
				tmp := append(res.TableConstraints[:], r)
				res.TableConstraints = tmp
			}
		}
	}

	// I think every createTable should only have one
	// tableConstraint and many columnDeclaration
	// Actually, more than one TableConstraint and many ColumnDeclaration
	// for _, tmp := range res {
	// 	slog.Debug("CreateDefinition: %v", tmp)
	// }
	return res
}

func (v *Visitor) VisitCreateDefinition(ctx ICreateDefinitionContext) interface{} {

	switch tx := ctx.(type) {
	case *ColumnDeclarationContext:
		slog.Debug("VisitCreateDefinition", "ctx", "ColumnDeclaration")

		var res *ColumnDeclaration
		res = v.VisitColumnDeclaration(tx).(*ColumnDeclaration)
		if res == nil {
			res = new(ColumnDeclaration)
		}

		if definitionCtx, ok := tx.ColumnDefinition().(*ColumnDefinitionContext); ok {
			tmp := v.VisitColumnDefinition(definitionCtx)
			if cd, ok := tmp.(*ColumnDefinition); ok {
				res.ColumnDefinition = cd
			}
		}
		return res
	case *ConstraintDeclarationContext:
		slog.Debug("VisitCreateDefinition", "ctx", "TableConstraint")
		if tmp := tx.TableConstraint(); tmp != nil {
			return v.VisitTableConstraint(tmp)
		}
	case *IndexDeclarationContext:
		slog.Debug("VisitCreateDefinition", "ctx", "IndexDeclaration")
		if tmp := tx.IndexColumnDefinition(); tmp != nil {
			return v.VisitIndexColumnDefinition(tmp)
		}
	default:
		slog.Warn("unsupported ICreateDefinitionContext", "ctx", ctx)
		return nil
	}
	return nil
}

func (v *Visitor) VisitIndexColumnDefinition(ctx IIndexColumnDefinitionContext) interface{} {
	// TODO:
	slog.Warn("unsupport VisitIndexColumnDefinition")
	return nil
}

// --- createTable end

// --- tableConstraint start

// VisitTableConstraint
func (v *Visitor) VisitTableConstraint(ctx ITableConstraintContext) *TableConstraint {

	var res TableConstraint
	var tmp interface{}

	switch ctx.(type) {
	case *PrimaryKeyTableConstraintContext:
		slog.Debug("VisitTableConstraint", "ctx", "*PrimaryKeyTableConstraintContext")
		tmp = v.VisitPrimaryKeyTableConstraint(ctx.(*PrimaryKeyTableConstraintContext))
		if val, ok := tmp.([]string); ok {
			res.ColumnPrimaryKey = val
		}
	case *UniqueKeyTableConstraintContext:
		slog.Debug("VisitTableConstraint", "ctx", "*UniqueKeyTableConstraintContext")
		tmp = v.VisitUniqueKeyTableConstraint(ctx.(*UniqueKeyTableConstraintContext))
		if val, ok := tmp.([]string); ok {
			res.ColumnUniqueKey = val
		}
	case *ForeignKeyTableConstraintContext:
		slog.Debug("VisitTableConstraint", "ctx", "*ForeignKeyTableConstraintContext")
		tmp = v.VisitForeignKeyTableConstraint(ctx.(*ForeignKeyTableConstraintContext))
		if val, ok := tmp.([]string); ok {
			res.ColumnForeignKey = val
		}
	}
	return &res
}

func (v *Visitor) VisitPrimaryKeyTableConstraint(ctx *PrimaryKeyTableConstraintContext) interface{} {

	var res []string
	if ctx.IndexColumnNames() != nil {
		if indexColumnsNamesCtx, ok := ctx.IndexColumnNames().(*IndexColumnNamesContext); ok {
			tmp := v.VisitIndexColumnNames(indexColumnsNamesCtx)
			if val, ok := tmp.([]string); ok {
				res = val
			}
		}
	}
	return res
}

func (v *Visitor) VisitUniqueKeyTableConstraint(ctx *UniqueKeyTableConstraintContext) interface{} {

	var res []string

	for _, uid := range ctx.AllUid() {
		if uu, ok := uid.(*UidContext); ok {
			if us, ok := v.VisitUid(uu).(string); ok {
				// unique_index_name
				slog.Debug("VisitUniqueKeyTableConstraint", "uk_uid", us)
			}
		}
	}
	if ctx.IndexColumnNames() != nil {
		if idxColNamesCtx, ok := ctx.IndexColumnNames().(*IndexColumnNamesContext); ok {
			tmp := v.VisitIndexColumnNames(idxColNamesCtx)
			if val, ok := tmp.([]string); ok {
				res = val
			}
		}
	}

	return res
}

func (v *Visitor) VisitForeignKeyTableConstraint(ctx *ForeignKeyTableConstraintContext) interface{} {

	var res []string
	if ctx.IndexColumnNames() != nil {
		if indexColumnsNamesCtx, ok := ctx.IndexColumnNames().(*IndexColumnNamesContext); ok {
			tmp := v.VisitIndexColumnNames(indexColumnsNamesCtx)
			if val, ok := tmp.([]string); ok {
				res = val
			}
		}
	}
	return res
}

// --- tableConstraint end

// --- columnDefinition start

func (v *Visitor) VisitColumnDefinition(ctx *ColumnDefinitionContext) interface{} {
	var constraint ColumnConstraint
	var definition ColumnDefinition

	definition.DataType = v.VisitDataType(ctx.DataType()).(*DataType)
	for _, cons := range ctx.AllColumnConstraint() {
		switch tx := cons.(type) {
		case *NullColumnConstraintContext:
			slog.Debug("VisitColumnDefinition", "ColumnConstraint", "Null")
			constraint.NotNull = v.VisitNullColumnConstraint(tx).(bool)
		case *DefaultColumnConstraintContext:
			slog.Debug("VisitColumnDefinition", "ColumnConstraint", "Default")
			constraint.DefaultValue = v.VisitDefaultColumnConstraint(tx).(*DefaultValue)
		case *AutoIncrementColumnConstraintContext:
			slog.Debug("VisitColumnDefinition", "ColumnConstraint", "AutoIncrement")
			constraint.AutoIncrement = v.VisitAutoIncrementColumnConstraint(tx).(bool)
		case *PrimaryKeyColumnConstraintContext:
			slog.Debug("VisitColumnDefinition", "ColumnConstraint", "PrimaryKey")
			ret := v.VisitPrimaryKeyColumnConstraint(tx)
			if c, ok := ret.(*primary); ok {
				// if primary, that means one of the primary
				// more confused, maybe later
				constraint.Primary = bool(*c)
			} else {
				c := ret.(*key)
				constraint.Key = bool(*c)
			}
		case *UniqueKeyColumnConstraintContext:
			slog.Debug("VisitColumnDefinition", "ColumnConstraint", "UniqueKey")
			constraint.Unique = v.VisitUniqueKeyColumnConstraint(tx).(bool)
		case *CommentColumnConstraintContext:
			slog.Debug("VisitColumnDefinition", "ColumnConstraint", "Comment")
			constraint.Comment = v.VisitCommentColumnConstraint(tx).(string)
		case *ReferenceColumnConstraintContext:
			slog.Warn("unsupport ReferenceColumnConstraint")
		case *StorageColumnConstraintContext:
			slog.Warn("unsupport StorageColumnConstraint")
		case *VisibilityColumnConstraintContext:
			slog.Warn("unsupport VisibilityColumnConstraint")
		case *InvisibilityColumnConstraintContext:
			slog.Warn("unsupport InvisibilityColumnConstraint")
		case *SerialDefaultColumnConstraintContext:
			slog.Warn("unsupport SerialDefaultColumnConstraint")
		case *GeneratedColumnConstraintContext:
			slog.Warn("unsupport GeneratedColumnConstraint")
		case *FormatColumnConstraintContext:
			slog.Warn("unsupport FormatColumnConstraint")
		case *CollateColumnConstraintContext:
			slog.Warn("unsupport CollateColumnConstraint")
		case *CheckColumnConstraintContext:
			slog.Warn("unsupport CheckColumnConstraint")
		}
	}
	definition.ColumnConstraint = &constraint
	return &definition
}

func (v *Visitor) VisitNullColumnConstraint(ctx *NullColumnConstraintContext) interface{} {
	if res, ok := ctx.NullNotnull().(*NullNotnullContext); ok {
		return v.VisitNullNotnull(res)
	}
	return false
}

func (v *Visitor) VisitNullNotnull(ctx *NullNotnullContext) interface{} {
	return ctx.NOT() != nil
}

func (v *Visitor) VisitDefaultColumnConstraint(ctx *DefaultColumnConstraintContext) interface{} {
	res := DefaultValue{}
	text := ctx.DefaultValue().GetText()
	text = WithTrimQuote(text)
	text = WithReplacer(text, "\r", "", "\t", "", "\n", "")
	if strings.HasPrefix(strings.ToUpper(text), "NULL") {
		// default NULL
		return &res
	}
	res.Value = text
	res.Is = true
	return &res
}

func (v *Visitor) VisitAutoIncrementColumnConstraint(ctx *AutoIncrementColumnConstraintContext) interface{} {
	// auto := ctx.AUTO_INCREMENT()
	// fmt.Printf("auto: %#v\ntext: %#v\n", auto, auto.GetSymbol().GetText())
	// When entering this function, it means that there is the AUTO_INCREMENT keyword, so it must be true
	return true
}

func (v *Visitor) VisitPrimaryKeyColumnConstraint(ctx *PrimaryKeyColumnConstraintContext) interface{} {
	if ctx.PRIMARY() == nil {
		var res key = true
		return &res
	}
	var res primary = true
	return &res
}

func (v *Visitor) VisitUniqueKeyColumnConstraint(ctx *UniqueKeyColumnConstraintContext) interface{} {
	// key := ctx.KEY().GetSymbol().GetText()
	// unique := ctx.UNIQUE().GetSymbol().GetText()
	// fmt.Printf("key: %#v\nunique: %#v\n", key, unique)
	// When entering this function, it means that there is the UNIQUE keyword, so it must be true
	return true
}

func (v *Visitor) VisitCommentColumnConstraint(ctx *CommentColumnConstraintContext) interface{} {
	commentStr := ctx.STRING_LITERAL().GetText()
	commentStr = WithTrimQuote(commentStr)
	commentStr = WithReplacer(commentStr, "\r", "", "\n", "")
	slog.Debug("VisitCommentColumnConstraint", "comment", commentStr)
	return commentStr
}

// ---  columnDeclaration

func (v *Visitor) VisitColumnDeclaration(ctx *ColumnDeclarationContext) interface{} {

	var res ColumnDeclaration

	if val, ok := ctx.FullColumnName().(*FullColumnNameContext); ok {
		if name, ok := v.VisitFullColumnName(val).(string); ok {
			slog.Debug("VisitColumnDeclaration", "FullColumnName", name)

			res.Name = name
			return &res
		}
	}
	return nil
}

func (v *Visitor) VisitFullColumnName(ctx *FullColumnNameContext) interface{} {
	if uctx, ok := ctx.Uid().(*UidContext); ok {
		return v.VisitUid(uctx)
	}
	return nil
}

// --- columnDefinition end

// --- dataType start

// VisitDataType
func (v *Visitor) VisitDataType(ctx IDataTypeContext) interface{} {
	switch tx := ctx.(type) {
	case *SpatialDataTypeContext:
		return v.VisitSpatialDataType(tx)
	case *LongVarbinaryDataTypeContext:
		return v.VisitLongVarbinaryDataType(tx)
	case *CollectionDataTypeContext:
		return v.VisitCollectionDataType(tx)
	case *NationalVaryingStringDataTypeContext:
		return v.VisitNationalVaryingStringDataType(tx)
	case *DimensionDataTypeContext:
		return v.VisitDimensionDataType(tx)
	case *StringDataTypeContext:
		return v.VisitStringDataType(tx)
	case *LongVarcharDataTypeContext:
		return v.VisitLongVarcharDataType(tx)
	case *NationalStringDataTypeContext:
		return v.VisitNationalStringDataType(tx)
	case *SimpleDataTypeContext:
		return v.VisitSimpleDataType(tx)
	// case *ConvertedDataTypeContext:
	default:
		return nil
	}
}

// VisitStringDataType  return data type token number
func (v *Visitor) VisitStringDataType(ctx *StringDataTypeContext) interface{} {
	res := DataType{}
	token := ctx.GetTypeName()
	res.Name = token.GetText()
	res.Name = strings.ToUpper(res.Name)
	res.Number = token.GetTokenType()
	res.Source = res.Name

	if ctx.LengthOneDimension() != nil {
		res.HasLength = true
		length := ctx.LengthOneDimension().GetText()
		res.Source += length
		length = WithTrimBracket(length)
		intLen, err := strconv.Atoi(length)
		if err != nil {
			slog.Warn("parse string length error", err)
		}
		res.Length = intLen
	}

	return &res
}

func (v *Visitor) VisitNationalVaryingStringDataType(ctx *NationalVaryingStringDataTypeContext) interface{} {
	res := DataType{}
	token := ctx.GetTypeName()
	res.Name = token.GetText()
	res.Name = strings.ToUpper(res.Name)
	res.Number = token.GetTokenType()
	res.Source = res.Name

	if ctx.LengthOneDimension() != nil {
		res.HasLength = true
		length := ctx.LengthOneDimension().GetText()
		res.Source += length
		length = WithTrimBracket(length)
		intLen, err := strconv.Atoi(length)
		if err != nil {
			slog.Warn("VisitNationalVaryingStringDataType", "parse string length error", err)
		}
		res.Length = intLen
	}

	if ctx.NATIONAL() != nil {
		res.Source = "NATIONAL " + res.Source
		res.IsNational = true
	}

	if ctx.BINARY() != nil {
		res.Source += " BINARY"
		res.IsBinary = true
	}

	if ctx.VARYING() != nil {
		res.Source += " VARYING"
		res.IsVarying = true
	}

	return &res
}

func (v *Visitor) VisitNationalStringDataType(ctx *NationalStringDataTypeContext) interface{} {
	res := DataType{}
	token := ctx.GetTypeName()
	res.Name = token.GetText()
	res.Name = strings.ToUpper(res.Name)
	res.Number = token.GetTokenType()
	res.Source = res.Name

	if ctx.LengthOneDimension() != nil {
		res.HasLength = true
		length := ctx.LengthOneDimension().GetText()
		res.Source += length
		length = WithTrimBracket(length)
		intLen, err := strconv.Atoi(length)
		if err != nil {
			slog.Warn("VisitNationalStringDataType", "parse string length error", err)
		}
		res.Length = intLen
	}

	if ctx.NATIONAL() != nil {
		res.Source = "NATIONAL " + res.Source
		res.IsNational = true
	}

	if ctx.BINARY() != nil {
		res.Source += " BINARY"
		res.IsBinary = true
	}

	if ctx.NCHAR() != nil {
		res.Source = "NCHAR " + res.Source
		res.IsNChar = true
	}

	return &res
}

func (v *Visitor) VisitDimensionDataType(ctx *DimensionDataTypeContext) interface{} {
	res := DataType{}
	token := ctx.GetTypeName()
	res.Name = token.GetText()
	res.Name = strings.ToUpper(res.Name)
	res.Number = token.GetTokenType()
	res.Source = res.Name

	if lenToken := ctx.LengthOneDimension(); lenToken != nil {
		length := lenToken.GetText()
		res.Source += length
		length = WithTrimBracket(length)
		intLen, err := strconv.Atoi(length)
		if err != nil {
			slog.Warn("VisitDimensionDataType", "parse dimension datatype length error", err)
		}
		res.HasLength = true
		res.Length = intLen
	}

	if twoToken := ctx.LengthTwoDimension(); twoToken != nil {
		length := twoToken.GetText()
		res.Source += length
		length = WithTrimBracket(length)
		lenArr := strings.Split(length, ",")
		len1, err := strconv.Atoi(lenArr[0])
		if err != nil {
			slog.Warn("VisitDimensionDataType", "parse dimension datatype length error", err, "len1", lenArr[0])
		}
		len2, err := strconv.Atoi(lenArr[1])
		if err != nil {
			slog.Warn("VisitDimensionDataType", "parse dimension datatype length error", err, "len2", lenArr[1])
		}
		res.HasTwoLength = true
		res.Len1 = len1
		res.Len2 = len2
	}

	if twoToken := ctx.LengthTwoOptionalDimension(); twoToken != nil {
		length := twoToken.GetText()
		res.Source += length
		length = WithTrimBracket(length)
		lenArr := strings.Split(length, ",")
		len1, err := strconv.Atoi(lenArr[0])
		if err != nil {
			slog.Warn("VisitDimensionDataType", "parse dimension datatype length error", err, "len1", lenArr[0])
		}
		len2, err := strconv.Atoi(lenArr[1])
		if err != nil {
			slog.Warn("VisitDimensionDataType", "parse dimension datatype length error", err, "len2", lenArr[1])
		}
		res.HasTwoLength = true
		res.Len1 = len1
		res.Len2 = len2
	}

	if tokens := ctx.AllSIGNED(); len(tokens) != 0 {
		res.IsSigned = true
		res.Source += " SIGNED"
	}
	if utokens := ctx.AllUNSIGNED(); len(utokens) != 0 {
		res.IsUnsigned = true
		res.Source += " UNSIGNED"
	}
	if fill := ctx.AllZEROFILL(); len(fill) != 0 {
		res.IsZeroFill = true
		res.Source += " ZEROFILL"
	}
	return &res
}

func (v *Visitor) VisitSimpleDataType(ctx *SimpleDataTypeContext) interface{} {
	res := DataType{}
	token := ctx.GetTypeName()
	res.Name = token.GetText()
	res.Name = strings.ToUpper(res.Name)
	res.Number = token.GetTokenType()
	res.Source = res.Name
	return &res
}

// VisitCollectionDataType
func (v *Visitor) VisitCollectionDataType(ctx *CollectionDataTypeContext) interface{} {
	res := DataType{}
	token := ctx.GetTypeName()

	res.Number = token.GetTokenType()
	res.Name = token.GetText()
	res.Name = strings.ToUpper(res.Name)
	res.Source = res.Name

	res.CollectionOptions = []string{}
	if ctx.CollectionOptions() != nil {
		if opsCtx, ok := ctx.CollectionOptions().(*CollectionOptionsContext); ok {
			ops := opsCtx.AllSTRING_LITERAL()
			res.Source += "("
			tmp := []string{}
			for _, ter := range ops {
				terWithQuote := ter.GetSymbol().GetText()
				res.Source += terWithQuote + ","
				terText := WithTrimQuote(terWithQuote)
				tmp = append(tmp, terText)
			}
			res.Source = res.Source[:len(res.Source)-1] + ")"
			res.CollectionOptions = tmp
		}
	}

	if bCtx := ctx.BINARY(); bCtx != nil {
		res.IsBinary = true
		res.Source += " BINARY"
	}
	return &res
}

func (v *Visitor) VisitSpatialDataType(ctx *SpatialDataTypeContext) interface{} {
	res := DataType{}
	token := ctx.GetTypeName()
	res.Name = token.GetText()
	res.Name = strings.ToUpper(res.Name)
	res.Number = token.GetTokenType()
	res.Source = res.Name

	return &res
}

func (v *Visitor) VisitLongVarcharDataType(ctx *LongVarcharDataTypeContext) interface{} {
	res := DataType{}
	token := ctx.GetTypeName()
	res.Name = token.GetText()
	res.Name = strings.ToUpper(res.Name)
	res.Number = token.GetTokenType()
	res.Source = res.Name

	if vCtx := ctx.VARCHAR(); vCtx != nil {
		res.IsVarchar = true
		res.Source += " VARCHAR"
	}

	if bCtx := ctx.BINARY(); bCtx != nil {
		res.IsBinary = true
		res.Source += " BINARY"
	}

	if setCtx := ctx.CharSet(); setCtx != nil {
		if setCtx.CHAR() != nil {
			slog.Debug("VisitLongVarcharDataType", "charset char", setCtx.CHAR().GetText())
			res.IsChar = true
			res.Source += " CHAR SET "
		}
		if cac := setCtx.CHARACTER(); cac != nil {
			slog.Debug("VisitLongVarcharDataType", "charset character", cac.GetText())
			res.IsCharacter = true
			res.Source += " CHARACTER SET "
		}
		if set := setCtx.CHARSET(); set != nil {
			slog.Debug("VisitLongVarcharDataType", "charset set", set.GetText())
			res.IsCharset = true
			res.Source += " CHARSET SET "
		}
	}

	if setNameCtx := ctx.CharsetName(); setNameCtx != nil {
		if base := setNameCtx.CharsetNameBase(); base != nil {
			slog.Debug("VisitLongVarcharDataType", "charsetname base", base.GetText())
		}
		name := setNameCtx.GetText()
		slog.Debug("VisitLongVarcharDataType", "charsetname", name)
		res.CharsetName = WithTrimQuote(name)
		res.Source += name
	}

	return &res
}

func (v *Visitor) VisitLongVarbinaryDataType(ctx *LongVarbinaryDataTypeContext) interface{} {
	res := DataType{}
	var symbol antlr.Token
	if longNode := ctx.LONG(); longNode != nil {
		symbol = longNode.GetSymbol()
	} else if binNode := ctx.VARBINARY(); binNode != nil {
		symbol = binNode.GetSymbol()
	} else {
		return &res
	}

	res.Number = symbol.GetTokenType()
	res.Name = symbol.GetText()
	res.Name = strings.ToUpper(res.Name)
	res.Source = res.Name

	return &res
}

func (v *Visitor) VisitConvertedDataType(ctx *ConvertedDataTypeContext) interface{} {
	res := DataType{}
	token := ctx.GetTypeName()
	res.Name = token.GetText()
	res.Name = strings.ToUpper(res.Name)
	res.Number = token.GetTokenType()
	res.Source = res.Name

	return &res
}

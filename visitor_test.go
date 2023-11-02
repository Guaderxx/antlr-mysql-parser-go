package sqlparser

import (
	"testing"

	"github.com/antlr4-go/antlr/v4"
	. "github.com/smartystreets/goconvey/convey"
)

func prepare(str string) *MySqlParser {
	input := antlr.NewInputStream(str)
	lexer := NewMySqlLexer(input)
	tokens := antlr.NewCommonTokenStream(lexer, antlr.LexerDefaultTokenChannel)
	parser := NewMySqlParser(tokens)
	el := NewErrorListener()
	parser.RemoveErrorListeners()
	parser.AddErrorListener(el)
	return parser
}

func TestVisitTableConstraint(t *testing.T) {
	Convey("TestVisitTableConstraint", t, func() {
		v := new(Visitor)

		Convey("TestVisitTableConstraint - UNIQUE", func() {
			str := "UNIQUE INDEX `data__update_UNIQUE` (`data` ASC, `update_time` DESC)"
			p := prepare(str)

			res := v.VisitTableConstraint(p.TableConstraint())
			So(res.ColumnUniqueKey, ShouldResemble, []string{"data", "update_time"})
		})

		Convey("TestVisitTableConstraint - PRIMARY", func() {
			str := "primary key (`user_id`, `group_id`)"
			p := prepare(str)

			res := v.VisitTableConstraint(p.TableConstraint())
			So(res.ColumnPrimaryKey, ShouldResemble, []string{"user_id", "group_id"})
		})

		Convey("TestVisitTableConstraint - FOREIGN", func() {
			// TODO:
		})
	})
}

func TestOnlyTableName(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"foo", "foo"},
		{"foo`.`bar", "bar"},
	}

	Convey("TestOnlyTableName", t, func() {
		for _, dt := range tests {
			val := onlyTableName(dt.name)
			So(dt.want, ShouldEqual, val)
		}
	})
}

func TestVisitCreateTable(t *testing.T) {
	Convey("TestVisitCreateTable", t, func() {
		v := new(Visitor)
		var str string
		var p *MySqlParser
		var res interface{}

		Convey("ColumnCreateTable", func() {
			str = "CREATE TABLE `user` (\n  " +
				"`id` bigint NOT NULL AUTO_INCREMENT,\n  " +
				"`number` varchar(255) NOT NULL DEFAULT '' COMMENT '学号',\n  " +
				"`name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '用户名称',\n " +
				" `password` varchar(255) NOT NULL DEFAULT '' COMMENT '用户密码',\n " +
				" `gender` char(5) NOT NULL COMMENT '男｜女｜未公开',\n  " +
				"`create_time` timestamp NULL DEFAULT NULL,\n  " +
				"`update_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,\n  " +
				"PRIMARY KEY (`id`),\n  " +
				"UNIQUE KEY `number_unique` (`number`) USING BTREE,\n  " +
				"UNIQUE KEY `number_unique2` (`number`" +
				") USING BTREE\n) ENGINE=InnoDB AUTO_INCREMENT=8 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;"
			p = prepare(str)
			res = v.VisitCreateTable(p.CreateTable())

			So(res, ShouldResemble, &CreateTable{
				Name: "user",
				Columns: []*ColumnDeclaration{
					{
						Name: "id",
						ColumnDefinition: &ColumnDefinition{
							DataType: &DataType{
								Name:   "BIGINT",
								Source: "BIGINT",
								Number: MySqlLexerBIGINT,
							},
							ColumnConstraint: &ColumnConstraint{
								NotNull:       true,
								AutoIncrement: true,
							},
						},
					},
					{
						Name: "number",
						ColumnDefinition: &ColumnDefinition{
							DataType: &DataType{
								Name:      "VARCHAR",
								Source:    "VARCHAR(255)",
								Number:    MySqlLexerVARCHAR,
								HasLength: true,
								Length:    255,
							},
							ColumnConstraint: &ColumnConstraint{
								NotNull: true,
								DefaultValue: &DefaultValue{
									Value: "",
									Is:    true,
								},
								Comment: "学号",
							},
						},
					},
					{
						Name: "name",
						ColumnDefinition: &ColumnDefinition{
							DataType: &DataType{
								Name:      "VARCHAR",
								Source:    "VARCHAR(255)",
								Number:    MySqlLexerVARCHAR,
								HasLength: true,
								Length:    255,
							},
							ColumnConstraint: &ColumnConstraint{
								NotNull: false,
								DefaultValue: &DefaultValue{
									Value: "",
									Is:    false,
								},
								Comment: "用户名称",
							},
						},
					},
					{
						Name: "password",
						ColumnDefinition: &ColumnDefinition{
							DataType: &DataType{
								Name:      "VARCHAR",
								Source:    "VARCHAR(255)",
								Number:    MySqlLexerVARCHAR,
								HasLength: true,
								Length:    255,
							},
							ColumnConstraint: &ColumnConstraint{
								NotNull: true,
								DefaultValue: &DefaultValue{
									Value: "",
									Is:    true,
								},
								Comment: "用户密码",
							},
						},
					},
					{
						Name: "gender",
						ColumnDefinition: &ColumnDefinition{
							DataType: &DataType{
								Name:      "CHAR",
								Source:    "CHAR(5)",
								Number:    MySqlLexerCHAR,
								HasLength: true,
								Length:    5,
							},
							ColumnConstraint: &ColumnConstraint{
								NotNull: true,
								Comment: "男｜女｜未公开",
							},
						},
					},
					{
						Name: "create_time",
						ColumnDefinition: &ColumnDefinition{
							DataType: &DataType{
								Name:   "TIMESTAMP",
								Source: "TIMESTAMP",
								Number: MySqlLexerTIMESTAMP,
							},
							ColumnConstraint: &ColumnConstraint{
								DefaultValue: &DefaultValue{
									Value: "",
									Is:    false,
								},
							},
						},
					},
					{
						Name: "update_time",
						ColumnDefinition: &ColumnDefinition{
							DataType: &DataType{
								Name:   "TIMESTAMP",
								Source: "TIMESTAMP",
								Number: MySqlLexerTIMESTAMP,
							},
							ColumnConstraint: &ColumnConstraint{
								DefaultValue: &DefaultValue{
									Value: "CURRENT_TIMESTAMPONUPDATECURRENT_TIMESTAMP",
									Is:    true,
								},
							},
						},
					},
				},
				Constraints: []*TableConstraint{
					{
						ColumnPrimaryKey: []string{"id"},
					},
					{
						ColumnUniqueKey: []string{"number"},
					},
					{
						ColumnUniqueKey: []string{"number"},
					},
				},
			})
		})

		Convey("CopyCreateTable", func() {
			str = "create table new_t  (like t1);"
			p = prepare(str)
			res = v.VisitCreateTable(p.CreateTable())

			So(res, ShouldBeNil)
		})

		Convey("QueryCreateTable", func() {
			str = `CREATE TABLE test (a INT NOT NULL AUTO_INCREMENT,PRIMARY KEY (a), KEY(b))ENGINE=InnoDB SELECT b,c FROM test2;`
			p = prepare(str)
			res = v.VisitCreateTable(p.CreateTable())

			So(res, ShouldBeNil)
		})
	})
}

func TestVisitColumnDeclaration(t *testing.T) {
	// TODO:
	Convey("TestVisitColumnDeclaration", t, func() {

		Convey("Test ColumnDeclaration FullColumnName", func() {

		})
	})
}

func TestVisitColumnDefinition(t *testing.T) {
	Convey("TestVisitColumnDefinition", t, func() {
		v := new(Visitor)

		Convey("PRIMARY", func() {
			str := `bigint(20) NOT NULL DEFAULT 'test default' PRIMARY KEY COMMENT 'test comment'`
			p := prepare(str)
			res := v.VisitColumnDefinition(p.ColumnDefinition().(*ColumnDefinitionContext)).(*ColumnDefinition)
			So(res, ShouldNotBeNil)
			So(res.DataType, ShouldResemble, &DataType{
				Name:      "BIGINT",
				Source:    "BIGINT(20)",
				Number:    MySqlLexerBIGINT,
				HasLength: true,
				Length:    20,
			})
			So(res.ColumnConstraint, ShouldResemble, &ColumnConstraint{
				NotNull: true,
				DefaultValue: &DefaultValue{
					Value: "test default",
					Is:    true,
				},
				Primary: true,
				Comment: "test comment",
			})
		})

		Convey("NULL", func() {
			str := `bigint(20) NULL KEY`
			p := prepare(str)
			res := v.VisitColumnDefinition(p.ColumnDefinition().(*ColumnDefinitionContext)).(*ColumnDefinition)

			So(res, ShouldNotBeNil)
			So(res.DataType, ShouldResemble, &DataType{
				Name:      "BIGINT",
				Source:    "BIGINT(20)",
				Number:    MySqlLexerBIGINT,
				HasLength: true,
				Length:    20,
			})
			So(res.ColumnConstraint, ShouldResemble, &ColumnConstraint{
				Key: true,
			})
		})

		Convey("ANTO_INCREMENT", func() {
			str := "bigint(20) NULL AUTO_INCREMENT UNIQUE KEY"
			p := prepare(str)
			res := v.VisitColumnDefinition(p.ColumnDefinition().(*ColumnDefinitionContext)).(*ColumnDefinition)

			So(res, ShouldNotBeNil)
			So(res.DataType, ShouldResemble, &DataType{
				Name:      "BIGINT",
				Source:    "BIGINT(20)",
				Number:    MySqlLexerBIGINT,
				HasLength: true,
				Length:    20,
			})
			So(res.ColumnConstraint, ShouldResemble, &ColumnConstraint{
				AutoIncrement: true,
				Unique:        true,
			})
		})

		Convey("DEFAULT NULL", func() {
			str := "bigint(20) NULL DEFAULT NULL AUTO_INCREMENT UNIQUE KEY"
			p := prepare(str)
			res := v.VisitColumnDefinition(p.ColumnDefinition().(*ColumnDefinitionContext)).(*ColumnDefinition)

			So(res, ShouldNotBeNil)
			So(res.DataType, ShouldResemble, &DataType{
				Name:      "BIGINT",
				Source:    "BIGINT(20)",
				Number:    MySqlLexerBIGINT,
				HasLength: true,
				Length:    20,
			})
			So(res.ColumnConstraint, ShouldResemble, &ColumnConstraint{
				AutoIncrement: true,
				Unique:        true,
				DefaultValue: &DefaultValue{
					Value: "",
					Is:    false,
				},
			})
		})

		Convey("DEFAULT Value", func() {
			str := "varchar(20) DEFAULT '' AUTO_INCREMENT UNIQUE KEY"
			p := prepare(str)
			res := v.VisitColumnDefinition(p.ColumnDefinition().(*ColumnDefinitionContext)).(*ColumnDefinition)

			So(res, ShouldNotBeNil)
			So(res.DataType, ShouldResemble, &DataType{
				Name:      "VARCHAR",
				Source:    "VARCHAR(20)",
				Number:    MySqlLexerVARCHAR,
				HasLength: true,
				Length:    20,
			})
			So(res.ColumnConstraint, ShouldResemble, &ColumnConstraint{
				AutoIncrement: true,
				Unique:        true,
				DefaultValue: &DefaultValue{
					Value: "",
					Is:    true,
				},
			})
		})

	})
}

func TestVisitDataType(t *testing.T) {
	v := new(Visitor)

	Convey("TestDataType", t, func() {
		Convey("TestDataType - String", func() {
			testData := map[string]DataType{
				`CHAR(10)`: DataType{
					Name:      "CHAR",
					Number:    MySqlLexerCHAR,
					Source:    "CHAR(10)",
					HasLength: true,
					Length:    10,
				},
				`CHARACTER(10)`: DataType{
					Name:      "CHARACTER",
					Number:    MySqlLexerCHARACTER,
					Source:    "CHARACTER(10)",
					HasLength: true,
					Length:    10,
				},
				`VARCHAR(10)`: DataType{
					Name:      "VARCHAR",
					Number:    MySqlLexerVARCHAR,
					Source:    "VARCHAR(10)",
					HasLength: true,
					Length:    10,
				},
				`TINYTEXT`: DataType{
					Name:   "TINYTEXT",
					Source: "TINYTEXT",
					Number: MySqlLexerTINYTEXT,
				},
				`TEXT`: DataType{
					Name:   "TEXT",
					Source: "TEXT",
					Number: MySqlLexerTEXT,
				},
				`MEDIUMTEXT`: DataType{
					Name:   "MEDIUMTEXT",
					Source: "MEDIUMTEXT",
					Number: MySqlLexerMEDIUMTEXT,
				},
				`LONGTEXT`: DataType{
					Name:   "LONGTEXT",
					Source: "LONGTEXT",
					Number: MySqlLexerLONGTEXT,
				},
				`NCHAR(20)`: DataType{
					Name:      "NCHAR",
					Number:    MySqlLexerNCHAR,
					Source:    "NCHAR(20)",
					HasLength: true,
					Length:    20,
				},
				`NVARCHAR(20)`: DataType{
					Name:      "NVARCHAR",
					Number:    MySqlLexerNVARCHAR,
					Source:    "NVARCHAR(20)",
					HasLength: true,
					Length:    20,
				},
				`LONG`: DataType{
					Name:   "LONG",
					Source: "LONG",
					Number: MySqlLexerLONG,
				},
			}
			for str, dt := range testData {
				p := prepare(str)
				res := v.VisitDataType(p.DataType()).(*DataType)
				So(*res, ShouldResemble, dt)
			}
		})

		Convey("TestDataType - NationalString", func() {
			testData := map[string]DataType{
				`NATIONAL VARCHAR(255)`: DataType{
					Name:       "VARCHAR",
					Number:     MySqlLexerVARCHAR,
					Source:     "NATIONAL VARCHAR(255)",
					HasLength:  true,
					Length:     255,
					IsNational: true,
				},
				`NATIONAL CHARACTER(255) BINARY`: DataType{
					Name:       "CHARACTER",
					Number:     MySqlLexerCHARACTER,
					Source:     "NATIONAL CHARACTER(255) BINARY",
					HasLength:  true,
					Length:     255,
					IsNational: true,
					IsBinary:   true,
				},
				`NCHAR VARCHAR(255) BINARY`: DataType{
					Name:      "VARCHAR",
					Number:    MySqlLexerVARCHAR,
					Source:    "NCHAR VARCHAR(255) BINARY",
					HasLength: true,
					Length:    255,
					IsBinary:  true,
					IsNChar:   true,
				},
				`NCHAR VARCHAR(200)`: DataType{
					Name:      "VARCHAR",
					Number:    MySqlLexerVARCHAR,
					Source:    "NCHAR VARCHAR(200)",
					HasLength: true,
					Length:    200,
					IsNChar:   true,
				},
			}

			for str, dt := range testData {
				p := prepare(str)
				res := v.VisitDataType(p.DataType()).(*DataType)
				So(*res, ShouldResemble, dt)
			}
		})

		Convey("TestDataType - NotionalVaringString", func() {
			testData := map[string]DataType{
				`NATIONAL CHAR VARYING (255)`: DataType{
					Name:       "CHAR",
					Number:     MySqlLexerCHAR,
					Source:     "NATIONAL CHAR(255) VARYING",
					HasLength:  true,
					Length:     255,
					IsNational: true,
					IsVarying:  true,
				},
				`NATIONAL CHAR VARYING (255) BINARY`: DataType{
					Name:       "CHAR",
					Number:     MySqlLexerCHAR,
					Source:     "NATIONAL CHAR(255) BINARY VARYING",
					HasLength:  true,
					Length:     255,
					IsNational: true,
					IsBinary:   true,
					IsVarying:  true,
				},
				`NATIONAL CHARACTER VARYING (255)`: DataType{
					Name:       "CHARACTER",
					Number:     MySqlLexerCHARACTER,
					Source:     "NATIONAL CHARACTER(255) VARYING",
					HasLength:  true,
					Length:     255,
					IsNational: true,
					IsVarying:  true,
				},
				`NATIONAL CHARACTER VARYING (255) BINARY`: DataType{
					Name:       "CHARACTER",
					Number:     MySqlLexerCHARACTER,
					Source:     "NATIONAL CHARACTER(255) BINARY VARYING",
					HasLength:  true,
					Length:     255,
					IsNational: true,
					IsBinary:   true,
					IsVarying:  true,
				},
			}

			for str, dt := range testData {
				p := prepare(str)
				res := v.VisitDataType(p.DataType()).(*DataType)
				So(*res, ShouldResemble, dt)
			}
		})

		Convey("TestDataType - DimensionDataType", func() {
			testData := map[string]DataType{
				`TINYINT(1)`: DataType{
					Name:      "TINYINT",
					Number:    MySqlParserTINYINT,
					Source:    "TINYINT(1)",
					HasLength: true,
					Length:    1,
				},
				`TINYINT(1) SIGNED`: DataType{
					Name:      "TINYINT",
					Number:    MySqlParserTINYINT,
					Source:    "TINYINT(1) SIGNED",
					HasLength: true,
					Length:    1,
					IsSigned:  true,
				},
				`TINYINT(1) UNSIGNED`: DataType{
					Name:       "TINYINT",
					Number:     MySqlParserTINYINT,
					Source:     "TINYINT(1) UNSIGNED",
					HasLength:  true,
					Length:     1,
					IsUnsigned: true,
				},
				`TINYINT(1) UNSIGNED ZEROFILL`: DataType{
					Name:       "TINYINT",
					Number:     MySqlParserTINYINT,
					Source:     "TINYINT(1) UNSIGNED ZEROFILL",
					HasLength:  true,
					Length:     1,
					IsUnsigned: true,
					IsZeroFill: true,
				},
				`SMALLINT(10)`: DataType{
					Name:      "SMALLINT",
					Number:    MySqlParserSMALLINT,
					Source:    "SMALLINT(10)",
					HasLength: true,
					Length:    10,
				},
				`SMALLINT(10) SIGNED`: DataType{
					Name:      "SMALLINT",
					Number:    MySqlParserSMALLINT,
					Source:    "SMALLINT(10) SIGNED",
					HasLength: true,
					Length:    10,
					IsSigned:  true,
				},
				`SMALLINT(10) UNSIGNED`: DataType{
					Name:       "SMALLINT",
					Number:     MySqlParserSMALLINT,
					Source:     "SMALLINT(10) UNSIGNED",
					HasLength:  true,
					Length:     10,
					IsUnsigned: true,
				},
				`SMALLINT(10) ZEROFILL`: DataType{
					Name:       "SMALLINT",
					Number:     MySqlParserSMALLINT,
					Source:     "SMALLINT(10) ZEROFILL",
					HasLength:  true,
					Length:     10,
					IsZeroFill: true,
				},
				`MEDIUMINT(10)`: DataType{
					Name:      "MEDIUMINT",
					Number:    MySqlParserMEDIUMINT,
					Source:    "MEDIUMINT(10)",
					HasLength: true,
					Length:    10,
				},
				`MEDIUMINT(10) SIGNED`: DataType{
					Name:      "MEDIUMINT",
					Number:    MySqlParserMEDIUMINT,
					Source:    "MEDIUMINT(10) SIGNED",
					HasLength: true,
					Length:    10,
					IsSigned:  true,
				},
				`MEDIUMINT(10) UNSIGNED`: DataType{
					Name:       "MEDIUMINT",
					Number:     MySqlParserMEDIUMINT,
					Source:     "MEDIUMINT(10) UNSIGNED",
					HasLength:  true,
					Length:     10,
					IsUnsigned: true,
				},
				`MEDIUMINT(10) ZEROFILL`: DataType{
					Name:       "MEDIUMINT",
					Number:     MySqlParserMEDIUMINT,
					Source:     "MEDIUMINT(10) ZEROFILL",
					HasLength:  true,
					Length:     10,
					IsZeroFill: true,
				},
				`INT(10)`: DataType{
					Name:      "INT",
					Number:    MySqlParserINT,
					Source:    "INT(10)",
					HasLength: true,
					Length:    10,
				},
				`INT(10) SIGNED`: DataType{
					Name:      "INT",
					Number:    MySqlParserINT,
					Source:    "INT(10) SIGNED",
					HasLength: true,
					Length:    10,
					IsSigned:  true,
				},
				`INT(10) UNSIGNED`: DataType{
					Name:       "INT",
					Number:     MySqlParserINT,
					Source:     "INT(10) UNSIGNED",
					HasLength:  true,
					Length:     10,
					IsUnsigned: true,
				},
				`INT(10) ZEROFILL`: DataType{
					Name:       "INT",
					Number:     MySqlParserINT,
					Source:     "INT(10) ZEROFILL",
					HasLength:  true,
					Length:     10,
					IsZeroFill: true,
				},
				`INTEGER(10)`: DataType{
					Name:      "INTEGER",
					Number:    MySqlParserINTEGER,
					Source:    "INTEGER(10)",
					HasLength: true,
					Length:    10,
				},
				`INTEGER(10) SIGNED`: DataType{
					Name:      "INTEGER",
					Number:    MySqlParserINTEGER,
					Source:    "INTEGER(10) SIGNED",
					HasLength: true,
					Length:    10,
					IsSigned:  true,
				},
				`INTEGER(10) UNSIGNED`: DataType{
					Name:       "INTEGER",
					Number:     MySqlParserINTEGER,
					Source:     "INTEGER(10) UNSIGNED",
					HasLength:  true,
					Length:     10,
					IsUnsigned: true,
				},
				`INTEGER(10) ZEROFILL`: DataType{
					Name:       "INTEGER",
					Number:     MySqlParserINTEGER,
					Source:     "INTEGER(10) ZEROFILL",
					HasLength:  true,
					Length:     10,
					IsZeroFill: true,
				},
				`BIGINT(20)`: DataType{
					Name:      "BIGINT",
					Number:    MySqlParserBIGINT,
					Source:    "BIGINT(20)",
					HasLength: true,
					Length:    20,
				},
				`BIGINT(20) SIGNED`: DataType{
					Name:      "BIGINT",
					Number:    MySqlParserBIGINT,
					Source:    "BIGINT(20) SIGNED",
					HasLength: true,
					Length:    20,
					IsSigned:  true,
				},
				`BIGINT(20) UNSIGNED`: DataType{
					Name:       "BIGINT",
					Number:     MySqlParserBIGINT,
					Source:     "BIGINT(20) UNSIGNED",
					HasLength:  true,
					Length:     20,
					IsUnsigned: true,
				},
				`BIGINT(20) ZEROFILL`: DataType{
					Name:       "BIGINT",
					Number:     MySqlParserBIGINT,
					Source:     "BIGINT(20) ZEROFILL",
					HasLength:  true,
					Length:     20,
					IsZeroFill: true,
				},
				`MIDDLEINT(20)`: DataType{
					Name:      "MIDDLEINT",
					Number:    MySqlParserMIDDLEINT,
					Source:    "MIDDLEINT(20)",
					HasLength: true,
					Length:    20,
				},
				`MIDDLEINT(20) SIGNED`: DataType{
					Name:      "MIDDLEINT",
					Number:    MySqlParserMIDDLEINT,
					Source:    "MIDDLEINT(20) SIGNED",
					HasLength: true,
					Length:    20,
					IsSigned:  true,
				},
				`MIDDLEINT(20) UNSIGNED`: DataType{
					Name:       "MIDDLEINT",
					Number:     MySqlParserMIDDLEINT,
					Source:     "MIDDLEINT(20) UNSIGNED",
					HasLength:  true,
					Length:     20,
					IsUnsigned: true,
				},
				`MIDDLEINT(20) ZEROFILL`: DataType{
					Name:       "MIDDLEINT",
					Number:     MySqlParserMIDDLEINT,
					Source:     "MIDDLEINT(20) ZEROFILL",
					HasLength:  true,
					Length:     20,
					IsZeroFill: true,
				},
				`INT1(2)`: DataType{
					Name:      "INT1",
					Number:    MySqlParserINT1,
					Source:    "INT1(2)",
					HasLength: true,
					Length:    2,
				},
				`INT1(2) SIGNED`: DataType{
					Name:      "INT1",
					Number:    MySqlParserINT1,
					Source:    "INT1(2) SIGNED",
					HasLength: true,
					Length:    2,
					IsSigned:  true,
				},
				`INT1(2) UNSIGNED`: DataType{
					Name:       "INT1",
					Number:     MySqlParserINT1,
					Source:     "INT1(2) UNSIGNED",
					HasLength:  true,
					Length:     2,
					IsUnsigned: true,
				},
				`INT1(2) ZEROFILL`: DataType{
					Name:       "INT1",
					Number:     MySqlParserINT1,
					Source:     "INT1(2) ZEROFILL",
					HasLength:  true,
					Length:     2,
					IsZeroFill: true,
				},
				`INT2(2)`: DataType{
					Name:      "INT2",
					Number:    MySqlParserINT2,
					Source:    "INT2(2)",
					HasLength: true,
					Length:    2,
				},
				`INT2(2) SIGNED`: DataType{
					Name:      "INT2",
					Number:    MySqlParserINT2,
					Source:    "INT2(2) SIGNED",
					HasLength: true,
					Length:    2,
					IsSigned:  true,
				},
				`INT2(2) UNSIGNED`: DataType{
					Name:       "INT2",
					Number:     MySqlParserINT2,
					Source:     "INT2(2) UNSIGNED",
					HasLength:  true,
					Length:     2,
					IsUnsigned: true,
				},
				`INT2(2) ZEROFILL`: DataType{
					Name:       "INT2",
					Number:     MySqlParserINT2,
					Source:     "INT2(2) ZEROFILL",
					HasLength:  true,
					Length:     2,
					IsZeroFill: true,
				},
				`INT3(20)`: DataType{
					Name:      "INT3",
					Number:    MySqlParserINT3,
					Source:    "INT3(20)",
					HasLength: true,
					Length:    20,
				},
				`INT3(3) SIGNED`: DataType{
					Name:      "INT3",
					Number:    MySqlParserINT3,
					Source:    "INT3(3) SIGNED",
					HasLength: true,
					Length:    3,
					IsSigned:  true,
				},
				`INT3(3) UNSIGNED`: DataType{
					Name:       "INT3",
					Number:     MySqlParserINT3,
					Source:     "INT3(3) UNSIGNED",
					HasLength:  true,
					Length:     3,
					IsUnsigned: true,
				},
				`INT3(3) ZEROFILL`: DataType{
					Name:       "INT3",
					Number:     MySqlParserINT3,
					Source:     "INT3(3) ZEROFILL",
					HasLength:  true,
					Length:     3,
					IsZeroFill: true,
				},
				`INT4(4)`: DataType{
					Name:      "INT4",
					Number:    MySqlParserINT4,
					Source:    "INT4(4)",
					HasLength: true,
					Length:    4,
				},
				`INT4(4) SIGNED`: DataType{
					Name:      "INT4",
					Number:    MySqlParserINT4,
					Source:    "INT4(4) SIGNED",
					HasLength: true,
					Length:    4,
					IsSigned:  true,
				},
				`INT4(4) UNSIGNED`: DataType{
					Name:       "INT4",
					Number:     MySqlParserINT4,
					Source:     "INT4(4) UNSIGNED",
					HasLength:  true,
					Length:     4,
					IsUnsigned: true,
				},
				`INT4(4) ZEROFILL`: DataType{
					Name:       "INT4",
					Number:     MySqlParserINT4,
					Source:     "INT4(4) ZEROFILL",
					HasLength:  true,
					Length:     4,
					IsZeroFill: true,
				},
				`INT8(8)`: DataType{
					Name:      "INT8",
					Number:    MySqlParserINT8,
					Source:    "INT8(8)",
					HasLength: true,
					Length:    8,
				},
				`INT8(8) SIGNED`: DataType{
					Name:      "INT8",
					Number:    MySqlParserINT8,
					Source:    "INT8(8) SIGNED",
					HasLength: true,
					Length:    8,
					IsSigned:  true,
				},
				`INT8(8) UNSIGNED`: DataType{
					Name:       "INT8",
					Number:     MySqlParserINT8,
					Source:     "INT8(8) UNSIGNED",
					HasLength:  true,
					Length:     8,
					IsUnsigned: true,
				},
				`INT8(8) ZEROFILL`: DataType{
					Name:       "INT8",
					Number:     MySqlParserINT8,
					Source:     "INT8(8) ZEROFILL",
					HasLength:  true,
					Length:     8,
					IsZeroFill: true,
				},
				`REAL(8,10) ZEROFILL`: DataType{
					Name:         "REAL",
					Number:       MySqlParserREAL,
					Source:       "REAL(8,10) ZEROFILL",
					HasTwoLength: true,
					Len1:         8,
					Len2:         10,
					IsZeroFill:   true,
				},
				`REAL ZEROFILL`: DataType{
					Name:       "REAL",
					Number:     MySqlParserREAL,
					Source:     "REAL ZEROFILL",
					IsZeroFill: true,
				},
				`REAL SIGNED ZEROFILL`: DataType{
					Name:       "REAL",
					Number:     MySqlParserREAL,
					Source:     "REAL SIGNED ZEROFILL",
					IsSigned:   true,
					IsZeroFill: true,
				},
				`REAL UNSIGNED ZEROFILL`: DataType{
					Name:       "REAL",
					Number:     MySqlParserREAL,
					Source:     "REAL UNSIGNED ZEROFILL",
					IsUnsigned: true,
					IsZeroFill: true,
				},
				`DOUBLE(8,10) ZEROFILL`: DataType{
					Name:         "DOUBLE",
					Number:       MySqlParserDOUBLE,
					Source:       "DOUBLE(8,10) ZEROFILL",
					HasTwoLength: true,
					Len1:         8,
					Len2:         10,
					IsZeroFill:   true,
				},
				// TODO:
				`DOUBLE PRECISION (8,10) ZEROFILL`: DataType{
					Name:         "DOUBLE",
					Number:       MySqlParserDOUBLE,
					Source:       "DOUBLE(8,10) ZEROFILL",
					HasTwoLength: true,
					Len1:         8,
					Len2:         10,
					IsZeroFill:   true,
				},
				`DOUBLE ZEROFILL`: DataType{
					Name:       "DOUBLE",
					Number:     MySqlParserDOUBLE,
					Source:     "DOUBLE ZEROFILL",
					IsZeroFill: true,
				},
				`DOUBLE SIGNED ZEROFILL`: DataType{
					Name:       "DOUBLE",
					Number:     MySqlParserDOUBLE,
					Source:     "DOUBLE SIGNED ZEROFILL",
					IsSigned:   true,
					IsZeroFill: true,
				},
				`DOUBLE UNSIGNED ZEROFILL`: DataType{
					Name:       "DOUBLE",
					Number:     MySqlParserDOUBLE,
					Source:     "DOUBLE UNSIGNED ZEROFILL",
					IsUnsigned: true,
					IsZeroFill: true,
				},
				`DECIMAL(8,10) ZEROFILL`: DataType{
					Name:         "DECIMAL",
					Number:       MySqlParserDECIMAL,
					Source:       "DECIMAL(8,10) ZEROFILL",
					HasTwoLength: true,
					Len1:         8,
					Len2:         10,
					IsZeroFill:   true,
				},
				`DECIMAL ZEROFILL`: DataType{
					Name:       "DECIMAL",
					Number:     MySqlParserDECIMAL,
					Source:     "DECIMAL ZEROFILL",
					IsZeroFill: true,
				},
				`DECIMAL SIGNED ZEROFILL`: DataType{
					Name:       "DECIMAL",
					Number:     MySqlParserDECIMAL,
					Source:     "DECIMAL SIGNED ZEROFILL",
					IsSigned:   true,
					IsZeroFill: true,
				},
				`DECIMAL UNSIGNED ZEROFILL`: DataType{
					Name:       "DECIMAL",
					Number:     MySqlParserDECIMAL,
					Source:     "DECIMAL UNSIGNED ZEROFILL",
					IsUnsigned: true,
					IsZeroFill: true,
				},
				`DEC(8,10) ZEROFILL`: DataType{
					Name:         "DEC",
					Number:       MySqlParserDEC,
					Source:       "DEC(8,10) ZEROFILL",
					HasTwoLength: true,
					Len1:         8,
					Len2:         10,
					IsZeroFill:   true,
				},
				`DEC ZEROFILL`: DataType{
					Name:       "DEC",
					Number:     MySqlParserDEC,
					Source:     "DEC ZEROFILL",
					IsZeroFill: true,
				},
				`DEC SIGNED ZEROFILL`: DataType{
					Name:       "DEC",
					Number:     MySqlParserDEC,
					Source:     "DEC SIGNED ZEROFILL",
					IsSigned:   true,
					IsZeroFill: true,
				},
				`DEC UNSIGNED ZEROFILL`: DataType{
					Name:       "DEC",
					Number:     MySqlParserDEC,
					Source:     "DEC UNSIGNED ZEROFILL",
					IsUnsigned: true,
					IsZeroFill: true,
				},
				`FIXED(8,10) ZEROFILL`: DataType{
					Name:         "FIXED",
					Number:       MySqlParserFIXED,
					Source:       "FIXED(8,10) ZEROFILL",
					HasTwoLength: true,
					Len1:         8,
					Len2:         10,
					IsZeroFill:   true,
				},
				`FIXED ZEROFILL`: DataType{
					Name:       "FIXED",
					Number:     MySqlParserFIXED,
					Source:     "FIXED ZEROFILL",
					IsZeroFill: true,
				},
				`FIXED SIGNED ZEROFILL`: DataType{
					Name:       "FIXED",
					Number:     MySqlParserFIXED,
					Source:     "FIXED SIGNED ZEROFILL",
					IsSigned:   true,
					IsZeroFill: true,
				},
				`FIXED UNSIGNED ZEROFILL`: DataType{
					Name:       "FIXED",
					Number:     MySqlParserFIXED,
					Source:     "FIXED UNSIGNED ZEROFILL",
					IsUnsigned: true,
					IsZeroFill: true,
				},
				`NUMERIC(8,10) ZEROFILL`: DataType{
					Name:         "NUMERIC",
					Number:       MySqlParserNUMERIC,
					Source:       "NUMERIC(8,10) ZEROFILL",
					HasTwoLength: true,
					Len1:         8,
					Len2:         10,
					IsZeroFill:   true,
				},
				`NUMERIC ZEROFILL`: DataType{
					Name:       "NUMERIC",
					Number:     MySqlParserNUMERIC,
					Source:     "NUMERIC ZEROFILL",
					IsZeroFill: true,
				},
				`NUMERIC SIGNED ZEROFILL`: DataType{
					Name:       "NUMERIC",
					Number:     MySqlParserNUMERIC,
					Source:     "NUMERIC SIGNED ZEROFILL",
					IsSigned:   true,
					IsZeroFill: true,
				},
				`NUMERIC UNSIGNED ZEROFILL`: DataType{
					Name:       "NUMERIC",
					Number:     MySqlParserNUMERIC,
					Source:     "NUMERIC UNSIGNED ZEROFILL",
					IsUnsigned: true,
					IsZeroFill: true,
				},
				`FLOAT(8,10) ZEROFILL`: DataType{
					Name:         "FLOAT",
					Number:       MySqlParserFLOAT,
					Source:       "FLOAT(8,10) ZEROFILL",
					HasTwoLength: true,
					Len1:         8,
					Len2:         10,
					IsZeroFill:   true,
				},
				`FLOAT ZEROFILL`: DataType{
					Name:       "FLOAT",
					Number:     MySqlParserFLOAT,
					Source:     "FLOAT ZEROFILL",
					IsZeroFill: true,
				},
				`FLOAT SIGNED ZEROFILL`: DataType{
					Name:       "FLOAT",
					Number:     MySqlParserFLOAT,
					Source:     "FLOAT SIGNED ZEROFILL",
					IsSigned:   true,
					IsZeroFill: true,
				},
				`FLOAT UNSIGNED ZEROFILL`: DataType{
					Name:       "FLOAT",
					Number:     MySqlParserFLOAT,
					Source:     "FLOAT UNSIGNED ZEROFILL",
					IsUnsigned: true,
					IsZeroFill: true,
				},
				`FLOAT4(8,10) ZEROFILL`: DataType{
					Name:         "FLOAT4",
					Number:       MySqlParserFLOAT4,
					Source:       "FLOAT4(8,10) ZEROFILL",
					HasTwoLength: true,
					Len1:         8,
					Len2:         10,
					IsZeroFill:   true,
				},
				`FLOAT4 ZEROFILL`: DataType{
					Name:       "FLOAT4",
					Number:     MySqlParserFLOAT4,
					Source:     "FLOAT4 ZEROFILL",
					IsZeroFill: true,
				},
				`FLOAT4 SIGNED ZEROFILL`: DataType{
					Name:       "FLOAT4",
					Number:     MySqlParserFLOAT4,
					Source:     "FLOAT4 SIGNED ZEROFILL",
					IsSigned:   true,
					IsZeroFill: true,
				},
				`FLOAT4 UNSIGNED ZEROFILL`: DataType{
					Name:       "FLOAT4",
					Number:     MySqlParserFLOAT4,
					Source:     "FLOAT4 UNSIGNED ZEROFILL",
					IsUnsigned: true,
					IsZeroFill: true,
				},
				`FLOAT8(8,10) ZEROFILL`: DataType{
					Name:         "FLOAT8",
					Number:       MySqlParserFLOAT8,
					Source:       "FLOAT8(8,10) ZEROFILL",
					HasTwoLength: true,
					Len1:         8,
					Len2:         10,
					IsZeroFill:   true,
				},
				`FLOAT8 ZEROFILL`: DataType{
					Name:       "FLOAT8",
					Number:     MySqlParserFLOAT8,
					Source:     "FLOAT8 ZEROFILL",
					IsZeroFill: true,
				},
				`FLOAT8 SIGNED ZEROFILL`: DataType{
					Name:       "FLOAT8",
					Number:     MySqlParserFLOAT8,
					Source:     "FLOAT8 SIGNED ZEROFILL",
					IsSigned:   true,
					IsZeroFill: true,
				},
				`FLOAT8 UNSIGNED ZEROFILL`: DataType{
					Name:       "FLOAT8",
					Number:     MySqlParserFLOAT8,
					Source:     "FLOAT8 UNSIGNED ZEROFILL",
					IsUnsigned: true,
					IsZeroFill: true,
				},
				`BIT`: DataType{
					Name:   "BIT",
					Number: MySqlParserBIT,
					Source: "BIT",
				},
				`BIT(1)`: DataType{
					Name:      "BIT",
					Number:    MySqlParserBIT,
					Source:    "BIT(1)",
					HasLength: true,
					Length:    1,
				},
				`TIME`: DataType{
					Name:   "TIME",
					Number: MySqlParserTIME,
					Source: "TIME",
				},
				`TIMESTAMP`: DataType{
					Name:   "TIMESTAMP",
					Number: MySqlParserTIMESTAMP,
					Source: "TIMESTAMP",
				},
				`DATETIME`: DataType{
					Name:   "DATETIME",
					Number: MySqlParserDATETIME,
					Source: "DATETIME",
				},
				`BINARY`: DataType{
					Name:   "BINARY",
					Number: MySqlParserBINARY,
					Source: "BINARY",
				},
				`VARBINARY`: DataType{
					Name:   "VARBINARY",
					Number: MySqlParserVARBINARY,
					Source: "VARBINARY",
				},
				`BLOB`: DataType{
					Name:   "BLOB",
					Number: MySqlParserBLOB,
					Source: "BLOB",
				},
				`YEAR`: DataType{
					Name:   "YEAR",
					Number: MySqlParserYEAR,
					Source: "YEAR",
				},
			}

			for str, dt := range testData {
				p := prepare(str)
				res := v.VisitDataType(p.DataType()).(*DataType)
				So(*res, ShouldResemble, dt)
			}

			testData = map[string]DataType{
				`TINYINT(1) UNSIGNED`: DataType{
					Name:       "TINYINT",
					Number:     MySqlParserTINYINT,
					Source:     "TINYINT(1) UNSIGNED",
					HasLength:  true,
					Length:     1,
					IsUnsigned: true,
				},
				`SMALLINT UNSIGNED`: DataType{
					Name:       "SMALLINT",
					Number:     MySqlParserSMALLINT,
					Source:     "SMALLINT UNSIGNED",
					IsUnsigned: true,
				},
				`BIGINT UNSIGNED`: DataType{
					Name:       "BIGINT",
					Number:     MySqlParserBIGINT,
					Source:     "BIGINT UNSIGNED",
					IsUnsigned: true,
				},
			}
			for str, dt := range testData {
				p := prepare(str)
				res := v.VisitDataType(p.DataType()).(*DataType)
				So(*res, ShouldResemble, dt)
			}
		})

		Convey("TestDataType - Simple", func() {
			testData := map[string]DataType{
				`DATE`: DataType{
					Name:   "DATE",
					Number: MySqlLexerDATE,
					Source: "DATE",
				},
				`TINYBLOB`: DataType{
					Name:   "TINYBLOB",
					Number: MySqlLexerTINYBLOB,
					Source: "TINYBLOB",
				},
				`MEDIUMBLOB`: DataType{
					Name:   "MEDIUMBLOB",
					Number: MySqlLexerMEDIUMBLOB,
					Source: "MEDIUMBLOB",
				},
				`LONGBLOB`: DataType{
					Name:   "LONGBLOB",
					Number: MySqlLexerLONGBLOB,
					Source: "LONGBLOB",
				},
				`BOOL`: DataType{
					Name:   "BOOL",
					Number: MySqlLexerBOOL,
					Source: "BOOL",
				},
				`BOOLEAN`: DataType{
					Name:   "BOOLEAN",
					Number: MySqlLexerBOOLEAN,
					Source: "BOOLEAN",
				},
				`SERIAL`: DataType{
					Name:   "SERIAL",
					Number: MySqlLexerSERIAL,
					Source: "SERIAL",
				},
			}

			for str, dt := range testData {
				p := prepare(str)
				res := v.VisitDataType(p.DataType()).(*DataType)
				So(*res, ShouldResemble, dt)
			}
		})

		Convey("TestDataType - Collection", func() {
			testData := map[string]*DataType{
				`ENUM('1','2')`: {
					Number:            MySqlLexerENUM,
					Name:              "ENUM",
					Source:            "ENUM('1','2')",
					CollectionOptions: []string{"1", "2"},
				},
				`SET('A','B')`: {
					Number:            MySqlLexerSET,
					Name:              "SET",
					Source:            "SET('A','B')",
					CollectionOptions: []string{"A", "B"},
				},
				`SET('A','B') BINARY`: {
					Number:            MySqlLexerSET,
					Name:              "SET",
					Source:            "SET('A','B') BINARY",
					CollectionOptions: []string{"A", "B"},
					IsBinary:          true,
				},
			}

			for str, data := range testData {
				p := prepare(str)
				res := v.VisitDataType(p.DataType()).(*DataType)
				So(res, ShouldResemble, data)
			}
		})

		Convey("TestDataType - Spatial", func() {
			testData := map[string]DataType{
				`GEOMETRYCOLLECTION`: DataType{
					Name:   "GEOMETRYCOLLECTION",
					Number: MySqlLexerGEOMETRYCOLLECTION,
					Source: "GEOMETRYCOLLECTION",
				},
				`GEOMCOLLECTION`: DataType{
					Name:   "GEOMCOLLECTION",
					Number: MySqlLexerGEOMCOLLECTION,
					Source: "GEOMCOLLECTION",
				},
				`LINESTRING`: DataType{
					Name:   "LINESTRING",
					Number: MySqlLexerLINESTRING,
					Source: "LINESTRING",
				},
				`MULTILINESTRING`: DataType{
					Name:   "MULTILINESTRING",
					Number: MySqlLexerMULTILINESTRING,
					Source: "MULTILINESTRING",
				},
				`MULTIPOINT`: DataType{
					Name:   "MULTIPOINT",
					Number: MySqlLexerMULTIPOINT,
					Source: "MULTIPOINT",
				},
				`MULTIPOLYGON`: DataType{
					Name:   "MULTIPOLYGON",
					Number: MySqlLexerMULTIPOLYGON,
					Source: "MULTIPOLYGON",
				},
				`POINT`: DataType{
					Name:   "POINT",
					Number: MySqlLexerPOINT,
					Source: "POINT",
				},
				`POLYGON`: DataType{
					Name:   "POLYGON",
					Number: MySqlLexerPOLYGON,
					Source: "POLYGON",
				},
				`JSON`: DataType{
					Name:   "JSON",
					Number: MySqlLexerJSON,
					Source: "JSON",
				},
				`GEOMETRY`: DataType{
					Name:   "GEOMETRY",
					Number: MySqlLexerGEOMETRY,
					Source: "GEOMETRY",
				},
			}

			for str, dt := range testData {
				p := prepare(str)
				res := v.VisitDataType(p.DataType()).(*DataType)
				So(*res, ShouldResemble, dt)
			}
		})

		Convey("TestDataType - LongVarchar", func() {
			testData := map[string]DataType{
				`LONG`: DataType{
					Name:   "LONG",
					Number: MySqlLexerLONG,
					Source: "LONG",
				},
				`LONG VARCHAR`: DataType{
					Name:      "LONG",
					Number:    MySqlLexerLONG,
					Source:    "LONG VARCHAR",
					IsVarchar: true,
				},
				`LONG VARCHAR BINARY`: DataType{
					Name:      "LONG",
					Number:    MySqlLexerLONG,
					Source:    "LONG VARCHAR BINARY",
					IsVarchar: true,
					IsBinary:  true,
				},
				`LONG VARCHAR BINARY CHARACTER SET 'utf8'`: DataType{
					Name:        "LONG",
					Number:      MySqlLexerLONG,
					Source:      "LONG VARCHAR BINARY CHARACTER SET 'utf8'",
					IsVarchar:   true,
					IsBinary:    true,
					IsCharacter: true,
					CharsetName: "utf8",
				},
				`LONG VARCHAR BINARY CHARSET 'utf8'`: DataType{
					Name:        "LONG",
					Number:      MySqlLexerLONG,
					Source:      "LONG VARCHAR BINARY CHARSET SET 'utf8'",
					IsVarchar:   true,
					IsBinary:    true,
					IsCharset:   true,
					CharsetName: "utf8",
				},
			}

			for str, dt := range testData {
				p := prepare(str)
				res := v.VisitDataType(p.DataType()).(*DataType)
				So(*res, ShouldResemble, dt)
			}
		})

		Convey("TestDataType - LongVarbinary", func() {
			testData := map[string]DataType{
				`LONG VARBINARY  `: DataType{
					Name:   "LONG",
					Number: MySqlLexerLONG,
					Source: "LONG",
				},
			}

			for str, dt := range testData {
				p := prepare(str)
				res := v.VisitDataType(p.DataType()).(*DataType)
				So(*res, ShouldResemble, dt)
			}
		})

	})
}

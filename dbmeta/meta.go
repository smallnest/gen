package dbmeta

import (
	"bytes"
	"database/sql"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/iancoleman/strcase"
)

type metaDataLoader func(db *sql.DB, sqlType, sqlDatabase, tableName string) (DbTableMeta, error)

var metaDataFuncs = make(map[string]metaDataLoader)

func init() {
	metaDataFuncs["sqlite3"] = NewSqliteMeta
	metaDataFuncs["sqlite"] = NewSqliteMeta
	metaDataFuncs["mssql"] = NewMsSqlMeta
	metaDataFuncs["postgres"] = NewPostgresMeta
	metaDataFuncs["mysql"] = NewMysqlMeta
}

type ColumnMeta interface {
	Name() string
	String() string
	//Length() (length int64, ok bool)
	//DecimalSize() (precision, scale int64, ok bool)
	//ScanType() reflect.Type
	Nullable() bool
	DatabaseTypeName() string
	GetGoType(gureguTypes bool) (string, error)
	//IsNullable() bool
	IsPrimaryKey() bool
	IsAutoIncrement() bool
	ColumnLength() int64
}

func (ci *columnMeta) IsPrimaryKey() bool {
	return ci.isPrimaryKey
}

func (ci *columnMeta) IsAutoIncrement() bool {
	return ci.isAutoIncrement
}

func (ci *columnMeta) ColumnLength() int64 {
	l, _ := ci.ct.Length()
	return l
}

type columnMeta struct {
	index           int
	ct              *sql.ColumnType
	nullable        bool
	isPrimaryKey    bool
	isAutoIncrement bool
	colDDL          string
}

// Name returns the name or alias of the column.
func (ci *columnMeta) Name() string {
	return ci.ct.Name()
}

func (ci *columnMeta) String() string {
	return ci.colDDL
}


// Nullable reports whether the column may be null.
// If a driver does not support this property ok will be false.
func (ci *columnMeta) Nullable() bool {
	return ci.nullable
}
func (ci *columnMeta) ColDDL() string {
	return ci.colDDL
}

// DatabaseTypeName returns the database system name of the column type. If an empty
// string is returned the driver type name is not supported.
// Consult your driver documentation for a list of driver data types. Length specifiers
// are not included.
// Common type include "VARCHAR", "TEXT", "NVARCHAR", "DECIMAL", "BOOL", "INT", "BIGINT".
func (ci *columnMeta) DatabaseTypeName() string {
	return ci.ct.DatabaseTypeName()
}

func (ci *columnMeta) GetGoType(gureguTypes bool) (string, error) {
	valueType, err := sqlTypeToGoType(strings.ToLower(ci.DatabaseTypeName()), ci.nullable, gureguTypes)
	if err != nil {
		return "", err
	}

	return valueType, nil
}

type DbTableMeta interface {
	Columns() []ColumnMeta
	SqlType() string
	SqlDatabase() string
	TableName() string
	DDL() string
}
type dbTableMeta struct {
	sqlType     string
	sqlDatabase string
	tableName   string
	columns     []ColumnMeta
	ddl         string
}

func (m *dbTableMeta) SqlType() string {
	return m.sqlType
}
func (m *dbTableMeta) SqlDatabase() string {
	return m.sqlDatabase
}
func (m *dbTableMeta) TableName() string {
	return m.tableName
}
func (m *dbTableMeta) Columns() []ColumnMeta {
	return m.columns
}
func (m *dbTableMeta) DDL() string {
	return m.ddl
}

type ModelInfo struct {
	PackageName     string
	StructName      string
	ShortStructName string
	TableName       string
	Fields          []string
	DBMeta          DbTableMeta
}

// commonInitialisms is a set of common initialisms.
// Only add entries that are highly unlikely to be non-initialisms.
// For instance, "ID" is fine (Freudian code is rare), but "AND" is not.
var commonInitialisms = map[string]bool{
	"API":   true,
	"ASCII": true,
	"CPU":   true,
	"CSS":   true,
	"DNS":   true,
	"EOF":   true,
	"GUID":  true,
	"HTML":  true,
	"HTTP":  true,
	"HTTPS": true,
	"ID":    true,
	"IP":    true,
	"JSON":  true,
	"LHS":   true,
	"QPS":   true,
	"RAM":   true,
	"RHS":   true,
	"RPC":   true,
	"SLA":   true,
	"SMTP":  true,
	"SSH":   true,
	"TLS":   true,
	"TTL":   true,
	"UI":    true,
	"UID":   true,
	"UUID":  true,
	"URI":   true,
	"URL":   true,
	"UTF8":  true,
	"VM":    true,
	"XML":   true,
}

var intToWordMap = []string{
	"zero",
	"one",
	"two",
	"three",
	"four",
	"five",
	"six",
	"seven",
	"eight",
	"nine",
}

// Constants for return types of golang
const (
	golangByteArray  = "[]byte"
	gureguNullInt    = "null.Int"
	sqlNullInt       = "sql.NullInt64"
	sqlNullBool      = "sql.NullBool"
	golangInt        = "int"
	golangInt64      = "int64"
	gureguNullFloat  = "null.Float"
	sqlNullFloat     = "sql.NullFloat64"
	golangFloat      = "float"
	golangFloat32    = "float32"
	golangFloat64    = "float64"
	gureguNullString = "null.String"
	sqlNullString    = "sql.NullString"
	gureguNullTime   = "null.Time"
	golangTime       = "time.Time"
	golangBool       = "bool"
)

// GenerateStruct generates a struct for the given table.
func GenerateStruct(sqlType string,
	db *sql.DB,
	sqlDatabase,
	tableName string,
	structName string,
	pkgName string,
	jsonAnnotation bool,
	gormAnnotation bool,
	gureguTypes bool,
	jsonNameFormat string,
	verbose bool) (*ModelInfo, error) {

	dbMetaFunc, haveMeta := metaDataFuncs[sqlType]
	if !haveMeta {
		dbMetaFunc = NewUnknownMeta
	}

	dbMeta, err := dbMetaFunc(db, sqlType, sqlDatabase, tableName)
	if err != nil {
		return nil, err
	}

	fields := generateFieldsTypes(dbMeta, 0, jsonAnnotation, gormAnnotation, gureguTypes, jsonNameFormat, verbose)

	var modelInfo = &ModelInfo{
		PackageName:     pkgName,
		StructName:      structName,
		TableName:       tableName,
		ShortStructName: strings.ToLower(string(structName[0])),
		Fields:          fields,
		DBMeta:          dbMeta,
	}

	return modelInfo, nil
}

// Generate fields string
func generateFieldsTypes(dbMeta DbTableMeta,
	depth int,
	jsonAnnotation bool,
	gormAnnotation bool,
	gureguTypes bool,
	jsonNameFormat string,
	verbose bool) []string {

	jsonNameFormat = strings.ToLower(jsonNameFormat)

	var fields []string
	var field = ""
	for _, c := range dbMeta.Columns() {
		key := c.Name()
		if verbose {
			//fmt.Printf("   [%s]   [%d] field: %s type: %s\n", dbMeta.TableName(), i, key, c.DatabaseTypeName())
		}

		valueType, err := c.GetGoType(gureguTypes)

		if err != nil { // unknown type
			if verbose {
				fmt.Printf("table: %s unable to generate struct field: %s type: %s\n", dbMeta.TableName(), key, c.DatabaseTypeName())
			}
			continue
		}
		fieldName := FmtFieldName(stringifyFirstChar(key))

		var annotations []string
		if gormAnnotation == true {
			annotations = append(annotations, createGormAnnotation(c))
		}

		if jsonAnnotation == true {
			annotations = append(annotations, createJsonAnnotation(jsonNameFormat, c))
		}

		if len(annotations) > 0 {
			field = fmt.Sprintf("%s %s `%s`",
				fieldName,
				valueType,
				strings.Join(annotations, " "))
		} else {
			field = fmt.Sprintf("%s %s",
				fieldName,
				valueType)
		}

		field = fmt.Sprintf("%s //%s", field, c.String())
		fields = append(fields, field)
	}
	return fields
}

func createJsonAnnotation(jsonNameFormat string, c ColumnMeta) string {

	var jsonName string
	switch jsonNameFormat {
	case "snake":
		jsonName = strcase.ToSnake(c.Name())
	case "camel":
		jsonName = strcase.ToCamel(c.Name())
	case "lower_camel":
		jsonName = strcase.ToLowerCamel(c.Name())
	case "none":
		jsonName = c.Name()
	default:
		jsonName = c.Name()
	}
	return fmt.Sprintf("json:\"%s\"", jsonName)
}

func createGormAnnotation(c ColumnMeta) string {
	buf := bytes.Buffer{}

	key := c.Name()

	dbType := strings.ToLower(c.DatabaseTypeName())

	charLen := -1
	if strings.Contains(dbType, "varchar") {

		re := regexp.MustCompile(`[-]?\d[\d,]*[\.]?[\d{2}]*`)
		submatchall := re.FindAllString(dbType, -1)
		if len(submatchall) > 0 {
			i, err := strconv.Atoi(submatchall[0])
			if err == nil {
				charLen = i
			}
		}
	}
	buf.WriteString("gorm:\"")

	if c.IsAutoIncrement(){
		buf.WriteString("AUTO_INCREMENT;")
	}


	buf.WriteString("column:")
	buf.WriteString(key)
	buf.WriteString(";")

	buf.WriteString("type:")
	buf.WriteString(c.DatabaseTypeName())
	buf.WriteString(";")

	if charLen != -1 {
		buf.WriteString(fmt.Sprintf("size:%d;", charLen))
	}

	if c.IsPrimaryKey() {
		buf.WriteString("primary_key")
	}

	buf.WriteString("\"")
	return buf.String()
}

func sqlTypeToGoType(mysqlType string, nullable bool, gureguTypes bool) (string, error) {
	mysqlType = strings.Trim(mysqlType, " \t")
	mysqlType = strings.ToLower(mysqlType)

	switch mysqlType {
	case "bit":
		if nullable {
			if gureguTypes {
				return gureguNullInt, nil
			}
			return sqlNullBool, nil
		}
		return golangBool, nil

	case "tinyint", "int", "smallint", "mediumint", "int4", "int2", "integer":
		if nullable {
			if gureguTypes {
				return gureguNullInt, nil
			}
			return sqlNullInt, nil
		}
		return golangInt, nil
	case "bigint", "int8":
		if nullable {
			if gureguTypes {
				return gureguNullInt, nil
			}
			return sqlNullInt, nil
		}
		return golangInt64, nil
	case "char", "enum", "varchar", "longtext", "mediumtext", "text", "tinytext", "varchar2", "json", "jsonb", "nvarchar", "nchar":
		if nullable {
			if gureguTypes {
				return gureguNullString, nil
			}
			return sqlNullString, nil
		}
		return "string", nil
	case "date", "datetime", "time", "timestamp", "smalldatetime":
		if nullable && gureguTypes {
			return gureguNullTime, nil
		}
		return golangTime, nil
	case "decimal", "double", "money", "real":
		if nullable {
			if gureguTypes {
				return gureguNullFloat, nil
			}
			return sqlNullFloat, nil
		}
		return golangFloat64, nil
	case "float":
		if nullable {
			if gureguTypes {
				return gureguNullFloat, nil
			}
			return sqlNullFloat, nil
		}
		return golangFloat32, nil
	case "binary", "blob", "longblob", "mediumblob", "varbinary":
		return golangByteArray, nil
	case "bool":
		return golangBool, nil
	}

	if strings.HasPrefix(mysqlType, "nvarchar") || strings.HasPrefix(mysqlType, "varchar") {
		if nullable {
			if gureguTypes {
				return gureguNullString, nil
			}
			return sqlNullString, nil
		}
		return "string", nil
	}

	if strings.HasPrefix(mysqlType, "numeric") {
		if nullable {
			if gureguTypes {
				return gureguNullFloat, nil
			}
			return sqlNullFloat, nil
		}
		return golangFloat64, nil
	}

	return "", fmt.Errorf("unknown sql type: %s", mysqlType)
}

func BuildDefaultTableDDL(tableName string, cols []*sql.ColumnType) string {
	buf := bytes.Buffer{}
	buf.WriteString("Table: ")
	buf.WriteString(tableName)
	buf.WriteString("\nn")

	for i, ct := range cols {
		buf.WriteString(fmt.Sprintf("[%d] %-20s %s\n", i, ct.Name(), ct.DatabaseTypeName()))
	}
	return buf.String()
}

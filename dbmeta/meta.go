package dbmeta

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/iancoleman/strcase"
)

type metaDataLoader func(db *sql.DB, sqlType, sqlDatabase, tableName string) (DbTableMeta, error)

var metaDataFuncs = make(map[string]metaDataLoader)

var sqlMappings = make(map[string]*SqlMapping)
var UseSqlTypeMappings bool

func init() {
	metaDataFuncs["sqlite3"] = NewSqliteMeta
	metaDataFuncs["sqlite"] = NewSqliteMeta
	metaDataFuncs["mssql"] = NewMsSqlMeta
	metaDataFuncs["postgres"] = NewPostgresMeta
	metaDataFuncs["mysql"] = NewMysqlMeta
}

type SqlMappings struct {
	SqlMappings []*SqlMapping `json:"mappings"`
}

type SqlMapping struct {
	SqlType        string `json:"sql_type"`
	GoType         string `json:"go_type"`
	ProtobufType   string `json:"protobuf_type"`
	GureguType     string `json:"guregu_type"`
	GoNullableType string `json:"go_nullable_type"`
}

type ColumnMeta interface {
	Name() string
	String() string
	Nullable() bool
	DatabaseTypeName() string
	GetGoType(gureguTypes bool) (string, error)
	Index() int
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

type columnMeta struct {
	index           int
	ct              *sql.ColumnType
	nullable        bool
	isPrimaryKey    bool
	isAutoIncrement bool
	colDDL          string
	columnLen       int64
}

// Name returns the name or alias of the column.
func (ci *columnMeta) ColumnLength() int64 {
	return ci.columnLen
}

func (ci *columnMeta) Name() string {
	return ci.ct.Name()
}

func (ci *columnMeta) Index() int {
	return ci.index
}

func (ci *columnMeta) String() string {
	return fmt.Sprintf("[%2d] %-20s nullable: %-6t isPrimaryKey: %-6t isAutoIncrement: %-6t ColumnLength: %d", ci.index, ci.ct.Name(), ci.nullable, ci.isPrimaryKey, ci.isAutoIncrement, ci.columnLen)
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
	addJsonAnnotation bool,
	addGormAnnotation bool,
	addDBAnnotation bool,
	addProtobufAnnotation bool,
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

	fields := generateFieldsTypes(dbMeta, addJsonAnnotation, addGormAnnotation, addDBAnnotation, addProtobufAnnotation, gureguTypes, jsonNameFormat, verbose)

	if verbose {
		fmt.Printf("tableName: %s\n", tableName)
		for _, c := range dbMeta.Columns() {
			fmt.Printf("    %s DatabaseTypeName: %s\n", c.String(), c.DatabaseTypeName())
		}
	}
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
	addJsonAnnotation bool,
	addGormAnnotation bool,
	addDBAnnotation bool,
	addProtobufAnnotation bool,
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
		if addGormAnnotation == true {
			annotations = append(annotations, createGormAnnotation(c))
		}

		if addJsonAnnotation == true {
			annotations = append(annotations, createJsonAnnotation(jsonNameFormat, c))
		}

		if addDBAnnotation == true {
			annotations = append(annotations, createDBAnnotation(c))
		}

		if addProtobufAnnotation == true {
			annnotation, err := createProtobufAnnotation(c)
			if err == nil {
				annotations = append(annotations, annnotation)
			}
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
	buf.WriteString("gorm:\"")

	if c.IsAutoIncrement() {
		buf.WriteString("AUTO_INCREMENT;")
	}

	buf.WriteString("column:")
	buf.WriteString(key)
	buf.WriteString(";")

	buf.WriteString("type:")
	buf.WriteString(c.DatabaseTypeName())
	buf.WriteString(";")

	if c.ColumnLength() > 0 {
		buf.WriteString(fmt.Sprintf("size:%d;", c.ColumnLength()))
	}

	if c.IsPrimaryKey() {
		buf.WriteString("primary_key")
	}

	buf.WriteString("\"")
	return buf.String()
}

func sqlTypeToGoTypeDefault(sqlType string, nullable bool, gureguTypes bool) (string, error) {
	sqlType = strings.Trim(sqlType, " \t")
	sqlType = strings.ToLower(sqlType)

	switch sqlType {
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

	if strings.HasPrefix(sqlType, "nvarchar") || strings.HasPrefix(sqlType, "varchar") {
		if nullable {
			if gureguTypes {
				return gureguNullString, nil
			}
			return sqlNullString, nil
		}
		return "string", nil
	}

	if strings.HasPrefix(sqlType, "numeric") {
		if nullable {
			if gureguTypes {
				return gureguNullFloat, nil
			}
			return sqlNullFloat, nil
		}
		return golangFloat64, nil
	}

	return "", fmt.Errorf("unknown sql type: %s", sqlType)
}

func BuildDefaultTableDDL(tableName string, cols []*sql.ColumnType) string {
	buf := bytes.Buffer{}
	buf.WriteString(fmt.Sprintf("Table: %s\n", tableName))

	for i, ct := range cols {
		buf.WriteString(fmt.Sprintf("[%d] %-20s %s\n", i, ct.Name(), ct.DatabaseTypeName()))
	}
	return buf.String()
}

func createDBAnnotation(c ColumnMeta) string {
	return fmt.Sprintf("db:\"%s\"", c.Name())
}

func createProtobufAnnotation(c ColumnMeta) (string, error) {
	protoBufType, err := sqlTypeToProtobufType(c)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("protobuf:\"%s,%d,opt,name=%s\"", protoBufType, c.Index(), c.Name()), nil
}

func sqlTypeToProtobufTypeDefault(c ColumnMeta) (string, error) {
	sqlType := strings.Trim(c.DatabaseTypeName(), " \t")
	sqlType = strings.ToLower(sqlType)

	switch sqlType {
	case "bit":
		return "bool", nil
	case "tinyint":
		return "uint8", nil
	case "smallint":
		return "int16", nil
	case "int":
		return "int32", nil
	case "bigint":
		return "int64", nil
	case "mediumint", "int4", "int2", "integer":
		return "int32", nil
	case "int8":
		return "int8", nil
	case "char", "enum", "varchar", "longtext", "mediumtext", "text", "tinytext", "varchar2", "json", "jsonb", "nvarchar", "nchar":
		return "string", nil
	case "date", "datetime", "time", "timestamp", "smalldatetime":
		return "uint64", nil
	case "decimal", "double", "money", "real":
		return "float", nil
	case "float":
		return "float", nil
	case "binary", "blob", "longblob", "mediumblob", "varbinary":
		return "bytes", nil
	case "bool":
		return "bool", nil
	}

	if strings.HasPrefix(sqlType, "nvarchar") || strings.HasPrefix(sqlType, "varchar") {
		return "string", nil
	}

	if strings.HasPrefix(sqlType, "numeric") {
		return "float", nil
	}

	return "", fmt.Errorf("unknown sql type: %s", sqlType)
}

func ProcessMappings(mappingJsonstring []byte) error {
	var mappings = &SqlMappings{}
	err := json.Unmarshal(mappingJsonstring, mappings)
	if err != nil {
		fmt.Printf("Error unmarshalling json error: %v\n", err)
		return err
	}

	fmt.Printf("Loaded %d mappings\n", len(mappings.SqlMappings))
	for i, value := range mappings.SqlMappings {
		fmt.Printf("    Mapping:[%2d] -> %s\n", i, value.SqlType)
		sqlMappings[value.SqlType] = value
	}
	return nil
}

func LoadMappings(mappingFileName string) error {
	mappingFile, err := os.Open(mappingFileName)
	defer mappingFile.Close()
	byteValue, err := ioutil.ReadAll(mappingFile)
	if err != nil {
		fmt.Printf("Error loading mapping file %s error: %v\n", mappingFileName, err)
		return err
	}
	return ProcessMappings(byteValue)
}

func sqlTypeToGoType(sqlType string, nullable bool, gureguTypes bool) (string, error) {
	sqlType = strings.Trim(sqlType, " \t")
	sqlType = strings.ToLower(sqlType)

	if UseSqlTypeMappings {
		mapping, ok := sqlMappings[sqlType]
		if !ok {
			return "", fmt.Errorf("unknown sql type: %s", sqlType)
		}

		if nullable && gureguTypes {
			return mapping.GureguType, nil
		} else if nullable {
			return mapping.GoNullableType, nil
		} else {
			return mapping.GoType, nil
		}

	} else {
		return sqlTypeToGoTypeDefault(sqlType, nullable, gureguTypes)
	}
}

func sqlTypeToProtobufType(c ColumnMeta) (string, error) {
	sqlType := strings.ToLower(c.DatabaseTypeName())
	sqlType = strings.Trim(sqlType, " \t")
	sqlType = strings.ToLower(sqlType)

	if UseSqlTypeMappings {
		mapping, ok := sqlMappings[sqlType]
		if !ok {
			return "", fmt.Errorf("unknown sql type: %s", sqlType)
		}
		return mapping.ProtobufType, nil

	} else {
		return sqlTypeToProtobufTypeDefault(c)
	}
}

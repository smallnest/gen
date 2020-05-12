package dbmeta

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/bxcodec/faker/v3"
	"github.com/iancoleman/strcase"
	"github.com/ompluscator/dynamic-struct"
)

type metaDataLoader func(db *sql.DB, sqlType, sqlDatabase, tableName string) (DbTableMeta, error)

var metaDataFuncs = make(map[string]metaDataLoader)
var sqlMappings = make(map[string]*SqlMapping)

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
	JsonType       string `json:"json_type"`
	ProtobufType   string `json:"protobuf_type"`
	GureguType     string `json:"guregu_type"`
	GoNullableType string `json:"go_nullable_type"`
}

type ColumnMeta interface {
	Name() string
	String() string
	Nullable() bool
	DatabaseTypeName() string
	DatabaseTypePretty() string
	Index() int
	IsPrimaryKey() bool
	IsAutoIncrement() bool
	ColumnType() string
	ColumnLength() int64
	DefaultValue() string
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
	columnType      string
	columnLen       int64
	defaultVal      string
}

func (ci *columnMeta) ColumnType() string {
	return ci.columnType
}

func (ci *columnMeta) ColumnLength() int64 {
	return ci.columnLen
}
func (ci *columnMeta) DefaultValue() string {
	return ci.defaultVal
}



func (ci *columnMeta) Name() string {
	return ci.ct.Name()
}

func (ci *columnMeta) Index() int {
	return ci.index
}

func (ci *columnMeta) String() string {
	return fmt.Sprintf("[%2d] %-45s  %-20s null: %-6t primary: %-6t auto: %-6t col: %-15s len: %-7d default: [%s]",
		ci.index, ci.ct.Name(), ci.DatabaseTypePretty(),
		ci.nullable, ci.isPrimaryKey,
		ci.isAutoIncrement, ci.columnType, ci.columnLen, ci.defaultVal)
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

func (ci *columnMeta) DatabaseTypePretty() string {
	if ci.columnLen > 0 {
		return fmt.Sprintf("%s(%d)", ci.columnType, ci.columnLen)
	} else {
		return ci.columnType
	}
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
	Index                 int
	IndexPlus1            int
	PackageName           string
	StructName            string
	ShortStructName       string
	TableName             string
	Fields                []string
	DBMeta                DbTableMeta
	Instance              interface{}
	CodeFields            []*FieldInfo
	PrimaryKeyField       int
	PrimaryKeyGoType      string
	PrimaryKeyFieldParser string
}

type FieldInfo struct {
	Index             int
	GoFieldName       string
	GoFieldType       string
	GoAnnotations     []string
	JsonFieldName     string
	ProtobufFieldName string
	ProtobufType      string
	ProtobufPos       int
	Comment           string
	Code              string
	FakeData          interface{}
	ColumnMeta        ColumnMeta
}

// GenerateStruct generates a struct for the given table.
func LoadMeta(sqlType string, db *sql.DB, sqlDatabase, tableName string, ) (DbTableMeta, error) {
	dbMetaFunc, haveMeta := metaDataFuncs[sqlType]
	if !haveMeta {
		dbMetaFunc = NewUnknownMeta
	}

	dbMeta, err := dbMetaFunc(db, sqlType, sqlDatabase, tableName)
	return dbMeta, err
}

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
	protobufNameFormat string,
	verbose bool) (*ModelInfo, error) {

	jsonNameFormat = strings.ToLower(jsonNameFormat)

	dbMeta, err := LoadMeta(sqlType, db, sqlDatabase, tableName)
	if err != nil {
		return nil, err
	}

	fields := generateFieldsTypes(dbMeta, addJsonAnnotation, addGormAnnotation, addDBAnnotation, addProtobufAnnotation, gureguTypes, jsonNameFormat, protobufNameFormat, verbose)

	if verbose {
		fmt.Printf("tableName: %s\n", tableName)
		for _, c := range dbMeta.Columns() {
			fmt.Printf("    %s\n", c.String())
		}
	}

	generator := dynamicstruct.NewStruct()
	keyField := -1
	for i, c := range fields {
		meta := dbMeta.Columns()[i]
		jsonName := formatFieldName(jsonNameFormat, meta)
		tag := fmt.Sprintf(`json:"%s"`, jsonName)
		fakeData := c.FakeData
		generator = generator.AddField(c.GoFieldName, fakeData, tag)
		if keyField == -1 && meta.IsPrimaryKey() {
			keyField = i
		}
	}

	instance := generator.Build().New()

	err = faker.FakeData(&instance)
	if err != nil {
		fmt.Println(err)
	}
	// fmt.Printf("%+v", instance)

	var code []string
	for _, f := range fields {
		code = append(code, f.Code)
	}

	primaryKeyFieldType := "interface{}"
	if keyField != -1 {
		primaryKeyFieldType = fields[keyField].GoFieldType
	}
	primaryKeyFieldParser := "parseString"
	switch primaryKeyFieldType {
	case "interface{}":
		primaryKeyFieldParser = "parseInterface"
	case "string":
		primaryKeyFieldParser = "parseString"
	case "int":
		primaryKeyFieldParser = "parseInt"

	case "int32":
		primaryKeyFieldParser = "parseInt32"
	case "int64":
		primaryKeyFieldParser = "parseInt64"

	}

	var modelInfo = &ModelInfo{
		PackageName:           pkgName,
		StructName:            structName,
		TableName:             tableName,
		ShortStructName:       strings.ToLower(string(structName[0])),
		Fields:                code,
		CodeFields:            fields,
		DBMeta:                dbMeta,
		Instance:              instance,
		PrimaryKeyField:       keyField,
		PrimaryKeyGoType:      primaryKeyFieldType,
		PrimaryKeyFieldParser: primaryKeyFieldParser,
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
	protobufNameFormat string,
	verbose bool) []*FieldInfo {

	var fields []*FieldInfo
	var field = ""
	for i, c := range dbMeta.Columns() {
		name := c.Name()
		if verbose {
			//fmt.Printf("   [%s]   [%d] field: %s type: %s\n", dbMeta.TableName(), i, key, c.DatabaseTypeName())
		}

		valueType, err := SqlTypeToGoType(strings.ToLower(c.DatabaseTypeName()), c.Nullable(), gureguTypes)
		if err != nil { // unknown type
			fmt.Printf("table: %s unable to generate struct field: %s type: %s error: %v\n", dbMeta.TableName(), name, c.DatabaseTypeName(), err)
			continue
		}
		fieldName := FmtFieldName(stringifyFirstChar(name))

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
			annnotation, err := createProtobufAnnotation(protobufNameFormat, c)
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
			field = fmt.Sprintf("%s %s", fieldName, valueType)
		}

		field = fmt.Sprintf("%s //%s", field, c.String())

		goType, _ := SqlTypeToGoType(strings.ToLower(c.DatabaseTypeName()), false, false)
		protobufType, _ := SqlTypeToProtobufType(c.DatabaseTypeName())
		fakeData := createFakeData(goType, fieldName)

		fi := &FieldInfo{
			Index:             i,
			Code:              field,
			GoFieldName:       fieldName,
			GoFieldType:       valueType,
			GoAnnotations:     annotations,
			FakeData:          fakeData,
			Comment:           c.String(),
			JsonFieldName:     formatFieldName(jsonNameFormat, c),
			ProtobufFieldName: formatFieldName(protobufNameFormat, c),
			ProtobufType:      protobufType,
			ProtobufPos:       i + 1,
			ColumnMeta:        c,
		}

		fields = append(fields, fi)
	}
	return fields
}

func formatFieldName(nameFormat string, c ColumnMeta) string {

	var jsonName string
	switch nameFormat {
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
	return jsonName
}

func createJsonAnnotation(nameFormat string, c ColumnMeta) string {

	name := formatFieldName(nameFormat, c)
	return fmt.Sprintf("json:\"%s\"", name)
}

func createDBAnnotation(c ColumnMeta) string {
	return fmt.Sprintf("db:\"%s\"", c.Name())
}

func createProtobufAnnotation(nameFormat string, c ColumnMeta) (string, error) {
	protoBufType, err := SqlTypeToProtobufType(c.DatabaseTypeName())
	if err != nil {
		return "", err
	}

	if protoBufType != "" {
		name := formatFieldName(nameFormat, c)
		return fmt.Sprintf("protobuf:\"%s,%d,opt,name=%s\"", protoBufType, c.Index(), name), nil
	} else {
		return "", fmt.Errorf("unknown sql name: %s", c.Name())
	}
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

	if c.DatabaseTypeName() != "" {
		buf.WriteString("type:")
		buf.WriteString(c.DatabaseTypeName())
		buf.WriteString(";")

		if c.ColumnLength() > 0 {
			buf.WriteString(fmt.Sprintf("size:%d;", c.ColumnLength()))
		}

		if c.DefaultValue() != "" {
			value := c.DefaultValue()
			value = strings.Replace(value, "\"", "'", -1)

			if value == "NULL" ||value == "null" {
				value = ""
			}

			if value != "" && strings.Index(value, "()") ==  -1 {
				buf.WriteString(fmt.Sprintf("default:%s;", value ))
			}
		}


		if c.IsPrimaryKey() {
			buf.WriteString("primary_key")
		}
	}

	buf.WriteString("\"")
	return buf.String()
}

func BuildDefaultTableDDL(tableName string, cols []ColumnMeta) string {
	buf := bytes.Buffer{}
	buf.WriteString(fmt.Sprintf("Table: %s\n", tableName))

	for _, ct := range cols {
		buf.WriteString(fmt.Sprintf("%s\n", ct.String()))
	}
	return buf.String()
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

func SqlTypeToGoType(sqlType string, nullable bool, gureguTypes bool) (string, error) {
	mapping, err := SqlTypeToMapping(sqlType)
	if err != nil {
		return "", err
	}

	if nullable && gureguTypes {
		return mapping.GureguType, nil
	} else if nullable {
		return mapping.GoNullableType, nil
	} else {
		return mapping.GoType, nil
	}
}

func SqlTypeToProtobufType(sqlType string) (string, error) {
	mapping, err := SqlTypeToMapping(sqlType)
	if err != nil {
		return "", err
	}
	return mapping.ProtobufType, nil
}

func SqlTypeToMapping(sqlType string) (*SqlMapping, error) {
	sqlType = cleanupSqlType(sqlType)

	mapping, ok := sqlMappings[sqlType]
	if !ok {
		return nil, fmt.Errorf("unknown sql type: %s", sqlType)
	}
	return mapping, nil
}

func cleanupSqlType(sqlType string) string {
	sqlType = strings.ToLower(sqlType)
	sqlType = strings.Trim(sqlType, " \t")
	sqlType = strings.ToLower(sqlType)
	idx := strings.Index(sqlType, "(")
	if idx > -1 {
		sqlType = sqlType[0:idx]
	}
	return sqlType
}

func GetMappings() map[string]*SqlMapping {
	return sqlMappings
}

func createFakeData(valueType string, name string) interface{} {

	switch valueType {
	case "[]byte":
		return []byte("hello world")
	case "bool":
		return true
	case "float32":
		return float32(1.0)
	case "float64":
		return float64(1.0)
	case "int":
		return int(1)
	case "int64":
		return int64(1)
	case "string":
		return "hello world"
	case "time.Time":
		return time.Now()
	case "interface{}":
		return 1
	default:
		return 1
	}

}

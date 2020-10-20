package dbmeta

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/bxcodec/faker/v3"
	"github.com/iancoleman/strcase"
	dynamicstruct "github.com/ompluscator/dynamic-struct"
)

type metaDataLoader func(db *sql.DB, sqlType, sqlDatabase, tableName string) (DbTableMeta, error)

var metaDataFuncs = make(map[string]metaDataLoader)
var sqlMappings = make(map[string]*SQLMapping)

func init() {
	metaDataFuncs["sqlite3"] = LoadSqliteMeta
	metaDataFuncs["sqlite"] = LoadSqliteMeta
	metaDataFuncs["mssql"] = LoadMsSQLMeta
	metaDataFuncs["postgres"] = LoadPostgresMeta
	metaDataFuncs["mysql"] = LoadMysqlMeta
}

// SQLMappings mappings for sql types to json, go etc
type SQLMappings struct {
	SQLMappings []*SQLMapping `json:"mappings"`
}

// SQLMapping mapping
type SQLMapping struct {
	// SQLType sql type reported from db
	SQLType string `json:"sql_type"`

	// GoType mapped go type
	GoType string `json:"go_type"`

	// JSONType mapped json type
	JSONType string `json:"json_type"`

	// ProtobufType mapped protobuf type
	ProtobufType string `json:"protobuf_type"`

	// GureguType mapped go type using Guregu
	GureguType string `json:"guregu_type"`

	// GoNullableType mapped go type using nullable
	GoNullableType string `json:"go_nullable_type"`

	// SwaggerType mapped type
	SwaggerType string `json:"swagger_type"`
}

func (m *SQLMapping) String() interface{} {
	return fmt.Sprintf("SQLType: %-15s  GoType: %-15s GureguType: %-15s GoNullableType: %-15s JSONType: %-15s ProtobufType: %-15s",
		m.SQLType,
		m.GoType, m.GureguType, m.GoNullableType,
		m.JSONType, m.ProtobufType)
}

// IsAutoIncrement return is column is a primary key column
func (ci *columnMeta) IsPrimaryKey() bool {
	return ci.isPrimaryKey
}

// IsArray return is column is an array type
func (ci *columnMeta) IsArray() bool {
	return ci.isArray
}

// IsAutoIncrement return is column is an auto increment column
func (ci *columnMeta) IsAutoIncrement() bool {
	return ci.isAutoIncrement
}

type columnMeta struct {
	index int
	// ct              *sql.ColumnType
	nullable         bool
	isPrimaryKey     bool
	isAutoIncrement  bool
	isArray          bool
	colDDL           string
	columnType       string
	columnLen        int64
	defaultVal       string
	notes            string
	comment          string
	databaseTypeName string
	name             string
}

// ColumnType column type
func (ci *columnMeta) ColumnType() string {
	return ci.columnType
}

// Notes notes on column generation
func (ci *columnMeta) Notes() string {
	return ci.notes
}

// Comment column comment
func (ci *columnMeta) Comment() string {
	return ci.comment
}

// ColumnLength column length for text or varhar
func (ci *columnMeta) ColumnLength() int64 {
	return ci.columnLen
}

// DefaultValue default value of column
func (ci *columnMeta) DefaultValue() string {
	return ci.defaultVal
}

// Name name of column
func (ci *columnMeta) Name() string {
	return ci.name
}

// Index index of column in db
func (ci *columnMeta) Index() int {
	return ci.index
}

// String friendly string for columnMeta
func (ci *columnMeta) String() string {
	return fmt.Sprintf("[%2d] %-45s  %-20s null: %-6t primary: %-6t isArray: %-6t auto: %-6t col: %-15s len: %-7d default: [%s]",
		ci.index, ci.name, ci.DatabaseTypePretty(),
		ci.nullable, ci.isPrimaryKey, ci.isArray,
		ci.isAutoIncrement, ci.columnType, ci.columnLen, ci.defaultVal)
}

// Nullable reports whether the column may be null.
// If a driver does not support this property ok will be false.
func (ci *columnMeta) Nullable() bool {
	return ci.nullable
}

// ColDDL string of the ddl for the column
func (ci *columnMeta) ColDDL() string {
	return ci.colDDL
}

// DatabaseTypeName returns the database system name of the column type. If an empty
// string is returned the driver type name is not supported.
// Consult your driver documentation for a list of driver data types. Length specifiers
// are not included.
// Common type include "VARCHAR", "TEXT", "NVARCHAR", "DECIMAL", "BOOL", "INT", "BIGINT".
func (ci *columnMeta) DatabaseTypeName() string {
	return ci.databaseTypeName
}

// DatabaseTypePretty string of the db type
func (ci *columnMeta) DatabaseTypePretty() string {
	if ci.columnLen > 0 {
		return fmt.Sprintf("%s(%d)", ci.columnType, ci.columnLen)
	}

	return ci.columnType
}

// DbTableMeta table meta data
type DbTableMeta interface {
	Columns() []ColumnMeta
	SQLType() string
	SQLDatabase() string
	TableName() string
	DDL() string
}

// ColumnMeta meta data for a column
type ColumnMeta interface {
	Name() string
	String() string
	Nullable() bool
	DatabaseTypeName() string
	DatabaseTypePretty() string
	Index() int
	IsPrimaryKey() bool
	IsAutoIncrement() bool
	IsArray() bool
	ColumnType() string
	Notes() string
	Comment() string
	ColumnLength() int64
	DefaultValue() string
}

type dbTableMeta struct {
	sqlType       string
	sqlDatabase   string
	tableName     string
	columns       []*columnMeta
	ddl           string
	primaryKeyPos int
}

// PrimaryKeyPos ordinal pos of primary key
func (m *dbTableMeta) PrimaryKeyPos() int {
	return m.primaryKeyPos
}

// SQLType sql db type
func (m *dbTableMeta) SQLType() string {
	return m.sqlType
}

// SQLDatabase sql database name
func (m *dbTableMeta) SQLDatabase() string {
	return m.sqlDatabase
}

// TableName sql table name
func (m *dbTableMeta) TableName() string {
	return m.tableName
}

// Columns ColumnMeta for columns in a sql table
func (m *dbTableMeta) Columns() []ColumnMeta {

	cols := make([]ColumnMeta, len(m.columns))
	for i, v := range m.columns {
		cols[i] = ColumnMeta(v)
	}
	return cols
}

// DDL string for a sql table
func (m *dbTableMeta) DDL() string {
	return m.ddl
}

// ModelInfo info for a sql table
type ModelInfo struct {
	Index           int
	IndexPlus1      int
	PackageName     string
	StructName      string
	ShortStructName string
	TableName       string
	Fields          []string
	DBMeta          DbTableMeta
	Instance        interface{}
	CodeFields      []*FieldInfo
}

// Notes notes on table generation
func (m *ModelInfo) Notes() string {
	buf := bytes.Buffer{}

	for i, j := range m.DBMeta.Columns() {
		if j.Notes() != "" {
			buf.WriteString(fmt.Sprintf("[%2d] %s\n", i, j.Notes()))
		}
	}

	for i, j := range m.CodeFields {
		if j.Notes != "" {
			buf.WriteString(fmt.Sprintf("[%2d] %s\n", i, j.Notes))
		}
	}

	return buf.String()
}

// FieldInfo codegen info for each column in sql table
type FieldInfo struct {
	Index                 int
	GoFieldName           string
	GoFieldType           string
	GoAnnotations         []string
	JSONFieldName         string
	ProtobufFieldName     string
	ProtobufType          string
	ProtobufPos           int
	Comment               string
	Notes                 string
	Code                  string
	FakeData              interface{}
	ColumnMeta            ColumnMeta
	PrimaryKeyFieldParser string
	PrimaryKeyArgName     string
	SQLMapping            *SQLMapping
	GormAnnotation        string
	JSONAnnotation        string
	XMLAnnotation         string
	DBAnnotation          string
	GoGoMoreTags          string
}

// GetFunctionName get function name
func GetFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

// LoadMeta loads the DbTableMeta data from the db connection for the table
func LoadMeta(sqlType string, db *sql.DB, sqlDatabase, tableName string) (DbTableMeta, error) {
	dbMetaFunc, haveMeta := metaDataFuncs[sqlType]
	if !haveMeta {
		dbMetaFunc = LoadUnknownMeta
	}

	dbMeta, err := dbMetaFunc(db, sqlType, sqlDatabase, tableName)
	//if err != nil {
	//	fmt.Printf("Error calling func: %s error: %v\n", GetFunctionName(dbMetaFunc), err)
	//}
	return dbMeta, err
}

// GenerateFieldsTypes FieldInfo slice from DbTableMeta
func (c *Config) GenerateFieldsTypes(dbMeta DbTableMeta) ([]*FieldInfo, error) {

	var fields []*FieldInfo
	field := ""
	for i, col := range dbMeta.Columns() {
		fieldName := col.Name()

		fi := &FieldInfo{
			Index: i,
		}

		valueType, err := SQLTypeToGoType(strings.ToLower(col.DatabaseTypeName()), col.Nullable(), c.UseGureguTypes)
		if err != nil { // unknown type
			fmt.Printf("table: %s unable to generate struct field: %s type: %s error: %v\n", dbMeta.TableName(), fieldName, col.DatabaseTypeName(), err)
			continue
		}

		fieldName = Replace(c.FieldNamingTemplate, fieldName)
		fieldName = checkDupeFieldName(fields, fieldName)

		fi.GormAnnotation = createGormAnnotation(col)
		fi.JSONAnnotation = createJSONAnnotation(c.JSONNameFormat, col)
		fi.XMLAnnotation = createXMLAnnotation(c.XMLNameFormat, col)
		fi.DBAnnotation = createDBAnnotation(col)

		var annotations []string
		if c.AddGormAnnotation {
			annotations = append(annotations, fi.GormAnnotation)
		}

		if c.AddJSONAnnotation {
			annotations = append(annotations, fi.JSONAnnotation)
		}

		if c.AddXMLAnnotation {
			annotations = append(annotations, fi.XMLAnnotation)
		}

		if c.AddDBAnnotation {
			annotations = append(annotations, fi.DBAnnotation)
		}

		gogoTags := []string{fi.GormAnnotation, fi.JSONAnnotation, fi.XMLAnnotation, fi.DBAnnotation}
		GoGoMoreTags := strings.Join(gogoTags, " ")

		if c.AddProtobufAnnotation {
			annotation, err := createProtobufAnnotation(c.ProtobufNameFormat, col)
			if err == nil {
				annotations = append(annotations, annotation)
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

		field = fmt.Sprintf("//%s\n    %s", col.String(), field)
		if col.Comment() != "" {
			field = fmt.Sprintf("%s // %s", field, col.Comment())
		}

		sqlMapping, _ := SQLTypeToMapping(strings.ToLower(col.DatabaseTypeName()))
		goType, _ := SQLTypeToGoType(strings.ToLower(col.DatabaseTypeName()), false, false)
		protobufType, _ := SQLTypeToProtobufType(col.DatabaseTypeName())

		// fmt.Printf("protobufType: %v  DatabaseTypeName: %v\n", protobufType, col.DatabaseTypeName())

		fakeData := createFakeData(goType, fieldName)

		//if c.Verbose {
		//	fmt.Printf("table: %-10s type: %-10s fieldname: %-20s val: %v\n", c.DatabaseTypeName(), goType, fieldName, fakeData)
		//	spew.Dump(fakeData)
		//}

		//fmt.Printf("%+v", fakeData)
		primaryKeyFieldParser := ""
		if col.IsPrimaryKey() {
			var ok bool
			primaryKeyFieldParser, ok = parsePrimaryKeys[goType]
			if !ok {
				primaryKeyFieldParser = "unsupported"
			}
		}

		fi.Code = field
		fi.GoFieldName = fieldName
		fi.GoFieldType = valueType
		fi.GoAnnotations = annotations
		fi.FakeData = fakeData
		fi.Comment = col.String()
		fi.JSONFieldName = formatFieldName(c.JSONNameFormat, col.Name())
		fi.ProtobufFieldName = formatFieldName(c.ProtobufNameFormat, col.Name())
		fi.ProtobufType = protobufType
		fi.ProtobufPos = i + 1
		fi.ColumnMeta = col
		fi.PrimaryKeyFieldParser = primaryKeyFieldParser
		fi.SQLMapping = sqlMapping
		fi.GoGoMoreTags = GoGoMoreTags

		fi.JSONFieldName = checkDupeJSONFieldName(fields, fi.JSONFieldName)
		fi.ProtobufFieldName = checkDupeProtoBufFieldName(fields, fi.ProtobufFieldName)

		fields = append(fields, fi)
	}
	return fields, nil
}

func formatFieldName(nameFormat string, name string) string {

	var jsonName string
	switch nameFormat {
	case "snake":
		jsonName = strcase.ToSnake(name)
	case "camel":
		jsonName = strcase.ToCamel(name)
	case "lower_camel":
		jsonName = strcase.ToLowerCamel(name)
	case "none":
		jsonName = name
	default:
		jsonName = name
	}
	return jsonName
}

func createJSONAnnotation(nameFormat string, c ColumnMeta) string {
	name := formatFieldName(nameFormat, c.Name())
	return fmt.Sprintf("json:\"%s\"", name)
}

func createXMLAnnotation(nameFormat string, c ColumnMeta) string {
	name := formatFieldName(nameFormat, c.Name())
	return fmt.Sprintf("xml:\"%s\"", name)
}

func createDBAnnotation(c ColumnMeta) string {
	return fmt.Sprintf("db:\"%s\"", c.Name())
}

func createProtobufAnnotation(nameFormat string, c ColumnMeta) (string, error) {
	protoBufType, err := SQLTypeToProtobufType(c.DatabaseTypeName())
	if err != nil {
		return "", err
	}

	if protoBufType != "" {
		name := formatFieldName(nameFormat, c.Name())
		return fmt.Sprintf("protobuf:\"%s,%d,opt,name=%s\"", protoBufType, c.Index(), name), nil
	}

	return "", fmt.Errorf("unknown sql name: %s", c.Name())
}

func createGormAnnotation(c ColumnMeta) string {
	buf := bytes.Buffer{}

	key := c.Name()
	buf.WriteString("gorm:\"")

	if c.IsPrimaryKey() {
		buf.WriteString("primary_key;")
	}
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

			if value == "NULL" || value == "null" {
				value = ""
			}

			if value != "" && !strings.Contains(value, "()") {
				buf.WriteString(fmt.Sprintf("default:%s;", value))
			}
		}

	}

	buf.WriteString("\"")
	return buf.String()
}

// BuildDefaultTableDDL create a ddl mock using the ColumnMeta data
func BuildDefaultTableDDL(tableName string, cols []*columnMeta) string {
	buf := bytes.Buffer{}
	buf.WriteString(fmt.Sprintf("Table: %s\n", tableName))

	for _, ct := range cols {
		buf.WriteString(fmt.Sprintf("%s\n", ct.String()))
	}
	return buf.String()
}

// ProcessMappings process the json for mappings to load sql mappings
func ProcessMappings(source string, mappingJsonstring []byte, verbose bool) error {
	var mappings = &SQLMappings{}
	err := json.Unmarshal(mappingJsonstring, mappings)
	if err != nil {
		fmt.Printf("Error unmarshalling json error: %v\n", err)
		return err
	}

	if verbose {
		fmt.Printf("Loaded %d mappings from: %s\n", len(mappings.SQLMappings), source)
	}
	for i, value := range mappings.SQLMappings {
		if verbose {
			fmt.Printf("    Mapping:[%2d] -> %s\n", i, value.SQLType)
		}

		sqlMappings[value.SQLType] = value
	}

	return nil
}

// LoadMappings load sql mappings to load mapping json file
func LoadMappings(mappingFileName string, verbose bool) error {
	mappingFile, err := os.Open(mappingFileName)
	if err != nil {
		fmt.Printf("Error loading mapping file %s error: %v\n", mappingFileName, err)
		return err
	}

	defer func() {
		_ = mappingFile.Close()
	}()
	byteValue, err := ioutil.ReadAll(mappingFile)
	if err != nil {
		fmt.Printf("Error loading mapping file %s error: %v\n", mappingFileName, err)
		return err
	}

	absPath, err := filepath.Abs(mappingFileName)
	if err != nil {
		absPath = mappingFileName
	}

	return ProcessMappings(absPath, byteValue, verbose)
}

// SQLTypeToGoType map a sql type to a go type
func SQLTypeToGoType(sqlType string, nullable bool, gureguTypes bool) (string, error) {
	mapping, err := SQLTypeToMapping(sqlType)
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

// SQLTypeToProtobufType map a sql type to a protobuf type
func SQLTypeToProtobufType(sqlType string) (string, error) {
	mapping, err := SQLTypeToMapping(sqlType)
	if err != nil {
		return "", err
	}
	return mapping.ProtobufType, nil
}

// SQLTypeToMapping retrieve a SQLMapping based on a sql type
func SQLTypeToMapping(sqlType string) (*SQLMapping, error) {
	sqlType = cleanupSQLType(sqlType)

	mapping, ok := sqlMappings[sqlType]
	if !ok {
		return nil, fmt.Errorf("unknown sql type: %s", sqlType)
	}
	return mapping, nil
}

func cleanupSQLType(sqlType string) string {
	sqlType = strings.ToLower(sqlType)
	sqlType = strings.Trim(sqlType, " \t")
	sqlType = strings.ToLower(sqlType)
	idx := strings.Index(sqlType, "(")
	if idx > -1 {
		sqlType = sqlType[0:idx]
	}
	return sqlType
}

// GetMappings get all mappings
func GetMappings() map[string]*SQLMapping {
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

// FindInSlice takes a slice and looks for an element in it. If found it will
// return it's key, otherwise it will return -1 and a bool of false.
func FindInSlice(slice []string, val string) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}

// LoadTableInfo load table info from db connection, and list of tables
func LoadTableInfo(db *sql.DB, dbTables []string, excludeDbTables []string, conf *Config) map[string]*ModelInfo {

	tableInfos := make(map[string]*ModelInfo)

	// generate go files for each table
	var tableIdx = 0
	for i, tableName := range dbTables {

		_, ok := FindInSlice(excludeDbTables, tableName)
		if ok {
			fmt.Printf("Skipping excluded table %s\n", tableName)
			continue
		}

		if strings.HasPrefix(tableName, "[") && strings.HasSuffix(tableName, "]") {
			tableName = tableName[1 : len(tableName)-1]
		}

		dbMeta, err := LoadMeta(conf.SQLType, db, conf.SQLDatabase, tableName)
		if err != nil {
			msg := fmt.Sprintf("Warning - LoadMeta skipping table info for %s error: %v\n", tableName, err)
			if au != nil {
				fmt.Print(au.Yellow(msg))
			} else {
				fmt.Printf(msg)
			}

			continue
		}

		modelInfo, err := GenerateModelInfo(tableInfos, dbMeta, tableName, conf)
		if err != nil {
			msg := fmt.Sprintf("Error - %v\n", err)
			if au != nil {
				fmt.Print(au.Red(msg))
			} else {
				fmt.Printf(msg)
			}

			continue
		}

		if len(modelInfo.Fields) == 0 {
			if conf.Verbose {
				fmt.Printf("[%d] Table: %s - No Fields Available\n", i, tableName)
			}
			continue
		}

		modelInfo.Index = tableIdx
		modelInfo.IndexPlus1 = tableIdx + 1
		tableIdx++

		tableInfos[tableName] = modelInfo
	}

	return tableInfos
}

// GenerateModelInfo generates a struct for the given table.
func GenerateModelInfo(tables map[string]*ModelInfo, dbMeta DbTableMeta,
	tableName string,
	conf *Config) (*ModelInfo, error) {

	structName := Replace(conf.ModelNamingTemplate, tableName)
	structName = CheckForDupeTable(tables, structName)

	fields, err := conf.GenerateFieldsTypes(dbMeta)
	if err != nil {
		return nil, err
	}

	if conf.Verbose {
		fmt.Printf("\ntableName: %s\n", tableName)
		for _, c := range dbMeta.Columns() {
			fmt.Printf("    %s\n", c.String())
		}
		fmt.Print("\n")
	}

	generator := dynamicstruct.NewStruct()

	noOfPrimaryKeys := 0
	for i, c := range fields {
		meta := dbMeta.Columns()[i]
		jsonName := formatFieldName(conf.JSONNameFormat, meta.Name())
		tag := fmt.Sprintf(`json:"%s"`, jsonName)
		fakeData := c.FakeData
		generator = generator.AddField(c.GoFieldName, fakeData, tag)
		if meta.IsPrimaryKey() {
			//c.PrimaryKeyArgName = RenameReservedName(strcase.ToLowerCamel(c.GoFieldName))
			c.PrimaryKeyArgName = fmt.Sprintf("arg%s", FmtFieldName(c.GoFieldName))
			noOfPrimaryKeys++
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

		if f.PrimaryKeyFieldParser == "unsupported" {
			return nil, fmt.Errorf("unable to generate code for table: %s, primary key column: [%d] %s has unsupported type: %s / %s",
				dbMeta.TableName(), f.ColumnMeta.Index(), f.ColumnMeta.Name(), f.ColumnMeta.DatabaseTypeName(), f.GoFieldType)
		}
		code = append(code, f.Code)
	}

	var modelInfo = &ModelInfo{
		PackageName:     conf.ModelPackageName,
		StructName:      structName,
		TableName:       tableName,
		ShortStructName: strings.ToLower(string(structName[0])),
		Fields:          code,
		CodeFields:      fields,
		DBMeta:          dbMeta,
		Instance:        instance,
	}

	return modelInfo, nil
}

// CheckForDupeTable check for duplicate table name, returns available name
func CheckForDupeTable(tables map[string]*ModelInfo, name string) string {
	found := false

	for _, model := range tables {
		if model.StructName == name {
			found = true
		}
	}
	if found {
		name = CheckForDupeTable(tables, name+"_")
	}

	if name == "Result" {
		name = "DBTableResult"
	}

	return name
}

func checkDupeFieldName(fields []*FieldInfo, fieldName string) string {
	var match bool
	for _, field := range fields {
		if fieldName == field.GoFieldName {
			match = true
			break
		}
	}

	if match {
		fieldName = checkDupeFieldName(fields, generateAlternativeName(fieldName))
	}

	return fieldName
}

func checkDupeJSONFieldName(fields []*FieldInfo, fieldName string) string {
	var match bool
	for _, field := range fields {
		if fieldName == field.JSONFieldName {
			match = true
			break
		}
	}

	if match {
		fieldName = checkDupeJSONFieldName(fields, generateAlternativeName(fieldName))
	}

	return fieldName
}

func checkDupeProtoBufFieldName(fields []*FieldInfo, fieldName string) string {
	var match bool
	for _, field := range fields {
		if fieldName == field.ProtobufFieldName {
			match = true
			break
		}
	}

	if match {
		fieldName = checkDupeProtoBufFieldName(fields, generateAlternativeName(fieldName))
	}

	return fieldName
}

// @TODO In progress - need more elegant renaming
func generateAlternativeName(name string) string {
	name = name + "alt1"
	return name
}

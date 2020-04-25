package dbmeta

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/jimsmart/schema"
)

type ModelInfo struct {
	PackageName     string
	StructName      string
	ShortStructName string
	TableName       string
	Fields          []string
	DBCols          []*sql.ColumnType
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
func GenerateStruct(db *sql.DB,
	sqlDatabase,
	tableName string,
	structName string,
	pkgName string,
	jsonAnnotation bool,
	gormAnnotation bool,
	gureguTypes bool,
	jsonNameFormat string,
	verbose bool) *ModelInfo {

	cols, _ := schema.Table(db, tableName)
	fields := generateFieldsTypes(db, tableName, cols, 0, jsonAnnotation, gormAnnotation, gureguTypes, jsonNameFormat, verbose)

	var modelInfo = &ModelInfo{
		PackageName:     pkgName,
		StructName:      structName,
		TableName:       tableName,
		ShortStructName: strings.ToLower(string(structName[0])),
		Fields:          fields,
		DBCols:          cols,
	}

	return modelInfo
}

// Generate fields string
func generateFieldsTypes(
	db *sql.DB,
	tableName string,
	columns []*sql.ColumnType,
	depth int,
	jsonAnnotation bool,
	gormAnnotation bool,
	gureguTypes bool,
	jsonNameFormat string,
	verbose bool) []string {

	//sort.Strings(keys)

	var fields []string
	var field = ""
	for i, c := range columns {
		nullable, _ := c.Nullable()
		key := c.Name()
		if verbose {
			fmt.Printf("   [%s]   [%d] field: %s type: %s\n", tableName, i, key, c.DatabaseTypeName())
		}

		valueType := sqlTypeToGoType(strings.ToLower(c.DatabaseTypeName()), nullable, gureguTypes)
		if valueType == "" { // unknown type
			if verbose {
				fmt.Printf("table: %s unable to generate struct field: %s type: %s\n", tableName, key, c.DatabaseTypeName())
			}
			continue
		}
		fieldName := FmtFieldName(stringifyFirstChar(key))

		var annotations []string
		if gormAnnotation == true {
			if i == 0 {
				annotations = append(annotations, fmt.Sprintf("gorm:\"column:%s;primary_key\"", key))
			} else {
				annotations = append(annotations, fmt.Sprintf("gorm:\"column:%s\"", key))
			}

		}
		if jsonAnnotation == true {
			var jsonName string
			switch jsonNameFormat {
			case "snake":
				jsonName = strcase.ToSnake(key)
			case "camel":
				jsonName = strcase.ToCamel(key)
			case "lower_camel":
				jsonName = strcase.ToLowerCamel(key)
			case "none":
				jsonName = key
			default:
				jsonName = key
			}

			annotations = append(annotations, fmt.Sprintf("json:\"%s\"", jsonName))
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

		fields = append(fields, field)
	}
	return fields
}

func generateMapTypes(db *sql.DB, columns []*sql.ColumnType, depth int, jsonAnnotation bool, gormAnnotation bool, gureguTypes bool) []string {

	//sort.Strings(keys)

	var fields []string
	var field = ""
	for i, c := range columns {
		nullable, _ := c.Nullable()
		key := c.Name()
		valueType := sqlTypeToGoType(strings.ToLower(c.DatabaseTypeName()), nullable, gureguTypes)
		if valueType == "" { // unknown type
			continue
		}
		fieldName := FmtFieldName(stringifyFirstChar(key))

		var annotations []string
		if gormAnnotation == true {
			if i == 0 {
				annotations = append(annotations, fmt.Sprintf("gorm:\"column:%s;primary_key\"", key))
			} else {
				annotations = append(annotations, fmt.Sprintf("gorm:\"column:%s\"", key))
			}

		}
		if jsonAnnotation == true {
			annotations = append(annotations, fmt.Sprintf("json:\"%s\"", key))
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

		fields = append(fields, field)
	}
	return fields
}

func sqlTypeToGoType(mysqlType string, nullable bool, gureguTypes bool) string {
	mysqlType = strings.Trim(mysqlType, " \t")
	mysqlType = strings.ToLower(mysqlType)

	switch mysqlType {
	case "tinyint", "int", "smallint", "mediumint", "int4", "int2", "integer":
		if nullable {
			if gureguTypes {
				return gureguNullInt
			}
			return sqlNullInt
		}
		return golangInt
	case "bigint", "int8":
		if nullable {
			if gureguTypes {
				return gureguNullInt
			}
			return sqlNullInt
		}
		return golangInt64
	case "char", "enum", "varchar", "longtext", "mediumtext", "text", "tinytext", "varchar2", "json", "jsonb", "nvarchar":
		if nullable {
			if gureguTypes {
				return gureguNullString
			}
			return sqlNullString
		}
		return "string"
	case "date", "datetime", "time", "timestamp":
		if nullable && gureguTypes {
			return gureguNullTime
		}
		return golangTime
	case "decimal", "double":
		if nullable {
			if gureguTypes {
				return gureguNullFloat
			}
			return sqlNullFloat
		}
		return golangFloat64
	case "float":
		if nullable {
			if gureguTypes {
				return gureguNullFloat
			}
			return sqlNullFloat
		}
		return golangFloat32
	case "binary", "blob", "longblob", "mediumblob", "varbinary":
		return golangByteArray
	case "bool":
		return golangBool
	}

	if strings.HasPrefix(mysqlType, "nvarchar") || strings.HasPrefix(mysqlType, "varchar") {
		if nullable {
			if gureguTypes {
				return gureguNullString
			}
			return sqlNullString
		}
		return "string"
	}

	if strings.HasPrefix(mysqlType, "numeric") {
		if nullable {
			if gureguTypes {
				return gureguNullFloat
			}
			return sqlNullFloat
		}
		return golangFloat64
	}

	return ""
}

func IsNullable(colType *sql.ColumnType) bool {
	nullable, _ := colType.Nullable()
	return nullable
}

func ColumnLength(colType *sql.ColumnType) int64 {
	len, _ := colType.Length()
	return len
}

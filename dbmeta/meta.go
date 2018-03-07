package dbmeta

import (
	"database/sql"
	"strings"
)

type ModelInfo struct {
	PackageName     string
	StructName      string
	ShortStructName string
	TableName       string
	Fields          []string
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
)

// GenerateStruct generates a struct for the given table.
func GenerateStruct(db *sql.DB, columnTypes map[string]map[string]string, tableName string, structName string, pkgName string, jsonAnnotation bool, gormAnnotation bool, gureguTypes bool) *ModelInfo {
	fields := generateMysqlTypes(db, columnTypes, 0, jsonAnnotation, gormAnnotation, gureguTypes)

	var modelInfo = &ModelInfo{
		PackageName:     pkgName,
		StructName:      structName,
		TableName:       tableName,
		ShortStructName: strings.ToLower(string(structName[0])),
		Fields:          fields,
	}

	return modelInfo
}

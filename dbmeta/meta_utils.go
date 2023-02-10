package dbmeta

import (
	"database/sql"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/logrusorgru/aurora"
)

var (
	au aurora.Aurora
)

func InitColorOutput(_au aurora.Aurora) {
	au = _au
}

// ParseSQLType parse sql type and return raw type and length
func ParseSQLType(dbType string) (resultType string, dbTypeLen int64) {

	resultType = strings.ToLower(dbType)
	dbTypeLen = -1
	idx1 := strings.Index(resultType, "(")
	idx2 := strings.Index(resultType, ")")

	if idx1 > -1 && idx2 > -1 {
		sizeStr := resultType[idx1+1 : idx2]
		resultType = resultType[0:idx1]
		i, err := strconv.Atoi(sizeStr)
		if err == nil {
			dbTypeLen = int64(i)
		}
	}

	// fmt.Printf("dbType: %-20s %-20s %d\n", dbType, resultType, dbTypeLen)
	return resultType, dbTypeLen
}

// TrimSpaceNewlineInString replace spaces in string
func TrimSpaceNewlineInString(s string) string {

	re := regexp.MustCompile(` +\r?\n +`)
	return re.ReplaceAllString(s, " ")
}

/*
SELECT
    TABLE_SCHEMA, TABLE_NAME, COLUMN_NAME, COLUMN_DEFAULT
FROM
    INFORMATION_SCHEMA.COLUMNS
WHERE
  TABLE_SCHEMA = @SchemaName
  AND TABLE_NAME = @TableName
  AND COLUMN_NAME = @ColumnName;
*/

// FindPrimaryKeyFromInformationSchema fetch primary key info from information_schema
func FindPrimaryKeyFromInformationSchema(db *sql.DB, tableName string) (primaryKey string, err error) {

	primaryKeySQL := fmt.Sprintf(`
SELECT Col.Column_Name from 
    INFORMATION_SCHEMA.TABLE_CONSTRAINTS Tab, 
    INFORMATION_SCHEMA.CONSTRAINT_COLUMN_USAGE Col 
WHERE 
    Col.Constraint_Name = Tab.Constraint_Name
    AND Col.Table_Name = Tab.Table_Name
    AND Constraint_Type = 'PRIMARY KEY'
    AND Col.Table_Name = '%s'
`, tableName)
	res, err := db.Query(primaryKeySQL)
	if err != nil {
		return "", fmt.Errorf("unable to load ddl from ms sql: %v", err)
	}
	defer res.Close()
	for res.Next() {
		var columnName string
		err = res.Scan(&columnName)
		if err != nil {
			return "", fmt.Errorf("unable to load identity info from ms sql Scan: %v", err)
		}

		return columnName, nil
	}
	return "", nil
}

// PostgresInformationSchema results from a query of the postgres InformationSchema db table
type PostgresInformationSchema struct {
	TableCatalog           string
	TableSchema            string
	TableName              string
	OrdinalPosition        int
	ColumnName             string
	DataType               string
	CharacterMaximumLength interface{}
	ColumnDefault          interface{}
	IsNullable             string
	IsIdentity             string
	PrimaryKey             bool
}

// LoadTableInfoFromPostgresInformationSchema fetch info from information_schema for postgres database
func LoadTableInfoFromPostgresInformationSchema(db *sql.DB, tableName string) (primaryKey map[string]*PostgresInformationSchema, err error) {
	colInfo := make(map[string]*PostgresInformationSchema)

	identitySQL := fmt.Sprintf(`
SELECT TABLE_CATALOG, table_schema, table_name, ordinal_position, column_name, data_type, character_maximum_length,
column_default, is_nullable, is_identity 
FROM information_schema.columns
WHERE table_name = '%s' 
ORDER BY table_name, ordinal_position;
`, tableName)

	res, err := db.Query(identitySQL)
	if err != nil {
		return nil, fmt.Errorf("unable to load ddl from %s: %v", tableName, err)
	}
	defer res.Close()
	for res.Next() {
		ci := &PostgresInformationSchema{}
		err = res.Scan(&ci.TableCatalog, &ci.TableSchema, &ci.TableName, &ci.OrdinalPosition, &ci.ColumnName, &ci.DataType, &ci.CharacterMaximumLength,
			&ci.ColumnDefault, &ci.IsNullable, &ci.IsIdentity)
		if err != nil {
			return nil, fmt.Errorf("unable to load identity info from postgres Scan: %v", err)
		}

		colInfo[ci.ColumnName] = ci
	}

	return colInfo, nil
}

// InformationSchema results from a query of the InformationSchema db table
type InformationSchema struct {
	TableCatalog           string
	TableSchema            string
	TableName              string
	OrdinalPosition        int
	ColumnName             string
	DataType               string
	CharacterMaximumLength interface{}
	ColumnDefault          interface{}
	IsNullable             string
}

// LoadTableInfoFromMSSqlInformationSchema fetch info from information_schema for ms sql database
func LoadTableInfoFromMSSqlInformationSchema(db *sql.DB, tableName string) (primaryKey map[string]*InformationSchema, err error) {
	colInfo := make(map[string]*InformationSchema)

	identitySQL := fmt.Sprintf(`
SELECT TABLE_CATALOG, TABLE_SCHEMA, TABLE_NAME, ORDINAL_POSITION, COLUMN_NAME, DATA_TYPE, character_maximum_length,
column_default, is_nullable 
FROM information_schema.columns
WHERE table_name = '%s' 
ORDER BY table_name, ordinal_position;
`, tableName)

	res, err := db.Query(identitySQL)
	if err != nil {
		return nil, fmt.Errorf("unable to load ddl from information_schema: %v", err)
	}
	defer res.Close()
	for res.Next() {
		ci := &InformationSchema{}
		err = res.Scan(&ci.TableCatalog, &ci.TableSchema, &ci.TableName, &ci.OrdinalPosition, &ci.ColumnName, &ci.DataType, &ci.CharacterMaximumLength,
			&ci.ColumnDefault, &ci.IsNullable)

		if err != nil {
			return nil, fmt.Errorf("unable to load identity info from information_schema Scan: %v", err)
		}

		colInfo[ci.ColumnName] = ci
	}

	return colInfo, nil
}

// GetFieldLenFromInformationSchema fetch field length from database
func GetFieldLenFromInformationSchema(db *sql.DB, tableSchema, tableName, columnName string) (int64, error) {
	sql := fmt.Sprintf(`
select CHARACTER_MAXIMUM_LENGTH 
from information_schema.columns
where table_schema = '%s' AND 
      table_name = '%s' AND       
      COLUMN_NAME = '%s'    
`, tableSchema, tableName, columnName)

	res, err := db.Query(sql)
	if err != nil {
		return -1, fmt.Errorf("unable to load col len from mysql: %v", err)
	}

	var colLen int64
	defer res.Close()
	if res.Next() {
		err = res.Scan(&colLen)
		if err != nil {
			return -1, fmt.Errorf("unable to load ddl from mysql Scan: %v", err)
		}
	}
	_ = res.Close()
	return colLen, nil

}

func cleanupDefault(val string) string {
	if len(val) < 2 {
		return val
	}

	if strings.HasPrefix(val, "(") && strings.HasSuffix(val, ")") {
		return cleanupDefault(val[1 : len(val)-1])
	}

	if strings.Index(val, "nextval(") == 0 && strings.Index(val, "::regclass)") > -1 {
		return ""
	}

	if strings.LastIndex(val, "::") > -1 {
		return cleanupDefault(val[0:strings.LastIndex(val, "::")])
	}
	if strings.HasPrefix(val, "'") && strings.HasSuffix(val, "'") {
		return cleanupDefault(val[1 : len(val)-1])
	}
	// 'G'::mpaa_rating
	// ('now'::text)::date

	return val
}

// BytesToString convert []uint8 to string
func BytesToString(bs []uint8) string {
	b := make([]byte, len(bs))
	for i, v := range bs {
		b[i] = byte(v)
	}
	return string(b)
}

func indexAt(s, sep string, n int) int {
	idx := strings.Index(s[n:], sep)
	if idx > -1 {
		idx += n
	}
	return idx
}

func updateDefaultPrimaryKey(m *dbTableMeta) *dbTableMeta {
	hasPrimary := false
	primaryKeyPos := -1
	for i, j := range m.columns {
		if j.IsPrimaryKey() {
			primaryKeyPos = i
			hasPrimary = true
			break
		}
	}

	if !hasPrimary && len(m.columns) > 0 {
		comments := fmt.Sprintf("Warning table: %s does not have a primary key defined, setting col position 1 %s as primary key\n", m.tableName, m.columns[0].Name())
		if au != nil {
			fmt.Print(au.Yellow(comments))
		} else {
			fmt.Printf(comments)
		}

		primaryKeyPos = 0
		m.columns[0].isPrimaryKey = true
		m.columns[0].notes = m.columns[0].notes + comments
	}

	if m.columns[primaryKeyPos].nullable {
		comments := fmt.Sprintf("Warning table: %s primary key column %s is nullable column, setting it as NOT NULL\n", m.tableName, m.columns[primaryKeyPos].Name())

		if au != nil {
			fmt.Print(au.Yellow(comments))
		} else {
			fmt.Printf(comments)
		}

		m.columns[primaryKeyPos].nullable = false
		m.columns[0].notes = m.columns[0].notes + comments
	}
	m.primaryKeyPos = primaryKeyPos
	return m
}

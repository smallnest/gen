package dbmeta

import (
	"database/sql"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

func ParseSqlType(dbType string) (resultType string, dbTypeLen int64) {

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

func FindPrimaryKeyFromInformationSchema(db *sql.DB, tableName string) (primaryKey string, err error) {

	primaryKeySql := fmt.Sprintf(`
SELECT Col.Column_Name from 
    INFORMATION_SCHEMA.TABLE_CONSTRAINTS Tab, 
    INFORMATION_SCHEMA.CONSTRAINT_COLUMN_USAGE Col 
WHERE 
    Col.Constraint_Name = Tab.Constraint_Name
    AND Col.Table_Name = Tab.Table_Name
    AND Constraint_Type = 'PRIMARY KEY'
    AND Col.Table_Name = '%s'
`, tableName)
	res, err := db.Query(primaryKeySql)
	if err != nil {
		return "", fmt.Errorf("unable to load ddl from ms sql: %v", err)
	}
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

type postgresInformationSchema struct {
	TABLE_CATALOG            string
	table_schema             string
	table_name               string
	ordinal_position         int
	column_name              string
	data_type                string
	character_maximum_length interface{}
	column_default           interface{}
	is_nullable              string
	is_identity              string
}

func LoadTableInfoFromPostgresInformationSchema(db *sql.DB, tableName string) (primaryKey map[string]*postgresInformationSchema, err error) {
	colInfo := make(map[string]*postgresInformationSchema)

	identitySql := fmt.Sprintf(`
SELECT TABLE_CATALOG, table_schema, table_name, ordinal_position, column_name, data_type, character_maximum_length,
column_default, is_nullable, is_identity 
FROM information_schema.columns
WHERE table_name = '%s' 
ORDER BY table_name, ordinal_position;
`, tableName)

	res, err := db.Query(identitySql)
	if err != nil {
		return nil, fmt.Errorf("unable to load ddl from postgres: %v", err)
	}

	for res.Next() {
		ci := &postgresInformationSchema{}
		err = res.Scan(&ci.TABLE_CATALOG, &ci.table_schema, &ci.table_name, &ci.ordinal_position, &ci.column_name, &ci.data_type, &ci.character_maximum_length,
			&ci.column_default, &ci.is_nullable, &ci.is_identity)
		if err != nil {
			return nil, fmt.Errorf("unable to load identity info from postgres Scan: %v", err)
		}

		colInfo[ci.column_name] = ci
	}

	return colInfo, nil
}

type mssqlInformationSchema struct {
	TABLE_CATALOG            string
	table_schema             string
	table_name               string
	ordinal_position         int
	column_name              string
	data_type                string
	character_maximum_length interface{}
	column_default           interface{}
	is_nullable              string
}

func LoadTableInfoFromMSSqlInformationSchema(db *sql.DB, tableName string) (primaryKey map[string]*mssqlInformationSchema, err error) {
	colInfo := make(map[string]*mssqlInformationSchema)

	identitySql := fmt.Sprintf(`
SELECT TABLE_CATALOG, TABLE_SCHEMA, TABLE_NAME, ORDINAL_POSITION, COLUMN_NAME, DATA_TYPE, character_maximum_length,
column_default, is_nullable 
FROM information_schema.columns
WHERE table_name = '%s' 
ORDER BY table_name, ordinal_position;
`, tableName)

	res, err := db.Query(identitySql)
	if err != nil {
		return nil, fmt.Errorf("unable to load ddl from postgres: %v", err)
	}

	for res.Next() {
		ci := &mssqlInformationSchema{}
		err = res.Scan(&ci.TABLE_CATALOG, &ci.table_schema, &ci.table_name, &ci.ordinal_position, &ci.column_name, &ci.data_type, &ci.character_maximum_length,
			&ci.column_default, &ci.is_nullable)

		if err != nil {
			return nil, fmt.Errorf("unable to load identity info from postgres Scan: %v", err)
		}

		colInfo[ci.column_name] = ci
	}

	return colInfo, nil
}

func GetFieldLenFromInformationSchema(db *sql.DB, tableSchema, tableName, columnName string) (int64, error) {
	sql := fmt.Sprintf(`
select CHARACTER_MAXIMUM_LENGTH 
from information_schema.columns
where table_schema = %s AND 
      table_name = '%s' AND       
      COLUMN_NAME = '%s'    
`, tableSchema, tableName, columnName)

	res, err := db.Query(sql)
	if err != nil {
		return -1, fmt.Errorf("unable to load col len from mysql: %v", err)
	}

	var colLen int64

	if res.Next() {
		err = res.Scan(&colLen)
		if err != nil {
			return -1, fmt.Errorf("unable to load ddl from mysql Scan: %v", err)
		}
	}

	return colLen, nil

}

func cleanupDefault(val string) string {
	if strings.HasPrefix(val, "(") && strings.HasSuffix(val, ")") {
		return cleanupDefault(val[1 : len(val)-1])
	}

	if strings.Index(val, "nextval(") == 0 && strings.Index(val, "::regclass)") > -1 {
		return ""
	}

	if strings.LastIndex(val, "::") > -1 {
		return cleanupDefault(val[0:strings.LastIndex(val, "::")])
	}
	// 'G'::mpaa_rating
	// ('now'::text)::date

	return val
}


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

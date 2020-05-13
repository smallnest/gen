package dbmeta

import (
	"database/sql"
	"fmt"

	"github.com/jimsmart/schema"
)

// LoadPostgresMeta fetch db meta data for Postgres database
func LoadPostgresMeta(db *sql.DB, sqlType, sqlDatabase, tableName string) (DbTableMeta, error) {
	m := &dbTableMeta{
		sqlType:     sqlType,
		sqlDatabase: sqlDatabase,
		tableName:   tableName,
	}

	cols, err := schema.Table(db, m.tableName)
	if err != nil {
		return nil, err
	}
	m.columns = make([]ColumnMeta, len(cols))

	colInfo, err := postgresLoadInfoSchema(db, tableName)
	if err != nil {
		return nil, fmt.Errorf("unable to load identity info schema from postgres table: %s error: %v", tableName, err)
	}

	err = postgresLoadPrimaryKey(db, tableName, colInfo)
	if err != nil {
		return nil, fmt.Errorf("unable to load primary key from postgres: %v", err)
	}

	for i, v := range cols {
		defaultVal := ""
		nullable, ok := v.Nullable()
		if !ok {
			nullable = false
		}
		isAutoIncrement := false
		isPrimaryKey := i == 0
		var maxLen int64

		maxLen = -1
		colInfo, ok := colInfo[v.Name()]
		if ok {
			nullable = colInfo.isNullable == "YES"
			isAutoIncrement = colInfo.isIdentity == "YES"
			isPrimaryKey = colInfo.primaryKey
			if colInfo.columnDefault != nil {
				defaultVal = cleanupDefault(fmt.Sprintf("%v", colInfo.columnDefault))
			}

			ml, ok := colInfo.characterMaximumLength.(int64)
			if ok {
				maxLen = ml
			}
		}

		definedType := v.DatabaseTypeName()

		if definedType == "" {
			definedType = "USER_DEFINED"
		}

		colDDL := v.DatabaseTypeName()

		colMeta := &columnMeta{
			index:           i,
			ct:              v,
			nullable:        nullable,
			isPrimaryKey:    isPrimaryKey,
			isAutoIncrement: isAutoIncrement,
			colDDL:          colDDL,
			columnLen:       maxLen,
			columnType:      definedType,
			defaultVal:      defaultVal,
		}

		m.columns[i] = colMeta

		// fmt.Printf("@@ Name: %v columnLen: %v\n", colMeta.Name(), colMeta.columnLen)

	}
	m.ddl = BuildDefaultTableDDL(tableName, m.columns)
	return m, nil
}

func postgresLoadPrimaryKey(db *sql.DB, tableName string, colInfo map[string]*postgresColumnInfo) error {
	primaryKeySQL := fmt.Sprintf(`
	SELECT c.column_name
	FROM information_schema.key_column_usage AS c
	LEFT JOIN information_schema.table_constraints AS t
	ON t.constraint_name = c.constraint_name
	WHERE t.table_name = '%s' AND t.constraint_type = 'PRIMARY KEY';
`, tableName)
	res, err := db.Query(primaryKeySQL)
	if err != nil {
		return fmt.Errorf("unable to load ddl from ms sql: %v", err)
	}
	for res.Next() {

		var columnName string
		err = res.Scan(&columnName)
		if err != nil {
			return fmt.Errorf("unable to load identity info from ms sql Scan: %v", err)
		}

		//fmt.Printf("## PRIMARY KEY COLUMN_NAME: %s\n", columnName)
		colInfo, ok := colInfo[columnName]
		if ok {
			colInfo.primaryKey = true
			//fmt.Printf("name: %s primary_key: %t\n", colInfo.name, colInfo.primary_key)
		}
	}
	return nil
}

type postgresColumnInfo struct {
	tableName              string
	columnName             string
	ordinalPosition        int
	tableSchema            string
	dataType               string
	characterMaximumLength interface{}
	columnDefault          interface{}
	isNullable             string
	isIdentity             string
	udtName                string
	numericPrecision       interface{}
	primaryKey             bool
}

func (ci *postgresColumnInfo) String() string {
	return fmt.Sprintf("[%2d] %-20s %-20s data_type: %v character_maximum_length: %v is_nullable: %v is_identity: %v", ci.ordinalPosition, ci.tableName, ci.columnName, ci.dataType, ci.characterMaximumLength, ci.isNullable, ci.isIdentity)
}

func postgresLoadInfoSchema(db *sql.DB, tableName string) (colInfo map[string]*postgresColumnInfo, err error) {
	colInfo = make(map[string]*postgresColumnInfo)

	identitySQL := fmt.Sprintf(`
SELECT table_name, table_schema, ordinal_position, column_name, data_type, character_maximum_length,
column_default, is_nullable, is_identity, udt_name, numeric_precision
FROM information_schema.columns
WHERE table_name = '%s' and table_schema = 'public'
ORDER BY table_name, ordinal_position;
`, tableName)

	res, err := db.Query(identitySQL)
	if err != nil {
		return nil, fmt.Errorf("unable to load ddl from postgres: %v", err)
	}

	for res.Next() {
		ci := &postgresColumnInfo{}
		err = res.Scan(&ci.tableName, &ci.tableSchema, &ci.ordinalPosition, &ci.columnName, &ci.dataType, &ci.characterMaximumLength,
			&ci.columnDefault, &ci.isNullable, &ci.isIdentity, &ci.udtName, &ci.numericPrecision)
		if err != nil {
			return nil, fmt.Errorf("unable to load identity info from postgres Scan: %v", err)
		}

		colInfo[ci.columnName] = ci
	}
	return colInfo, nil
}

/*
https://dataedo.com/kb/query/postgresql/list-table-default-constraints

select col.table_schema,
       col.table_name,
       col.column_name,
       col.column_default
from information_schema.columns col
where col.column_default is not null
      and col.table_schema not in('information_schema', 'pg_catalog')
order by col.column_name;
*/

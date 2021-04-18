package dbmeta

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jimsmart/schema"
)

// LoadPostgresMeta fetch db meta data for Postgres database
func LoadPostgresMeta(db *sql.DB, sqlType, sqlDatabase, tableName string) (DbTableMeta, error) {
	m := &dbTableMeta{
		sqlType:     sqlType,
		sqlDatabase: sqlDatabase,
		tableName:   tableName,
	}
	cols, err := schema.ColumnTypes(db, sqlDatabase, tableName)
	if err != nil {
		cols, err = schema.ColumnTypes(db, "", tableName)
		if err != nil {
			return nil, err
		}
	}
	m.columns = make([]*columnMeta, len(cols))

	colInfo, err := LoadTableInfoFromPostgresInformationSchema(db, tableName)
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
			nullable = colInfo.IsNullable == "YES"
			isAutoIncrement = colInfo.IsIdentity == "YES"
			isPrimaryKey = colInfo.PrimaryKey

			if colInfo.ColumnDefault != nil {
				defaultVal = cleanupDefault(fmt.Sprintf("%v", colInfo.ColumnDefault))
			}

			ml, ok := colInfo.CharacterMaximumLength.(int64)
			if ok {
				maxLen = ml
			}
		}

		definedType := v.DatabaseTypeName()
		colDDL := v.DatabaseTypeName()
		if definedType == "" {
			definedType = "USER_DEFINED"
			colDDL = "VARCHAR"
		}

		colMeta := &columnMeta{
			index:            i,
			name:             v.Name(),
			databaseTypeName: colDDL,
			nullable:         nullable,
			isPrimaryKey:     isPrimaryKey,
			isAutoIncrement:  isAutoIncrement,
			colDDL:           colDDL,
			columnLen:        maxLen,
			columnType:       definedType,
			defaultVal:       defaultVal,
		}

		m.columns[i] = colMeta
	}

	m.ddl = BuildDefaultTableDDL(tableName, m.columns)
	m = updateDefaultPrimaryKey(m)

	for _, v := range m.columns {
		if !v.isAutoIncrement && v.isPrimaryKey {
			val := fmt.Sprintf("%v", v.defaultVal)
			if strings.Index(val, "()") > -1 {
				v.isAutoIncrement = true
			}
		}
	}
	for _, v := range m.columns {
		if strings.HasPrefix(v.DatabaseTypeName(), "_") {
			v.isArray = true
		}
	}
	return m, nil
}

func postgresLoadPrimaryKey(db *sql.DB, tableName string, colInfo map[string]*PostgresInformationSchema) error {
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

	defer res.Close()
	for res.Next() {

		var columnName string
		err = res.Scan(&columnName)
		if err != nil {
			return fmt.Errorf("unable to load identity info from ms sql Scan: %v", err)
		}

		// fmt.Printf("## PRIMARY KEY COLUMN_NAME: %s\n", columnName)
		colInfo, ok := colInfo[columnName]
		if ok {
			colInfo.PrimaryKey = true
			// fmt.Printf("name: %s primary_key: %t\n", colInfo.name, colInfo.primary_key)
		}
	}
	return nil
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

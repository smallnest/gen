package dbmeta

import (
	"database/sql"
	"fmt"

	"github.com/jimsmart/schema"
)

func NewPostgresMeta(db *sql.DB, sqlType, sqlDatabase, tableName string) (DbTableMeta, error) {
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

	colInfo := make(map[string]*postgresColumnInfo)

	identitySql := fmt.Sprintf(`
SELECT table_name, table_schema, ordinal_position, column_name, data_type, character_maximum_length,
column_default, is_nullable, is_identity, udt_name, numeric_precision
FROM information_schema.columns
WHERE table_name = '%s' and table_schema = 'public'
ORDER BY table_name, ordinal_position;
`, tableName)

	res, err := db.Query(identitySql)
	if err != nil {
		return nil, fmt.Errorf("unable to load ddl from postgres: %v", err)
	}

	for res.Next() {
		ci := &postgresColumnInfo{}
		err = res.Scan(&ci.table_name, &ci.table_schema, &ci.ordinal_position, &ci.column_name, &ci.data_type, &ci.character_maximum_length,
			&ci.column_default, &ci.is_nullable, &ci.is_identity, &ci.udt_name, &ci.numeric_precision)
		if err != nil {
			return nil, fmt.Errorf("unable to load identity info from postgres Scan: %v", err)
		}

		colInfo[ci.column_name] = ci
	}

	primaryKeySql := fmt.Sprintf(`
	SELECT c.column_name
	FROM information_schema.key_column_usage AS c
	LEFT JOIN information_schema.table_constraints AS t
	ON t.constraint_name = c.constraint_name
	WHERE t.table_name = '%s' AND t.constraint_type = 'PRIMARY KEY';
`, tableName)
	res, err = db.Query(primaryKeySql)
	if err != nil {
		return nil, fmt.Errorf("unable to load ddl from ms sql: %v", err)
	}
	for res.Next() {

		var columnName string
		err = res.Scan(&columnName)
		if err != nil {
			return nil, fmt.Errorf("unable to load identity info from ms sql Scan: %v", err)
		}

		//fmt.Printf("## PRIMARY KEY COLUMN_NAME: %s\n", columnName)
		colInfo, ok := colInfo[columnName]
		if ok {
			colInfo.primary_key = true
			//fmt.Printf("name: %s primary_key: %t\n", colInfo.name, colInfo.primary_key)
		}
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
			nullable = colInfo.is_nullable == "YES"
			isAutoIncrement = colInfo.is_identity == "YES"
			isPrimaryKey = colInfo.primary_key
			if colInfo.column_default != nil {
				defaultVal = cleanupDefault(fmt.Sprintf("%v", colInfo.column_default))
			}

			ml, ok := colInfo.character_maximum_length.(int64)
			if ok {
				// fmt.Printf("@@ Name: %v maxLen: %v\n", v.Name(), ml)
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
	if err != nil {
		return nil, err
	}
	m.ddl = BuildDefaultTableDDL(tableName, m.columns)
	return m, nil
}

type postgresColumnInfo struct {
	table_name               string
	column_name              string
	ordinal_position         int
	table_schema             string
	data_type                string
	character_maximum_length interface{}
	column_default           interface{}
	is_nullable              string
	is_identity              string
	udt_name                 string
	numeric_precision        interface{}
	primary_key              bool
}

func (ci *postgresColumnInfo) String() string {
	return fmt.Sprintf("[%2d] %-20s %-20s data_type: %v character_maximum_length: %v is_nullable: %v is_identity: %v", ci.ordinal_position, ci.table_name, ci.column_name, ci.data_type, ci.character_maximum_length, ci.is_nullable, ci.is_identity)
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

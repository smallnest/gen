package dbmeta

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jimsmart/schema"
)

func NewMsSqlMeta(db *sql.DB, sqlType, sqlDatabase, tableName string) (DbTableMeta, error) {
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

	colInfo := make(map[string]*msSqlColumnInfo)

	identitySql := fmt.Sprintf("SELECT name, is_identity, is_nullable, max_length FROM sys.columns WHERE  object_id = object_id('dbo.%s')", tableName)

	res, err := db.Query(identitySql)
	if err != nil {
		return nil, fmt.Errorf("unable to load ddl from ms sql: %v", err)
	}

	for res.Next() {
		var name string
		var is_identity, is_nullable bool
		var max_length int64
		err = res.Scan(&name, &is_identity, &is_nullable, &max_length)
		if err != nil {
			return nil, fmt.Errorf("unable to load identity info from ms sql Scan: %v", err)
		}

		colInfo[name] = &msSqlColumnInfo{
			name:        name,
			is_identity: is_identity,
			is_nullable: is_nullable,
			max_length:  max_length,
		}

	}

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

	infoSchema, err := LoadTableInfoFromMSSqlInformationSchema(db, tableName)
	if err != nil {
		fmt.Printf("error calling LoadTableInfoFromMSSqlInformationSchema table: %s error: %v\n", tableName, err)
	}

	for i, v := range cols {

		nullable, ok := v.Nullable()
		if !ok {
			nullable = false
		}
		isAutoIncrement := false
		isPrimaryKey := i == 0
		var columnLen int64 = -1

		colInfo, ok := colInfo[v.Name()]
		if ok {
			// fmt.Printf("name: %s DatabaseTypeName: %s primary_key: %t is_identity: %t is_nullable: %t max_length: %d\n", colInfo.name, v.DatabaseTypeName(), colInfo.primary_key, colInfo.is_identity, colInfo.is_nullable, colInfo.max_length)
			isPrimaryKey = colInfo.primary_key
			nullable = colInfo.is_nullable
			isAutoIncrement = colInfo.is_identity
			dbType := strings.ToLower(v.DatabaseTypeName())
			if strings.Contains(dbType, "char") || strings.Contains(dbType, "text") {
				columnLen = colInfo.max_length
			}
		} else {
			fmt.Printf("name: %s DatabaseTypeName: %s NOT FOUND in colInfo\n", v.Name(), v.DatabaseTypeName())
		}

		defaultVal := ""
		columnType := v.DatabaseTypeName()
		colDDL := v.DatabaseTypeName()

		if infoSchema != nil {
			infoSchemaColInfo, ok := infoSchema[v.Name()]
			if ok {
				if infoSchemaColInfo.column_default != nil {
					defaultVal = fmt.Sprintf("%v", infoSchemaColInfo.column_default)
					defaultVal = cleanupDefault(defaultVal)
				}
			}
		}

		colMeta := &columnMeta{
			index:           i,
			ct:              v,
			nullable:        nullable,
			isPrimaryKey:    isPrimaryKey,
			isAutoIncrement: isAutoIncrement,
			colDDL:          colDDL,
			defaultVal:      defaultVal,
			columnType:      columnType,
			columnLen:       columnLen,
		}

		m.columns[i] = colMeta
	}
	if err != nil {
		return nil, err
	}
	m.ddl = BuildDefaultTableDDL(tableName, m.columns)

	return m, nil
}

type msSqlColumnInfo struct {
	name        string
	is_identity bool
	is_nullable bool
	primary_key bool
	max_length  int64
}

/*
https://www.mssqltips.com/sqlservertip/1512/finding-and-listing-all-columns-in-a-sql-server-database-with-default-values/


SELECT SO.NAME AS "Table Name", SC.NAME AS "Column Name", SM.TEXT AS "Default Value"
FROM dbo.sysobjects SO INNER JOIN dbo.syscolumns SC ON SO.id = SC.id
LEFT JOIN dbo.syscomments SM ON SC.cdefault = SM.id
WHERE SO.xtype = 'U'
ORDER BY SO.[name], SC.colid


*/

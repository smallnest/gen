package dbmeta

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jimsmart/schema"
)

// LoadMsSQLMeta fetch db meta data for MS SQL database
func LoadMsSQLMeta(db *sql.DB, sqlType, sqlDatabase, tableName string) (DbTableMeta, error) {
	m := &dbTableMeta{
		sqlType:     sqlType,
		sqlDatabase: sqlDatabase,
		tableName:   tableName,
	}

	cols, err := schema.ColumnTypes(db, sqlDatabase, tableName)
	if err != nil {
		return nil, err
	}

	m.columns = make([]*columnMeta, len(cols))
	colInfo, err := msSQLloadFromSysColumns(db, tableName)
	if err != nil {
		return nil, fmt.Errorf("unable to load ddl from ms sql: %v", err)
	}

	err = msSQLLoadPrimaryKey(db, tableName, colInfo)
	if err != nil {
		return nil, fmt.Errorf("unable to load ddl from ms sql: %v", err)
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
			isPrimaryKey = colInfo.primaryKey
			nullable = colInfo.isNullable
			isAutoIncrement = colInfo.isIdentity
			dbType := strings.ToLower(v.DatabaseTypeName())

			if strings.Contains(dbType, "char") || strings.Contains(dbType, "text") {
				columnLen = colInfo.maxLength
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
				if infoSchemaColInfo.ColumnDefault != nil {
					defaultVal = fmt.Sprintf("%v", infoSchemaColInfo.ColumnDefault)
					defaultVal = cleanupDefault(defaultVal)
				}
			}
		}

		colMeta := &columnMeta{
			index:            i,
			name:             v.Name(),
			databaseTypeName: columnType,
			nullable:         nullable,
			isPrimaryKey:     isPrimaryKey,
			isAutoIncrement:  isAutoIncrement,
			colDDL:           colDDL,
			defaultVal:       defaultVal,
			columnType:       columnType,
			columnLen:        columnLen,
		}

		m.columns[i] = colMeta
	}

	m.ddl = BuildDefaultTableDDL(tableName, m.columns)
	m = updateDefaultPrimaryKey(m)
	return m, nil
}

func msSQLLoadPrimaryKey(db *sql.DB, tableName string, colInfo map[string]*msSQLColumnInfo) error {
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
			colInfo.primaryKey = true
			// fmt.Printf("name: %s primary_key: %t\n", colInfo.name, colInfo.primary_key)
		}
	}
	return nil
}

func msSQLloadFromSysColumns(db *sql.DB, tableName string) (colInfo map[string]*msSQLColumnInfo, err error) {
	colInfo = make(map[string]*msSQLColumnInfo)

	identitySQL := fmt.Sprintf(`
SELECT name, is_identity, is_nullable, max_length 
FROM sys.columns 
WHERE  object_id = object_id('dbo.%s')`, tableName)

	res, err := db.Query(identitySQL)
	if err != nil {
		return nil, fmt.Errorf("unable to load ddl from ms sql: %v", err)
	}

	defer res.Close()
	for res.Next() {
		var name string
		var isIdentity, isNullable bool
		var maxLength int64
		err = res.Scan(&name, &isIdentity, &isNullable, &maxLength)
		if err != nil {
			return nil, fmt.Errorf("unable to load identity info from ms sql Scan: %v", err)
		}

		colInfo[name] = &msSQLColumnInfo{
			name:       name,
			isIdentity: isIdentity,
			isNullable: isNullable,
			maxLength:  maxLength,
		}
	}
	return colInfo, err
}

type msSQLColumnInfo struct {
	name       string
	isIdentity bool
	isNullable bool
	primaryKey bool
	maxLength  int64
}

/*
https://www.mssqltips.com/sqlservertip/1512/finding-and-listing-all-columns-in-a-sql-server-database-with-default-values/
*/

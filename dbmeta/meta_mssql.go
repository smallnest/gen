package dbmeta

import (
	"database/sql"
	"fmt"

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
	m.ddl = BuildDefaultTableDDL(tableName, cols)
	m.columns = make([]ColumnMeta, len(cols))

	colInfo := make(map[string]*columnInfo)

	identitySql := fmt.Sprintf("SELECT name, is_identity, is_nullable FROM sys.columns WHERE  object_id = object_id('dbo.%s')", tableName)

	res, err := db.Query(identitySql)
	if err != nil {
		return nil, fmt.Errorf("unable to load ddl from ms sql: %v", err)
	}

	for res.Next() {
		var name string
		var is_identity, is_nullable bool
		err = res.Scan(&name, &is_identity, &is_nullable)
		if err != nil {
			return nil, fmt.Errorf("unable to load identity info from ms sql Scan: %v", err)
		}

		colInfo[name] = &columnInfo{
			name:        name,
			is_identity: is_identity,
			is_nullable: is_nullable,
		}

	}

	primaryKeySql := fmt.Sprintf("SELECT Col.Column_Name from \n    INFORMATION_SCHEMA.TABLE_CONSTRAINTS Tab, \n    INFORMATION_SCHEMA.CONSTRAINT_COLUMN_USAGE Col \nWHERE \n    Col.Constraint_Name = Tab.Constraint_Name\n    AND Col.Table_Name = Tab.Table_Name\n    AND Constraint_Type = 'PRIMARY KEY'\n    AND Col.Table_Name = '%s'", tableName)
	res, err = db.Query(primaryKeySql)
	if err != nil {
		return nil, fmt.Errorf("unable to load ddl from ms sql: %v", err)
	}
	for res.Next() {

		var columnName string
		err = res.Scan( &columnName)
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

		nullable, ok := v.Nullable()
		if !ok {
			nullable = false
		}
		isAutoIncrement := false
		isPrimaryKey := i == 0

		colInfo, ok := colInfo[v.Name()]
		if ok {
			//fmt.Printf("name: %s primary_key: %t is_identity: %t is_nullable: %t\n", colInfo.name, colInfo.primary_key, colInfo.is_identity, colInfo.is_nullable)
			isPrimaryKey = colInfo.primary_key
			nullable = colInfo.is_nullable
			isAutoIncrement = colInfo.is_identity
		}

		colDDL := v.DatabaseTypeName()

		colMeta := &columnMeta{
			index:           i,
			ct:              v,
			nullable:        nullable,
			isPrimaryKey:    isPrimaryKey,
			isAutoIncrement: isAutoIncrement,
			colDDL:          colDDL,
		}
		m.columns[i] = colMeta
	}
	if err != nil {
		return nil, err
	}

	return m, nil
}

type columnInfo struct {
	name        string
	is_identity bool
	is_nullable bool
	primary_key bool
}

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

	for i, v := range cols {

		nullable, ok := v.Nullable()
		if !ok {
			nullable = false
		}
		isAutoIncrement := false
		isPrimaryKey := i==0
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

func fetchMSSqlTable(db *sql.DB, tableName string) string {
	/*
		SELECT ORDINAL_POSITION, COLUMN_NAME, DATA_TYPE, CHARACTER_MAXIMUM_LENGTH
		       , IS_NULLABLE
		FROM INFORMATION_SCHEMA.COLUMNS
		WHERE TABLE_NAME = 'Customers'

		SELECT CONSTRAINT_NAME
		FROM INFORMATION_SCHEMA.CONSTRAINT_TABLE_USAGE
		WHERE TABLE_NAME = 'Customers'

		SELECT name, type_desc, is_unique, is_primary_key
		FROM sys.indexes
		WHERE [object_id] = OBJECT_ID('dbo.Customers')

		sp_help 'TableName'


	*/
	var ddl string
	sql := fmt.Sprintf("%s", tableName)
	db.Query(sql)

	row := db.QueryRow(sql, 0)
	_ = row.Scan(&ddl)

	fmt.Printf("fetchMSsqlTable: %s\n\n\n", ddl)
	return ddl
}

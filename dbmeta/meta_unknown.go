package dbmeta

import (
	"database/sql"
	"regexp"
	"strconv"
	"strings"

	"github.com/jimsmart/schema"
)

func NewUnknownMeta(db *sql.DB, sqlType, sqlDatabase, tableName string) (DbTableMeta, error) {
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
		isPrimaryKey := i == 0
		colDDL := v.DatabaseTypeName()
		colMeta := &columnMeta{
			index:           i,
			ct:              v,
			nullable:        nullable,
			isPrimaryKey:    isPrimaryKey,
			isAutoIncrement: isAutoIncrement,
			colDDL:          colDDL,
		}

		dbType := strings.ToLower(colMeta.DatabaseTypeName())
		if strings.Contains(dbType, "varchar") {
			re := regexp.MustCompile(`[-]?\d[\d,]*[\.]?[\d{2}]*`)
			submatchall := re.FindAllString(dbType, -1)
			if len(submatchall) > 0 {
				i, err := strconv.Atoi(submatchall[0])
				if err == nil {
					colMeta.columnLen = int64(i)
				}
			}
		}

		m.columns[i] = colMeta
	}
	if err != nil {
		return nil, err
	}

	return m, nil
}

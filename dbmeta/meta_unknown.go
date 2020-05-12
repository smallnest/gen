package dbmeta

import (
	"database/sql"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/jimsmart/schema"
)


// NewUnknownMeta fetch db meta data for unknown database type
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

	m.columns = make([]ColumnMeta, len(cols))

	infoSchema, err := LoadTableInfoFromMSSqlInformationSchema(db, tableName)
	if err != nil {
		fmt.Printf("NOTICE unable to load InformationSchema table: %s error: %v\n", tableName, err)
	}

	for i, v := range cols {

		nullable, ok := v.Nullable()
		if !ok {
			nullable = false
		}
		isAutoIncrement := false
		isPrimaryKey := i == 0
		colDDL := v.DatabaseTypeName()

		defaultVal := ""
		columnType, columnLen := ParseSQLType(v.DatabaseTypeName())

		if columnLen == -1 {

			dbType := strings.ToLower(v.DatabaseTypeName())
			if strings.Contains(dbType, "varchar") {
				re := regexp.MustCompile(`[-]?\d[\d,]*[\.]?[\d{2}]*`)
				submatchall := re.FindAllString(dbType, -1)
				if len(submatchall) > 0 {
					i, err := strconv.Atoi(submatchall[0])
					if err == nil {
						columnLen = int64(i)
					}
				}
			}
		}

		if infoSchema != nil {
			infoSchemaColInfo, ok := infoSchema[v.Name()]
			if ok {
				if infoSchemaColInfo.ColumnDefault != nil {
					defaultVal = BytesToString(infoSchemaColInfo.ColumnDefault.([]uint8))
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

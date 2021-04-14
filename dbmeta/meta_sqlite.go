package dbmeta

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jimsmart/schema"
)

// LoadSqliteMeta fetch db meta data for Sqlite3 database
func LoadSqliteMeta(db *sql.DB, sqlType, sqlDatabase, tableName string) (DbTableMeta, error) {
	if tableName == "sqlite_sequence" || tableName == "sqlite_stat1" {
		return nil, fmt.Errorf("unsupported table: %s", tableName)
	}

	m := &dbTableMeta{
		sqlType:     sqlType,
		sqlDatabase: sqlDatabase,
		tableName:   tableName,
	}

	ddl, err := sqliteLoadDDL(db, tableName)
	if err != nil {
		return nil, fmt.Errorf("unable to load ddl from sqlite_master: %v", err)
	}

	m.ddl = ddl

	colsInfos, err := sqliteLoadPragma(db, tableName)
	if err != nil {
		return nil, fmt.Errorf("unable to load PRAGMA table_info %s: %v", m.tableName, err)
	}

	colsDDL := sqliteParseDDL(ddl)

	cols, err := schema.ColumnTypes(db, sqlDatabase, tableName)
	if err != nil {
		return nil, err
	}

	m.columns = make([]*columnMeta, len(cols))

	for i, v := range cols {
		colDDL := colsDDL[v.Name()]

		colDDLLower := strings.ToLower(colDDL)
		notNull := strings.Index(colDDLLower, "not null") > -1
		isPrimaryKey := strings.Index(colDDLLower, "primary key") > -1
		isAutoIncrement := strings.Index(colDDLLower, "autoincrement") > -1
		defaultVal := ""
		columnLen := int64(-1)
		columnType := v.DatabaseTypeName()

		details, ok := colsInfos[v.Name()]
		if ok {
			isPrimaryKey = details.primaryKey == 1
			if details.dfltValue != nil {
				defaultVal = details.dfltValue.(string)
			}

			notNull = details.notnull == 1
			columnType, columnLen = ParseSQLType(details.dataType)
		}

		if isPrimaryKey {
			notNull = true
		}
		// fmt.Printf("%s: notNull: %v isPrimaryKey: %v isAutoIncrement: %v\n",colDDL, notNull, isPrimaryKey, isAutoIncrement)

		colMeta := &columnMeta{
			index:            i,
			name:             v.Name(),
			databaseTypeName: columnType,
			nullable:         !notNull,
			isPrimaryKey:     isPrimaryKey,
			isAutoIncrement:  isAutoIncrement,
			colDDL:           colDDL,
			defaultVal:       defaultVal,
			columnType:       columnType,
			columnLen:        columnLen,
		}

		m.columns[i] = colMeta
	}

	m = updateDefaultPrimaryKey(m)
	return m, nil
}

func sqliteLoadPragma(db *sql.DB, tableName string) (colsInfos map[string]*sqliteColumnInfo, err error) {
	pragmaSQL := fmt.Sprintf("PRAGMA table_info('%s');", tableName)
	res, err := db.Query(pragmaSQL)
	if err != nil {
		return nil, fmt.Errorf("unable to load PRAGMA table_info %s: %v", tableName, err)
	}

	defer res.Close()
	colsInfos = make(map[string]*sqliteColumnInfo)
	for res.Next() {
		ci := &sqliteColumnInfo{}
		err = res.Scan(&ci.cid, &ci.name, &ci.dataType, &ci.notnull, &ci.dfltValue, &ci.primaryKey)
		if err != nil {
			return nil, fmt.Errorf("unable to load identity info from sqlite Scan: %v", err)
		}
		colsInfos[ci.name] = ci

		// fmt.Printf("cid: |%2d| name: |%-20s| data_type: |%-20s| notnull: |%d| dflt_value: |%-10T| dflt_value: |%-10v| primary_key: |%d|\n",
		//	ci.cid, ci.name, ci.data_type, ci.notnull, ci.dflt_value, ci.dflt_value, ci.primary_key)
	}
	return colsInfos, nil
}

func sqliteParseDDL(ddl string) map[string]string {
	idx1 := strings.Index(ddl, "(")
	idx2 := strings.LastIndex(ddl, ")")

	if idx1 > -1 && idx2 > -1 {
		ddl = ddl[idx1+1 : idx2]
	}

	ddl = strings.Replace(ddl, "\r", "", -1)
	ddl = strings.Replace(ddl, "\n", " ", -1)
	ddl = strings.TrimPrefix(ddl, "\n")
	ddl = strings.TrimSuffix(ddl, "\n")

	colsDDL := make(map[string]string)

	lines := strings.Split(ddl, ",")
	for _, line := range lines {
		line = strings.Replace(line, "\n", " ", -1)

		line := strings.TrimSpace(line)

		line = TrimSpaceNewlineInString(line)
		line = strings.TrimPrefix(line, "\n")
		line = strings.TrimSuffix(line, "\n")
		line = strings.TrimSuffix(line, ",")
		line = strings.Trim(line, " ")
		line = strings.Trim(line, ",")

		if len(line) == 0 {
			continue
		}

		if strings.HasPrefix(line, "FOREIGN KEY") || strings.HasPrefix(line, "CONSTRAINT") {
			continue
		}

		// fmt.Printf("[%2d] %s\n", i, line)

		parts := strings.Split(line, " ")
		name := parts[0]
		colDDL := strings.Join(parts[1:], " ")

		name = strings.Trim(name, " \t[]\"")
		colDDL = strings.Trim(colDDL, " \t,")

		if len(colDDL) == 0 {
			continue
		}

		colsDDL[name] = colDDL

	}
	return colsDDL
}

func sqliteLoadDDL(db *sql.DB, tableName string) (string, error) {
	var ddl string
	ddlSQL := fmt.Sprintf("SELECT sql FROM sqlite_master WHERE type='table' and name = '%s';", tableName)
	//_, err := db.Query(ddlSQL)
	//if err != nil {
	//return "", fmt.Errorf("unable to load ddl from sqlite_master: %v", err)
	//}

	row := db.QueryRow(ddlSQL, 0)
	err := row.Scan(&ddl)
	if err != nil {
		return "", err
	}

	return ddl, nil
}

type sqliteColumnInfo struct {
	cid        int
	name       string
	dataType   string
	notnull    int
	dfltValue  interface{}
	primaryKey int
}

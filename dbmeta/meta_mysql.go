package dbmeta

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	"github.com/jimsmart/schema"
)

// LoadMysqlMeta fetch db meta data for MySQL database
func LoadMysqlMeta(db *sql.DB, sqlType, sqlDatabase, tableName string) (DbTableMeta, error) {
	m := &dbTableMeta{
		sqlType:     sqlType,
		sqlDatabase: sqlDatabase,
		tableName:   tableName,
	}

	cols, err := schema.ColumnTypes(db, sqlDatabase, tableName)
	if err != nil {
		return nil, err
	}

	ddl, err := mysqlLoadDDL(db, tableName)
	if err != nil {
		return nil, fmt.Errorf("mysqlLoadDDL - unable to load ddl from mysql: %v", err)
	}

	m.ddl = ddl
	colsDDL, primaryKeys := mysqlParseDDL(ddl)

	infoSchema, err := LoadTableInfoFromMSSqlInformationSchema(db, tableName)
	if err != nil {
		fmt.Printf("error calling LoadTableInfoFromMSSqlInformationSchema table: %s error: %v\n", tableName, err)
	}

	m.columns = make([]*columnMeta, len(cols))

	for i, v := range cols {
		notes := ""
		nullable, ok := v.Nullable()
		if !ok {
			nullable = false
		}

		colDDL := colsDDL[v.Name()]
		isAutoIncrement := strings.Index(colDDL, "AUTO_INCREMENT") > -1
		isUnsigned := strings.Index(colDDL, " unsigned ") > -1 || strings.Index(colDDL, " UNSIGNED ") > -1

		_, isPrimaryKey := find(primaryKeys, v.Name())
		defaultVal := ""
		columnType, columnLen := ParseSQLType(v.DatabaseTypeName())

		if isUnsigned {
			notes = notes + " column is set for unsigned"
			columnType = "u" + columnType
		}

		comment := ""
		commentIdx := strings.Index(colDDL, "COMMENT '")
		if commentIdx > -1 {
			re := regexp.MustCompile("COMMENT '(.*?)'")
			match := re.FindStringSubmatch(colDDL)
			if len(match) > 0 {
				comment = match[1]
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
			notes:            strings.Trim(notes, " "),
			comment:          comment,
		}

		dbType := strings.ToLower(colMeta.DatabaseTypeName())
		// fmt.Printf("dbType: %s\n", dbType)

		if strings.Contains(dbType, "char") || strings.Contains(dbType, "text") {
			columnLen, err := GetFieldLenFromInformationSchema(db, sqlDatabase, tableName, v.Name())
			if err == nil {
				colMeta.columnLen = columnLen
			}
		}

		m.columns[i] = colMeta
	}

	m = updateDefaultPrimaryKey(m)
	return m, nil
}

func mysqlLoadDDL(db *sql.DB, tableName string) (ddl string, err error) {
	ddlSQL := fmt.Sprintf("SHOW CREATE TABLE `%s`;", tableName)
	res, err := db.Query(ddlSQL)
	if err != nil {
		return "", fmt.Errorf("unable to load ddl from mysql: %v", err)
	}

	defer res.Close()
	var ddl1 string
	var ddl2 string
	if res.Next() {
		err = res.Scan(&ddl1, &ddl2)
		if err != nil {
			return "", fmt.Errorf("unable to load ddl from mysql Scan: %v", err)
		}
	}
	return ddl2, nil
}

func mysqlParseDDL(ddl string) (colsDDL map[string]string, primaryKeys []string) {
	colsDDL = make(map[string]string)
	lines := strings.Split(ddl, "\n")
	for _, line := range lines {
		line = strings.Trim(line, " \t")
		if strings.HasPrefix(line, "CREATE TABLE") || strings.HasPrefix(line, "(") || strings.HasPrefix(line, ")") {
			continue
		}

		if line[0] == '`' {
			idx := indexAt(line, "`", 1)
			if idx > 0 {
				name := line[1:idx]
				colDDL := line[idx+1 : len(line)-1]
				colsDDL[name] = colDDL
			}
		} else if strings.HasPrefix(line, "PRIMARY KEY") {
			primaryKeyNums := strings.Count(line, "`") / 2
			count := 0
			currentIdx := 0
			idxL := 0
			idxR := 0
			for {
				if count >= primaryKeyNums {
					break
				}
				count++
				idxL = indexAt(line, "`", currentIdx)
				currentIdx = idxL + 1
				idxR = indexAt(line, "`", currentIdx)
				currentIdx = idxR + 1
				primaryKeys = append(primaryKeys, line[idxL+1:idxR])
			}
		}
	}
	return
}

func find(slice []string, val string) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}

/*
https://dataedo.com/kb/query/mysql/list-table-default-constraints

select table_schema as database_name,
       table_name,
       column_name,
       column_default
from information_schema.columns
where  column_default is not null
      and table_schema not in ('information_schema', 'sys',
                               'performance_schema','mysql')
--    and table_schema = 'your database name'
order by table_schema,
         table_name,
         ordinal_position;
*/

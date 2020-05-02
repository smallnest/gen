package dbmeta

import (
	"database/sql"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/jimsmart/schema"
)

func NewSqliteMeta(db *sql.DB, sqlType, sqlDatabase, tableName string) (DbTableMeta, error) {
	m := &dbTableMeta{
		sqlType:     sqlType,
		sqlDatabase: sqlDatabase,
		tableName:   tableName,
	}
	sql := fmt.Sprintf("SELECT sql FROM sqlite_master WHERE type='table' and name = '%s';", m.tableName)
	_, err := db.Query(sql)
	if err != nil {
		return nil, fmt.Errorf("unable to load ddl from sqlite_master: %v", err)
	}

	row := db.QueryRow(sql, 0)
	err = row.Scan(&m.ddl)
	if err != nil {
		return nil, err
	}

	/*

		CREATE TABLE "employees"
		(
		    [EmployeeId] INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
		    [LastName] NVARCHAR(20)  NOT NULL,
		    [FirstName] NVARCHAR(20)  NOT NULL,
		    [Title] NVARCHAR(30),
		    [ReportsTo] INTEGER,
		    [BirthDate] DATETIME,
		    [HireDate] DATETIME,
		    [Address] NVARCHAR(70),
		    [City] NVARCHAR(40),
		    [State] NVARCHAR(40),
		    [Country] NVARCHAR(40),
		    [PostalCode] NVARCHAR(10),
		    [Phone] NVARCHAR(24),
		    [Fax] NVARCHAR(24),
		    [Email] NVARCHAR(60),
		    FOREIGN KEY ([ReportsTo]) REFERENCES "employees" ([EmployeeId])
				ON DELETE NO ACTION ON UPDATE NO ACTION
		)
	*/

	colsDDL := make(map[string]string)
	lines := strings.Split(m.ddl, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "CREATE TABLE") || strings.HasPrefix(line, "(") || strings.HasPrefix(line, ")") {
			continue
		}
		line = strings.Trim(line, " \t")

		if line[0] == '[' {
			idx := strings.Index(line, "]")
			if idx > 0 {
				name := line[1:idx]
				colDDL := line[idx+1:]

				name = strings.Trim(name, " \t")
				colDDL = strings.Trim(colDDL, " \t")
				colDDL = colDDL[0 : len(colDDL)-1]

				sz := len(colDDL)

				if sz > 0 && colDDL[sz-1] == ',' {
					colDDL = colDDL[0 : sz-1]
				}

				//fmt.Printf("name: [%s]\n", name)
				//fmt.Printf("colDDL: [%s]\n", colDDL)
				colsDDL[name] = colDDL
			}
		}
	}

	cols, err := schema.Table(db, m.tableName)
	if err != nil {
		return nil, err
	}

	m.columns = make([]ColumnMeta, len(cols))

	for i, v := range cols {
		colDDL := colsDDL[v.Name()]
		notNull := strings.Index(colDDL, "NOT NULL") > -1
		isPrimaryKey := strings.Index(colDDL, "PRIMARY KEY") > -1
		isAutoIncrement := strings.Index(colDDL, "AUTOINCREMENT") > -1

		// fmt.Printf("%s: notNull: %v isPrimaryKey: %v isAutoIncrement: %v\n",colDDL, notNull, isPrimaryKey, isAutoIncrement)

		colMeta := &columnMeta{
			index:           i,
			ct:              v,
			nullable:        !notNull,
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

	return m, nil
}

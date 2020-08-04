package main

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/droundy/goopt"
	"github.com/gobuffalo/packr/v2"
	"github.com/jimsmart/schema"

	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"

	"github.com/smallnest/gen/dbmeta"
)

var (
	sqlType       = goopt.String([]string{"--sqltype"}, "mysql", "sql database type such as [ mysql, mssql, postgres, sqlite, etc. ]")
	sqlConnStr    = goopt.String([]string{"-c", "--connstr"}, "nil", "database connection string")
	sqlDatabase   = goopt.String([]string{"-d", "--database"}, "nil", "Database to for connection")
	sqlTable      = goopt.String([]string{"-t", "--table"}, "", "Table to build struct from")
	baseTemplates *packr.Box
)

func init() {
	// Setup goopts
	goopt.Description = func() string {
		return "ORM and RESTful meta data viewer for SQl databases"
	}
	goopt.Version = "v0.9.27 (08/04/2020)"
	goopt.Summary = `dbmeta [-v] --sqltype=mysql --connstr "user:password@/dbname" --database <databaseName> 

           sqltype - sql database type such as [ mysql, mssql, postgres, sqlite, etc. ]

`

	//Parse options
	goopt.Parse(nil)

}

func main() {
	baseTemplates = packr.New("gen", "./template")

	var err error
	var content []byte
	content, err = baseTemplates.Find("mapping.json")
	if err != nil {
		fmt.Printf("Error getting default map[mapping file error: %v\n", err)
		return
	}

	err = dbmeta.ProcessMappings("internal", content, false)
	if err != nil {
		fmt.Printf("Error processing default mapping file error: %v\n", err)
		return
	}

	// Username is required
	if sqlConnStr == nil || *sqlConnStr == "" || *sqlConnStr == "nil" {
		fmt.Printf("sql connection string is required! Add it with --connstr=s\n\n")
		fmt.Println(goopt.Usage())
		return
	}

	if sqlDatabase == nil || *sqlDatabase == "" || *sqlDatabase == "nil" {
		fmt.Printf("Database can not be null\n\n")
		fmt.Println(goopt.Usage())
		return
	}

	var db *sql.DB
	db, err = sql.Open(*sqlType, *sqlConnStr)
	if err != nil {
		fmt.Printf("Error in open database: %v\n\n", err.Error())
		return
	}
	defer func() { _ = db.Close() }()

	err = db.Ping()
	if err != nil {
		fmt.Printf("Error connecting to database: %v\n\n", err.Error())
		return
	}

	var dbTables []string
	// parse or read tables
	if *sqlTable != "" && *sqlTable != "all" {
		dbTables = strings.Split(*sqlTable, ",")
		fmt.Printf("showing meta for table(s): %s\n", *sqlTable)
	} else {
		fmt.Printf("showing meta for all tables\n")
		dbTables, err = schema.TableNames(db)
		if err != nil {
			fmt.Printf("Error in fetching tables information from %s information schema from %s\n", *sqlType, *sqlConnStr)
			return
		}
	}

	for _, tableName := range dbTables {
		fmt.Printf("---------------------------\n")
		if strings.HasPrefix(tableName, "[") && strings.HasSuffix(tableName, "]") {
			tableName = tableName[1 : len(tableName)-1]
		}

		fmt.Printf("[%s]\n", tableName)

		tableInfo, err := dbmeta.LoadMeta(*sqlType, db, *sqlDatabase, tableName)

		if err != nil {
			fmt.Printf("Error getting table info for %s error: %v\n\n\n\n", tableName, err)
			continue
		}

		fmt.Printf("\n\nDDL\n%s\n\n\n", tableInfo.DDL())

		for _, col := range tableInfo.Columns() {
			fmt.Printf("%s\n", col.String())

			colMapping, err := dbmeta.SQLTypeToMapping(strings.ToLower(col.DatabaseTypeName()))
			if err != nil { // unknown type
				fmt.Printf("unable to find mapping for db type: %s\n", col.DatabaseTypeName())
				continue
			}
			fmt.Printf("     %s\n", colMapping.String())
		}

		primaryCnt := dbmeta.PrimaryKeyCount(tableInfo)
		fmt.Printf("primaryCnt: %d\n", primaryCnt)

		fmt.Printf("\n\n")
		delSQL, err := dbmeta.GenerateDeleteSQL(tableInfo)
		if err == nil {
			fmt.Printf("delSQL: %s\n", delSQL)
		}

		updateSQL, err := dbmeta.GenerateUpdateSQL(tableInfo)
		if err == nil {
			fmt.Printf("updateSQL: %s\n", updateSQL)
		}

		insertSQL, err := dbmeta.GenerateInsertSQL(tableInfo)
		if err == nil {
			fmt.Printf("insertSQL: %s\n", insertSQL)
		}

		selectOneSQL, err := dbmeta.GenerateSelectOneSQL(tableInfo)
		if err == nil {
			fmt.Printf("selectOneSQL: %s\n", selectOneSQL)
		}

		selectMultiSQL, err := dbmeta.GenerateSelectMultiSQL(tableInfo)
		if err == nil {
			fmt.Printf("selectMultiSQL: %s\n", selectMultiSQL)
		}
	}
}

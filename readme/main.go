package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"path/filepath"

	_ "github.com/denisenkom/go-mssqldb"
	"github.com/droundy/goopt"
	"github.com/gobuffalo/packr/v2"
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
	templateDir   = goopt.String([]string{"--templateDir"}, "./template", "Template Dir")
	baseTemplates *packr.Box
)

func init() {
	// Setup goopts
	goopt.Description = func() string {
		return "ORM and RESTful meta data viewer for SQl databases"
	}
	goopt.Version = "0.9.6 (06/06/2020)"
	goopt.Summary = `dbmeta [-v] --sqltype=mysql --connstr "user:password@/dbname" --database <databaseName> 

           sqltype - sql database type such as [ mysql, mssql, postgres, sqlite, etc. ]

`
	//Parse options
	goopt.Parse(nil)
}

func GenHelp() string {
	cmd := exec.Command("./gen", "-h")
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}

	//fmt.Printf("%s\n", stdoutStderr)
	return string(stdoutStderr)
}

func main() {

	baseTemplates = packr.New("gen", "../template")

	err := loadDefaultDBMappings()
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

	db, err := initializeDB()
	if err != nil {
		return
	}

	defer db.Close()

	conf := dbmeta.NewConfig(LoadTemplate)
	initialize(conf)

	dbTables := []string{*sqlTable}
	tableInfos := dbmeta.LoadTableInfo(db, dbTables, conf)
	conf.ContextMap["tableInfos"] = tableInfos

	for tableName, modelInfo := range tableInfos {
		fmt.Printf("%-15s %v\n", tableName, modelInfo.StructName)
		ctx := conf.CreateContextForTableFile(tableInfos[*sqlTable])

		genreadme(conf, "code_dao_gorm.md.tmpl", "./code_dao_gorm.md", ctx)
		genreadme(conf, "code_dao_sqlx.md.tmpl", "./code_dao_sqlx.md", ctx)
		genreadme(conf, "code_http.md.tmpl", "./code_http.md", ctx)

		help := GenHelp()
		conf.ContextMap["GenHelp"] = help
		genreadme(conf, "GEN_README.md.tmpl", "./README.md", ctx)
	}

}

func genreadme(conf *dbmeta.Config, templateName, outputFile string, ctx map[string]interface{}) {
	template, err := LoadTemplate(templateName)
	if err != nil {
		fmt.Printf("Error loading template %v\n", err)
		return
	}

	sample := `
{{ range $i, $table := .tables }}
    {{$singular   := singular $table -}}
    {{$plural     := pluralize $table -}}
    {{$title      := title $table -}}
    {{$lower      := toLower $table -}}
    {{$lowerCamel := toLowerCamelCase $table -}}
    {{$snakeCase  := toSnakeCase $table -}}
    {{ printf "[%-2d] %-20s %-20s %-20s %-20s %-20s %-20s %-20s" $i $table $singular $plural $title $lower $lowerCamel $snakeCase}}{{- end }}


{{ range $i, $table := .tables }}
   {{$name := toUpper $table -}}
   {{$filename  := printf "My%s" $name -}}
   {{ printf "[%-2d] %-20s %-20s" $i $table $filename}}
   {{ GenerateTableFile $table  "custom.go.tmpl" "test" $filename true}}
{{- end }}
`
	ctx["AdvancesSample"] = sample

	if template != "" {
		// fmt.Printf("%s\n", template)
		conf.WriteTemplate(templateName, template, ctx, outputFile, false)
	}
}

func initialize(conf *dbmeta.Config) {
	outDir := "."
	module := "github.com/alexj212/test"
	modelPackageName := "model"
	daoPackageName := "dao"
	apiPackageName := "api"

	conf.SqlType = *sqlType
	conf.SqlDatabase = *sqlDatabase
	conf.ModelPackageName = modelPackageName
	conf.DaoPackageName = daoPackageName
	conf.ApiPackageName = apiPackageName

	conf.AddJSONAnnotation = true
	conf.AddGormAnnotation = true
	conf.AddProtobufAnnotation = true
	conf.AddDBAnnotation = true
	conf.UseGureguTypes = false
	conf.JsonNameFormat = ""
	conf.ProtobufNameFormat = ""
	conf.Verbose = false
	conf.OutDir = outDir
	conf.Overwrite = true

	conf.SqlConnStr = *sqlConnStr
	conf.ServerPort = 8080
	conf.ServerHost = "127.0.0.1"
	conf.Overwrite = true

	conf.Module = module
	conf.ModelFQPN = module + "/" + modelPackageName
	conf.DaoFQPN = module + "/" + daoPackageName
	conf.ApiFQPN = module + "/" + apiPackageName

	conf.Swagger.Version = "1.0.0"
	conf.Swagger.BasePath = "/"
	conf.Swagger.Title = fmt.Sprintf("Sample CRUD api for %s db", *sqlDatabase)
	conf.Swagger.Description = fmt.Sprintf("Sample CRUD api for %s db", *sqlDatabase)
	conf.Swagger.TOS = "My Custom TOS"
	conf.Swagger.ContactName = ""
	conf.Swagger.ContactURL = ""
	conf.Swagger.ContactEmail = ""
	conf.Swagger.Host = fmt.Sprintf("%s:%d", conf.ServerHost, conf.ServerPort)
}

func initializeDB() (db *sql.DB, err error) {

	db, err = sql.Open(*sqlType, *sqlConnStr)
	if err != nil {
		fmt.Printf("Error in open database: %v\n\n", err.Error())
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		fmt.Printf("Error pinging database: %v\n\n", err.Error())
		return
	}

	return
}

func loadDefaultDBMappings() error {
	var err error
	var content []byte
	content, err = baseTemplates.Find("mapping.json")
	if err != nil {
		return err
	}

	err = dbmeta.ProcessMappings(content, false)
	if err != nil {
		return err
	}
	return nil
}

func LoadTemplate(filename string) (content string, err error) {
	if *templateDir != "" {
		fpath := filepath.Join(*templateDir, filename)
		var b []byte
		b, err = ioutil.ReadFile(fpath)
		if err == nil {
			fmt.Printf("Loaded template from file: %s\n", fpath)
			content = string(b)
			return content, nil
		}
	}
	content, err = baseTemplates.FindString(filename)
	if err != nil {
		return "", fmt.Errorf("%s not found", filename)
	}
	fmt.Printf("Loaded template from app: %s\n", filename)

	return content, nil
}

package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"go/format"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"github.com/droundy/goopt"
	"github.com/howeyc/gopass"
	"github.com/jimsmart/schema"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/jinzhu/inflection"
	"github.com/serenize/snaker"
	"github.com/smallnest/gen/dbmeta"
	gtmpl "github.com/smallnest/gen/template"
)

var (
	mysqlHost     = goopt.String([]string{"-H", "--host"}, "", "Host to check mysql status of")
	mysqlPort     = goopt.Int([]string{"--mysql_port"}, 3306, "Specify a port to connect to")
	mysqlUser     = goopt.String([]string{"-u", "--user"}, "user", "user to connect to database")
	mysqlPassword *string

	mysqlDatabase = goopt.String([]string{"-d", "--database"}, "nil", "Database to for connection")
	mysqlTable    = goopt.String([]string{"-t", "--table"}, "", "Table to build struct from")

	packageName = goopt.String([]string{"--package"}, "", "name to set for package")

	jsonAnnotation = goopt.Flag([]string{"--json"}, []string{"--no-json"}, "Add json annotations (default)", "Disable json annotations")
	gormAnnotation = goopt.Flag([]string{"--gorm"}, []string{}, "Add gorm annotations (tags)", "")
	gureguTypes    = goopt.Flag([]string{"--guregu"}, []string{}, "Add guregu null types", "")

	rest = goopt.Flag([]string{"--rest"}, []string{}, "Enable generating RESTful api", "")

	verbose = goopt.Flag([]string{"-v", "--verbose"}, []string{}, "Enable verbose output", "")
)

func init() {
	goopt.OptArg([]string{"-p", "--password"}, "", "Mysql password", getMysqlPassword)

	// Setup goopts
	goopt.Description = func() string {
		return "ORM and RESTful API generator for Mysql"
	}
	goopt.Version = "0.1"
	goopt.Summary = "gen [-H] [-p] [-v] --user user --package pkgName --database databaseName --table tableName [--json] [--gorm] [--guregu]"

	//Parse options
	goopt.Parse(nil)

}

func main() {
	// Username is required
	if mysqlUser == nil || *mysqlUser == "user" {
		fmt.Println("Username is required! Add it with --user=name")
		return
	}

	if mysqlPassword == nil || *mysqlPassword == "" {
		fmt.Print("Password: ")
		pass, err := gopass.GetPasswd()
		stringPass := string(pass)
		mysqlPassword = &stringPass
		if err != nil {
			fmt.Println("Error reading password: " + err.Error())
			return
		}
	}

	if *verbose {
		fmt.Println("Connecting to mysql server " + *mysqlHost + ":" + strconv.Itoa(*mysqlPort))
	}

	if mysqlDatabase == nil || *mysqlDatabase == "" {
		fmt.Println("Database can not be null")
		return
	}

	var err error
	var db *sql.DB
	if *mysqlPassword != "" {
		db, err = sql.Open("mysql", *mysqlUser+":"+*mysqlPassword+"@tcp("+*mysqlHost+":"+strconv.Itoa(*mysqlPort)+")/"+*mysqlDatabase+"?&parseTime=True")
	} else {
		db, err = sql.Open("mysql", *mysqlUser+"@tcp("+*mysqlHost+":"+strconv.Itoa(*mysqlPort)+")/"+*mysqlDatabase+"?&parseTime=True")
	}
	defer db.Close()

	// parse or read tables
	var tables []string
	if *mysqlTable != "" {
		tables = strings.Split(*mysqlTable, ",")
	} else {
		tables, err = schema.TableNames(db)
		if err != nil {
			fmt.Println("Error in fetching tables information from mysql information schema")
			return
		}
	}
	// if packageName is not set we need to default it
	if packageName == nil || *packageName == "" {
		*packageName = "model"
	}
	os.Mkdir(*packageName, 0777)

	apiName := "api"
	if *rest {
		os.Mkdir(apiName, 0777)
	}

	t, err := getTemplate(gtmpl.ModelTmpl)
	if err != nil {
		fmt.Println("Error in lading model template")
		return
	}

	ct, err := getTemplate(gtmpl.ControllerTmpl)
	if err != nil {
		fmt.Println("Error in lading controller template")
		return
	}

	var structNames []string

	// generate go files for each table
	for _, tableName := range tables {
		structName := dbmeta.FmtFieldName(tableName)
		structName = inflection.Singular(structName)
		structNames = append(structNames, structName)

		columnDataTypes, err := dbmeta.GetColumnsFromMysqlTable(db, *mysqlDatabase, tableName)

		if err != nil {
			fmt.Println("Error in selecting column data information from mysql information schema")
			return
		}

		modelInfo := dbmeta.GenerateStruct(db, *columnDataTypes, tableName, structName, *packageName, *jsonAnnotation, *gormAnnotation, *gureguTypes)

		var buf bytes.Buffer
		err = t.Execute(&buf, modelInfo)
		if err != nil {
			fmt.Println("Error in rendering model: " + err.Error())
			return
		}
		data, err := format.Source(buf.Bytes())
		if err != nil {
			fmt.Println("Error in formating source: " + err.Error())
			return
		}
		ioutil.WriteFile(filepath.Join(*packageName, inflection.Singular(tableName)+".go"), data, 0777)

		if *rest {
			//write api
			buf.Reset()
			err = ct.Execute(&buf, structName)
			if err != nil {
				fmt.Println("Error in rendering controller: " + err.Error())
				return
			}
			data, err = format.Source(buf.Bytes())
			if err != nil {
				fmt.Println("Error in formating source: " + err.Error())
				return
			}
			ioutil.WriteFile(filepath.Join(apiName, inflection.Singular(tableName)+".go"), data, 0777)
		}
	}

	if *rest {
		rt, err := getTemplate(gtmpl.RouterTmpl)
		if err != nil {
			fmt.Println("Error in lading router template")
			return
		}
		var buf bytes.Buffer
		err = rt.Execute(&buf, structNames)
		if err != nil {
			fmt.Println("Error in rendering router: " + err.Error())
			return
		}
		data, err := format.Source(buf.Bytes())
		if err != nil {
			fmt.Println("Error in formating source: " + err.Error())
			return
		}
		ioutil.WriteFile(filepath.Join(apiName, "router.go"), data, 0777)
	}
}

func getMysqlPassword(password string) error {
	mysqlPassword = new(string)
	*mysqlPassword = password
	return nil
}

func getTemplate(t string) (*template.Template, error) {
	var funcMap = template.FuncMap{
		"pluralize":        inflection.Plural,
		"title":            strings.Title,
		"toLower":          strings.ToLower,
		"toLowerCamelCase": camelToLowerCamel,
		"toSnakeCase":      snaker.CamelToSnake,
	}

	tmpl, err := template.New("model").Funcs(funcMap).Parse(t)

	if err != nil {
		return nil, err
	}

	return tmpl, nil
}

func camelToLowerCamel(s string) string {
	ss := strings.Split(s, "")
	ss[0] = strings.ToLower(ss[0])

	return strings.Join(ss, "")
}

package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"go/format"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	_ "github.com/denisenkom/go-mssqldb"
	"github.com/droundy/goopt"
	"github.com/gobuffalo/packd"
	"github.com/gobuffalo/packr/v2"
	"github.com/jimsmart/schema"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/jinzhu/inflection"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/serenize/snaker"

	"github.com/smallnest/gen/dbmeta"
)

type swaggerInfo struct {
	Version      string
	Host         string
	BasePath     string
	Title        string
	Description  string
	TOS          string
	ContactName  string
	ContactUrl   string
	ContactEmail string
}

var (
	sqlType         = goopt.String([]string{"--sqltype"}, "mysql", "sql database type such as mysql, postgres, etc.")
	sqlConnStr      = goopt.String([]string{"-c", "--connstr"}, "nil", "database connection string")
	sqlDatabase     = goopt.String([]string{"-d", "--database"}, "nil", "Database to for connection")
	sqlTable        = goopt.String([]string{"-t", "--table"}, "", "Table to build struct from")
	templateDir     = goopt.String([]string{"--templateDir"}, "", "Template Dir")
	saveTemplateDir = goopt.String([]string{"--save"}, "", "Save templates to dir")

	modelPackageName = goopt.String([]string{"--model"}, "model", "name to set for model package")
	daoPackageName   = goopt.String([]string{"--dao"}, "dao", "name to set for dao package")
	apiPackageName   = goopt.String([]string{"--api"}, "api", "name to set for api package")
	outDir           = goopt.String([]string{"--out"}, ".", "output dir")
	module           = goopt.String([]string{"--module"}, "model", "module path")

	jsonAnnotation   = goopt.Flag([]string{"--json"}, []string{"--no-json"}, "Add json annotations (default)", "Disable json annotations")
	jsonNameFormat   = goopt.String([]string{"--json-fmt"}, "snake", "json name format [snake | camel | lower_camel | none")
	gormAnnotation   = goopt.Flag([]string{"--gorm"}, []string{}, "Add gorm annotations (tags)", "")
	gureguTypes      = goopt.Flag([]string{"--guregu"}, []string{}, "Add guregu null types", "")
	modGenerate      = goopt.Flag([]string{"--mod"}, []string{}, "Generate go.mod in output dir", "")
	makefileGenerate = goopt.Flag([]string{"--makefile"}, []string{}, "Generate MakefileTmpl in output dir", "")
	serverGenerate   = goopt.Flag([]string{"--server"}, []string{}, "Generate server app output dir", "")
	overwrite        = goopt.Flag([]string{"--overwrite"}, []string{"--no-overwrite"}, "Overwrite existing files (default)", "disable overwriting files")

	rest = goopt.Flag([]string{"--rest"}, []string{}, "Enable generating RESTful api", "")

	swaggerVersion      = goopt.String([]string{"--swagger_version"}, "1.0", "swagger version")
	swaggerHost         = goopt.String([]string{"--swagger_host"}, "localhost:8080", "swagger host")
	swaggerBasePath     = goopt.String([]string{"--swagger_path"}, "/", "swagger base path")
	swaggerTos          = goopt.String([]string{"--swagger_tos"}, "", "swagger tos url")
	swaggerContactName  = goopt.String([]string{"--swagger_contact_name"}, "Me", "swagger contact name")
	swaggerContactUrl   = goopt.String([]string{"--swagger_contact_url"}, "http://me.com/terms.html", "swagger contact url")
	swaggerContactEmail = goopt.String([]string{"--swagger_contact_email"}, "me@me.com", "swagger contact email")

	verbose = goopt.Flag([]string{"-v", "--verbose"}, []string{}, "Enable verbose output", "")

	baseTemplates  *packr.Box
	ModelTmpl      string
	ControllerTmpl string
	DaoTmpl        string
	RouterTmpl     string
	DaoInitTmpl    string
	GoModuleTmpl   string
	MainServerTmpl string
	HttpUtilsTmpl  string
	ReadMeTmpl     string
	GitIgnoreTmpl  string
	MakefileTmpl   string

	modelFQPN string
	daoFQPN   string
	apiFQPN   string

	SwaggerInfo = &swaggerInfo{
		Version:      "1.0",
		Host:         "localhost:8080",
		BasePath:     "/",
		Title:        "Swagger Example API",
		Description:  "This is a sample server Petstore server.",
		TOS:          "",
		ContactName:  "",
		ContactUrl:   "",
		ContactEmail: "",
	}
)

func init() {
	// Setup goopts
	goopt.Description = func() string {
		return "ORM and RESTful API generator for Mysql"
	}
	goopt.Version = "0.2"
	goopt.Summary = `gen [-v] --connstr "user:password@/dbname" --package pkgName --database databaseName --table tableName [--json] [--gorm] [--guregu]`

	//Parse options
	goopt.Parse(nil)

}

func main() {

	baseTemplates = packr.New("gen", "./template")

	if *verbose {
		fmt.Printf("Base Template Details: Name: %v Path: %v ResolutionDir: %v\n", baseTemplates.Name, baseTemplates.Path, baseTemplates.ResolutionDir)
	}

	if *verbose {
		for i, file := range baseTemplates.List() {
			fmt.Printf("   [%d] [%s]\n", i, file)
		}
	}

	if *saveTemplateDir != "" {
		fmt.Printf("Saving templates to %s\n", *saveTemplateDir)
		err := SaveAssets(*saveTemplateDir, baseTemplates)
		if err != nil {
			fmt.Printf("Error saving: %v\n", err)
		}
		return
	}

	// Username is required
	if sqlConnStr == nil || *sqlConnStr == "" || *sqlConnStr == "nil" {
		fmt.Println("sql connection string is required! Add it with --connstr=s")
		return
	}

	if sqlDatabase == nil || *sqlDatabase == "" || *sqlDatabase == "nil" {
		fmt.Println("Database can not be null")
		return
	}

	var db, err = sql.Open(*sqlType, *sqlConnStr)
	if err != nil {
		fmt.Println("Error in open database: " + err.Error())
		return
	}
	defer db.Close()
	err = db.Ping()
	if err != nil {
		fmt.Println("Error connecting to database: " + err.Error())
		return
	}

	// parse or read tables
	var tables []string
	if *sqlTable != "" {
		tables = strings.Split(*sqlTable, ",")
	} else {
		tables, err = schema.TableNames(db)
		if err != nil {
			fmt.Println("Error in fetching tables information from mysql information schema")
			return
		}
	}

	if outDir == nil || *outDir == "" {
		*outDir = "."
	}

	// if packageName is not set we need to default it
	if modelPackageName == nil || *modelPackageName == "" {
		*modelPackageName = "model"
	}
	if daoPackageName == nil || *daoPackageName == "" {
		*daoPackageName = "dao"
	}
	if apiPackageName == nil || *apiPackageName == "" {
		*apiPackageName = "api"
	}

	modelDir := filepath.Join(*outDir, *modelPackageName)
	apiDir := filepath.Join(*outDir, *apiPackageName)
	daoDir := filepath.Join(*outDir, *daoPackageName)

	err = os.Mkdir(*outDir, 0777)
	if err != nil && !*overwrite {
		fmt.Printf("unable to create outDir: %s error: %v\n", *outDir, err)
		return
	}

	err = os.Mkdir(modelDir, 0777)
	if err != nil && !*overwrite {
		fmt.Printf("unable to create modelDir: %s error: %v\n", modelDir, err)
		return
	}

	err = os.Mkdir(daoDir, 0777)
	if err != nil && !*overwrite {
		fmt.Printf("unable to create daoDir: %s error: %v\n", daoDir, err)
		return
	}

	if *rest {
		err = os.Mkdir(apiDir, 0777)
		if err != nil && !*overwrite {
			fmt.Printf("unable to create apiDir: %s error: %v\n", apiDir, err)
			return
		}
	}

	modelFQPN = *module + "/" + *modelPackageName
	daoFQPN = *module + "/" + *daoPackageName
	apiFQPN = *module + "/" + *apiPackageName

	if ControllerTmpl, err = loadTemplate("controller.go.tmpl"); err != nil {
		fmt.Printf("Error loading template %v\n", err)
		return
	}
	if DaoTmpl, err = loadTemplate("dao.go.tmpl"); err != nil {
		fmt.Printf("Error loading template %v\n", err)
		return
	}
	if DaoInitTmpl, err = loadTemplate("dao_init.go.tmpl"); err != nil {
		fmt.Printf("Error loading template %v\n", err)
		return
	}
	if GitIgnoreTmpl, err = loadTemplate("GitIgnore.tmpl"); err != nil {
		fmt.Printf("Error loading template %v\n", err)
		return
	}
	if GoModuleTmpl, err = loadTemplate("GoMod.tmpl"); err != nil {
		fmt.Printf("Error loading template %v\n", err)
		return
	}
	if HttpUtilsTmpl, err = loadTemplate("http_utils.go.tmpl"); err != nil {
		fmt.Printf("Error loading template %v\n", err)
		return
	}
	if MainServerTmpl, err = loadTemplate("MainServer.go.tmpl"); err != nil {
		fmt.Printf("Error loading template %v\n", err)
		return
	}
	if MakefileTmpl, err = loadTemplate("Makefile.tmpl"); err != nil {
		fmt.Printf("Error loading template %v\n", err)
		return
	}
	if ModelTmpl, err = loadTemplate("model.go.tmpl"); err != nil {
		fmt.Printf("Error loading template %v\n", err)
		return
	}
	if ReadMeTmpl, err = loadTemplate("README.md.tmpl"); err != nil {
		fmt.Printf("Error loading template %v\n", err)
		return
	}
	if RouterTmpl, err = loadTemplate("router.go.tmpl"); err != nil {
		fmt.Printf("Error loading template %v\n", err)
		return
	}

	SwaggerInfo.Version = *swaggerVersion
	SwaggerInfo.Host = *swaggerHost
	SwaggerInfo.BasePath = *swaggerBasePath
	SwaggerInfo.Title = fmt.Sprintf("Sample CRUD api for %s db", *sqlDatabase)
	SwaggerInfo.Description = fmt.Sprintf("Sample CRUD api for %s db", *sqlDatabase)
	SwaggerInfo.TOS = *swaggerTos
	SwaggerInfo.ContactName = *swaggerContactName
	SwaggerInfo.ContactUrl = *swaggerContactUrl
	SwaggerInfo.ContactEmail = *swaggerContactEmail

	var structNames []string
	*jsonNameFormat = strings.ToLower(*jsonNameFormat)

	// generate go files for each table
	for i, tableName := range tables {
		structName := dbmeta.FmtFieldName(tableName)
		structNameInflection := inflection.Singular(structName)
		structName = inflection.Singular(structName)
		tableInfo := dbmeta.GenerateStruct(db,
			*sqlDatabase,
			tableName,
			structName,
			*modelPackageName,
			*jsonAnnotation,
			*gormAnnotation,
			*gureguTypes,
			*jsonNameFormat,
			*verbose)

		if len(tableInfo.Fields) == 0 {
			if *verbose {
				fmt.Printf("[%d] Table: %s - No Fields Available\n", i, tableName)
			}

			continue
		}

		var modelInfo = map[string]interface{}{
			"ModelPackageName": *modelPackageName,
			"StructName":       structName,
			"TableName":        tableName,
			"ShortStructName":  strings.ToLower(string(structName[0])),
			"Fields":           tableInfo.Fields,
			"DBColumns":        tableInfo.DBCols,
		}

		structNames = append(structNames, structName)
		if *verbose {
			fmt.Printf("[%d] Table: %s Struct: %s inflection: %s\n", i, tableName, structName, structNameInflection)
		}

		modelFile := filepath.Join(modelDir, inflection.Singular(tableName)+".go")
		writeTemplate("model", ModelTmpl, modelInfo, modelFile, *overwrite, true)

		if *rest {
			restData := map[string]interface{}{
				"StructName": structName,
				"TableName":  tableName,
			}

			restFile := filepath.Join(apiDir, inflection.Singular(tableName)+".go")
			writeTemplate("rest", ControllerTmpl, restData, restFile, *overwrite, true)
		}

		//write dao
		daoData := map[string]interface{}{
			"StructName": structName,
			"TableName":  tableName,
		}
		outputFile := filepath.Join(daoDir, inflection.Singular(tableName)+".go")
		writeTemplate("dao", DaoTmpl, daoData, outputFile, *overwrite, true)
	}

	if *rest {
		data := map[string]interface{}{
			"structs": structNames,
		}

		writeTemplate("router", RouterTmpl, data, filepath.Join(apiDir, "router.go"), *overwrite, true)
		writeTemplate("example server", HttpUtilsTmpl, data, filepath.Join(apiDir, "http_utils.go"), *overwrite, true)
	}

	data := map[string]interface{}{
		"structs": structNames,
	}

	writeTemplate("daoBase", DaoInitTmpl, data, filepath.Join(daoDir, "dao_base.go"), *overwrite, true)

	if *modGenerate {
		data := map[string]interface{}{}
		writeTemplate("go.mod", GoModuleTmpl, data, filepath.Join(*outDir, "go.mod"), *overwrite, false)
	}

	if *makefileGenerate {
		data := map[string]interface{}{
			"deps": "go list -f '{{ join .Deps  \"\\n\"}}' .",
		}
		writeTemplate("makefile", MakefileTmpl, data, filepath.Join(*outDir, "Makefile"), *overwrite, false)
		writeTemplate("gitignore", GitIgnoreTmpl, data, filepath.Join(*outDir, ".gitignore"), *overwrite, false)

		cmdLine := strings.Join(os.Args, " ")
		data = map[string]interface{}{"CommandLine": cmdLine}
		writeTemplate("readme", ReadMeTmpl, data, filepath.Join(*outDir, "README.md"), *overwrite, false)
	}

	if *serverGenerate {
		data := map[string]interface{}{
			"models": structNames,
		}

		serverDir := filepath.Join(*outDir, "app/server")
		err = os.MkdirAll(serverDir, 0777)
		if err != nil {
			fmt.Printf("unable to create serverDir: %s error: %v\n", serverDir, err)
			return
		}
		writeTemplate("example server", MainServerTmpl, data, filepath.Join(serverDir, "main.go"), *overwrite, true)
	}
}

func writeTemplate(name, templateStr string, data map[string]interface{}, outputFile string, overwrite, formatOutput bool) {
	if !overwrite && Exists(outputFile) {
		fmt.Printf("not overwriting %s\n", outputFile)
		return
	}

	data["DatabaseName"] = *sqlDatabase
	data["module"] = *module
	data["modelFQPN"] = modelFQPN
	data["daoFQPN"] = daoFQPN
	data["apiFQPN"] = apiFQPN
	data["modelPackageName"] = *modelPackageName
	data["daoPackageName"] = *daoPackageName
	data["apiPackageName"] = *apiPackageName
	data["sqlType"] = *sqlType
	data["sqlConnStr"] = *sqlConnStr
	data["SwaggerInfo"] = SwaggerInfo

	rt, err := getTemplate(templateStr)
	if err != nil {
		fmt.Printf("Error in loading %s template\n", name)
		return
	}
	var buf bytes.Buffer
	err = rt.Execute(&buf, data)
	if err != nil {
		fmt.Printf("Error in rendering %s: %s\n", name, err.Error())
		return
	}

	if formatOutput {
		formattedSource, err := format.Source(buf.Bytes())
		if err != nil {
			fmt.Printf("Error in formatting %s source: %s\n", name, err.Error())
			return
		}
		err = ioutil.WriteFile(outputFile, formattedSource, 0777)
	} else {
		err = ioutil.WriteFile(outputFile, buf.Bytes(), 0777)
	}

	if err != nil {
		fmt.Printf("error writing %s - error: %v\n", outputFile, err)
		return
	}

	if *verbose {
		fmt.Printf("writing %s\n", outputFile)
	}
}

// Exists reports whether the named file or directory exists.
func Exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func getTemplate(t string) (*template.Template, error) {
	var funcMap = template.FuncMap{
		"pluralize":         inflection.Plural,
		"title":             strings.Title,
		"toLower":           strings.ToLower,
		"toLowerCamelCase":  camelToLowerCamel,
		"toSnakeCase":       snaker.CamelToSnake,
		"markdownCodeBlock": markdownCodeBlock,
		"wrapBash":          wrapBash,
		"IsNullable":        dbmeta.IsNullable,
		"ColumnLength":      dbmeta.ColumnLength,
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

func markdownCodeBlock(contentType, content string) string {
	return fmt.Sprintf("```%s\n%s\n```\n", contentType, content)
}

func wrapBash(content string) string {
	parts := strings.Split(content, " ")
	return strings.Join(parts, " \\\n    ")
}

// SaveTemplates will save the prepacked templates for local editing. File structure will be recreated under the output dir.
func SaveAssets(outputDir string, box *packr.Box) error {
	fmt.Printf("SaveAssets: %v\n", outputDir)
	if outputDir == "" {
		outputDir = "."
	}

	if strings.HasSuffix(outputDir, "/") {
		outputDir = outputDir[:len(outputDir)-1]
	}

	if outputDir == "" {
		outputDir = "."
	}

	_ = box.Walk(func(s string, file packd.File) error {
		fileName := fmt.Sprintf("%s/%s", outputDir, s)

		fi, err := file.FileInfo()
		if err == nil {
			if !fi.IsDir() {

				err := WriteNewFile(fileName, file)
				if err != nil {
					return err
				}
			}
		}
		return nil
	})

	return nil
}

// WriteNewFile will attempt to write a file with the filename and path, a Reader and the FileMode of the file to be created.
// If an error is encountered an error will be returned.
func WriteNewFile(fpath string, in io.Reader) error {
	err := os.MkdirAll(filepath.Dir(fpath), 0775)
	if err != nil {
		return fmt.Errorf("%s: making directory for file: %v", fpath, err)
	}

	out, err := os.Create(fpath)
	if err != nil {
		return fmt.Errorf("%s: creating new file: %v", fpath, err)
	}
	defer func() {
		_ = out.Close()
	}()

	fmt.Printf("WriteNewFile: %s\n", fpath)

	_, err = io.Copy(out, in)
	if err != nil {
		return fmt.Errorf("%s: writing file: %v", fpath, err)
	}
	return nil
}

func loadTemplate(filename string) (content string, err error) {
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
	if *verbose {
		fmt.Printf("Loaded template from app: %s\n", filename)
	}

	return content, nil
}

package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
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

type swaggerInfoDetails struct {
	Version      string
	Host         string
	BasePath     string
	Title        string
	Description  string
	TOS          string
	ContactName  string
	ContactURL   string
	ContactEmail string
}

var (
	sqlType         = goopt.String([]string{"--sqltype"}, "mysql", "sql database type such as [ mysql, mssql, postgres, sqlite, etc. ]")
	sqlConnStr      = goopt.String([]string{"-c", "--connstr"}, "nil", "database connection string")
	sqlDatabase     = goopt.String([]string{"-d", "--database"}, "nil", "Database to for connection")
	sqlTable        = goopt.String([]string{"-t", "--table"}, "", "Table to build struct from")
	templateDir     = goopt.String([]string{"--templateDir"}, "", "Template Dir")
	saveTemplateDir = goopt.String([]string{"--save"}, "", "Save templates to dir")

	modelPackageName = goopt.String([]string{"--model"}, "model", "name to set for model package")
	daoPackageName   = goopt.String([]string{"--dao"}, "dao", "name to set for dao package")
	apiPackageName   = goopt.String([]string{"--api"}, "api", "name to set for api package")
	outDir           = goopt.String([]string{"--out"}, ".", "output dir")
	module           = goopt.String([]string{"--module"}, "example.com/example", "module path")
	overwrite        = goopt.Flag([]string{"--overwrite"}, []string{"--no-overwrite"}, "Overwrite existing files (default)", "disable overwriting files")
	contextFileName  = goopt.String([]string{"--context"}, "", "context file (json) to populate context with")
	mappingFileName  = goopt.String([]string{"--mapping"}, "", "mapping file (json) to map sql types to golang/protobuf etc")
	exec             = goopt.String([]string{"--exec"}, "", "execute script for custom code generation")

	jsonAnnotation     = goopt.Flag([]string{"--json"}, []string{"--no-json"}, "Add json annotations (default)", "Disable json annotations")
	jsonNameFormat     = goopt.String([]string{"--json-fmt"}, "snake", "json name format [snake | camel | lower_camel | none]")
	gormAnnotation     = goopt.Flag([]string{"--gorm"}, []string{}, "Add gorm annotations (tags)", "")
	protobufAnnotation = goopt.Flag([]string{"--protobuf"}, []string{}, "Add protobuf annotations (tags)", "")
	protoNameFormat    = goopt.String([]string{"--proto-fmt"}, "snake", "proto name format [snake | camel | lower_camel | none]")
	dbAnnotation       = goopt.Flag([]string{"--db"}, []string{}, "Add db annotations (tags)", "")
	gureguTypes        = goopt.Flag([]string{"--guregu"}, []string{}, "Add guregu null types", "")

	copyTemplates    = goopt.Flag([]string{"--copy-templates"}, []string{}, "Copy regeneration templates to project directory", "")
	modGenerate      = goopt.Flag([]string{"--mod"}, []string{}, "Generate go.mod in output dir", "")
	makefileGenerate = goopt.Flag([]string{"--makefile"}, []string{}, "Generate Makefile in output dir", "")
	serverGenerate   = goopt.Flag([]string{"--server"}, []string{}, "Generate server app output dir", "")
	daoGenerate      = goopt.Flag([]string{"--generate-dao"}, []string{}, "Generate dao functions", "")
	projectGenerate  = goopt.Flag([]string{"--generate-proj"}, []string{}, "Generate project readme an d gitignore", "")
	restAPIGenerate  = goopt.Flag([]string{"--rest"}, []string{}, "Enable generating RESTful api", "")

	serverHost          = goopt.String([]string{"--host"}, "localhost", "host for server")
	serverPort          = goopt.Int([]string{"--port"}, 8080, "port for server")
	swaggerVersion      = goopt.String([]string{"--swagger_version"}, "1.0", "swagger version")
	swaggerBasePath     = goopt.String([]string{"--swagger_path"}, "/", "swagger base path")
	swaggerTos          = goopt.String([]string{"--swagger_tos"}, "", "swagger tos url")
	swaggerContactName  = goopt.String([]string{"--swagger_contact_name"}, "Me", "swagger contact name")
	swaggerContactURL   = goopt.String([]string{"--swagger_contact_url"}, "http://me.com/terms.html", "swagger contact url")
	swaggerContactEmail = goopt.String([]string{"--swagger_contact_email"}, "me@me.com", "swagger contact email")

	verbose = goopt.Flag([]string{"-v", "--verbose"}, []string{}, "Enable verbose output", "")

	baseTemplates *packr.Box

	modelFQPN   string
	daoFQPN     string
	apiFQPN     string
	cmdLine     string
	structNames []string
	tableInfos  = make(map[string]*dbmeta.ModelInfo)
	tables      []string
	contextMap  map[string]interface{}

	swaggerInfo = &swaggerInfoDetails{
		Version:      "1.0",
		BasePath:     "/",
		Title:        "Swagger Example API",
		Description:  "This is a sample server Petstore server.",
		TOS:          "",
		ContactName:  "",
		ContactURL:   "",
		ContactEmail: "",
	}
)

func init() {
	// Setup goopts
	goopt.Description = func() string {
		return "ORM and RESTful API generator for SQl databases"
	}

	goopt.Version = "0.9.3 (05/14/2020)"
	goopt.Summary = `gen [-v] --sqltype=mysql --connstr "user:password@/dbname" --database <databaseName> --module=example.com/example [--json] [--gorm] [--guregu] [--generate-dao] [--generate-proj]

           sqltype - sql database type such as [ mysql, mssql, postgres, sqlite, etc. ]

`

	//Parse options
	goopt.Parse(nil)

}

func saveTemplates() {
	fmt.Printf("Saving templates to %s\n", *saveTemplateDir)
	err := SaveAssets(*saveTemplateDir, baseTemplates)
	if err != nil {
		fmt.Printf("Error saving: %v\n", err)
	}

}
func listTemplates() {
	for i, file := range baseTemplates.List() {
		fmt.Printf("   [%d] [%s]\n", i, file)
	}
}

func loadContextMapping() {
	contextFile, err := os.Open(*contextFileName)
	if err != nil {
		fmt.Printf("Error loading context file %s error: %v\n", *contextFileName, err)
		return
	}

	defer contextFile.Close()
	jsonParser := json.NewDecoder(contextFile)
	err = jsonParser.Decode(&contextMap)
	if err != nil {
		fmt.Printf("Error loading context file %s error: %v\n", *contextFileName, err)
		return
	}

	fmt.Printf("Loaded Context from %s with %d defaults\n", *contextFileName, len(contextMap))
	for key, value := range contextMap {
		fmt.Printf("    Context:%s -> %s\n", key, value)
	}
}

func main() {

	baseTemplates = packr.New("gen", "./template")

	if *verbose {
		listTemplates()
	}

	if *saveTemplateDir != "" {
		saveTemplates()
		return
	}

	if *contextFileName != "" {
		loadContextMapping()
	}

	err := loadDefaultDBMappings()
	if err != nil {
		fmt.Printf("Error processing default mapping file error: %v\n", err)
		return
	}

	if *mappingFileName != "" {
		err := dbmeta.LoadMappings(*mappingFileName)
		if err != nil {
			fmt.Printf("Error loading mappings file %s error: %v\n", *mappingFileName, err)
			return
		}
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

	var dbTables []string
	// parse or read tables
	if *sqlTable != "" {
		dbTables = strings.Split(*sqlTable, ",")
	} else {
		dbTables, err = schema.TableNames(db)
		if err != nil {
			fmt.Printf("Error in fetching tables information from mysql information schema\n")
			return
		}
	}

	fmt.Printf("Generating code for the following tables (%d)\n", len(dbTables))
	for i, tableName := range dbTables {
		fmt.Printf("[%d] %s\n", i, tableName)
	}

	initialize()

	loadTableInfo(db, dbTables)

	if *exec != "" {
		executeCustomScript()
		return
	}

	generate()
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

func initialize() {
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

	modelFQPN = *module + "/" + *modelPackageName
	daoFQPN = *module + "/" + *daoPackageName
	apiFQPN = *module + "/" + *apiPackageName

	swaggerInfo.Version = *swaggerVersion
	swaggerInfo.BasePath = *swaggerBasePath
	swaggerInfo.Title = fmt.Sprintf("Sample CRUD api for %s db", *sqlDatabase)
	swaggerInfo.Description = fmt.Sprintf("Sample CRUD api for %s db", *sqlDatabase)
	swaggerInfo.TOS = *swaggerTos
	swaggerInfo.ContactName = *swaggerContactName
	swaggerInfo.ContactURL = *swaggerContactURL
	swaggerInfo.ContactEmail = *swaggerContactEmail
	swaggerInfo.Host = fmt.Sprintf("%s:%d", *serverHost, *serverPort)
	cmdLine = strings.Join(os.Args, " ")
}

func loadDefaultDBMappings() error {
	var err error
	var content []byte
	content, err = baseTemplates.Find("mapping.json")
	if err != nil {
		return err
	}

	err = dbmeta.ProcessMappings(content)
	if err != nil {
		return err
	}
	return nil
}

func executeCustomScript() {
	fmt.Printf("Executing script %s\n", *exec)

	b, err := ioutil.ReadFile(*exec)
	if err != nil {
		fmt.Printf("Error Loading exec script: %s, error: %v\n", *exec, err)
		return
	}
	content := string(b)
	data := map[string]interface{}{}
	execTemplate("exec", content, data)
}

func loadTableInfo(db *sql.DB, dbTables []string) {
	// generate go files for each table
	var tableIdx = 0
	for i, tableName := range dbTables {
		if strings.HasPrefix(tableName, "[") && strings.HasSuffix(tableName, "]") {
			tableName = tableName[1 : len(tableName)-1]
		}
		structName := dbmeta.FmtFieldName(tableName)
		structName = inflection.Singular(structName)

		tableInfo, err := dbmeta.GenerateStruct(*sqlType,
			db,
			*sqlDatabase,
			tableName,
			structName,
			*modelPackageName,
			*jsonAnnotation,
			*gormAnnotation,
			*protobufAnnotation,
			*dbAnnotation,
			*gureguTypes,
			*jsonNameFormat,
			*protoNameFormat,
			*verbose)

		if err != nil {
			fmt.Printf("Error getting table info for %s error: %v\n", tableName, err)
			continue
		}

		if len(tableInfo.Fields) == 0 {
			if *verbose {
				fmt.Printf("[%d] Table: %s - No Fields Available\n", i, tableName)
			}
			continue
		}

		tableInfo.Index = tableIdx
		tableInfo.IndexPlus1 = tableIdx + 1
		tableIdx++

		tableInfos[tableName] = tableInfo
		structNames = append(structNames, structName)
		tables = append(tables, tableName)
	}

}

func execTemplate(name, templateStr string, data map[string]interface{}) {

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
	data["serverPort"] = *serverPort
	data["serverHost"] = *serverHost
	data["SwaggerInfo"] = swaggerInfo
	data["structs"] = structNames
	data["tableInfos"] = tableInfos
	data["tables"] = tables
	data["CommandLine"] = cmdLine
	data["outDir"] = *outDir

	rt, err := getTemplate(name, templateStr)
	if err != nil {
		fmt.Printf("Error in loading %s template, error: %v\n", name, err)
		return
	}
	var buf bytes.Buffer
	err = rt.Execute(&buf, data)
	if err != nil {
		fmt.Printf("Error in rendering %s: %s\n", name, err.Error())
		return
	}

	fmt.Printf("%s\n", buf.String())
}

func generate() {
	var err error

	*jsonNameFormat = strings.ToLower(*jsonNameFormat)
	modelDir := filepath.Join(*outDir, *modelPackageName)
	apiDir := filepath.Join(*outDir, *apiPackageName)
	daoDir := filepath.Join(*outDir, *daoPackageName)

	err = os.MkdirAll(*outDir, 0777)
	if err != nil && !*overwrite {
		fmt.Printf("unable to create outDir: %s error: %v\n", *outDir, err)
		return
	}

	err = os.MkdirAll(modelDir, 0777)
	if err != nil && !*overwrite {
		fmt.Printf("unable to create modelDir: %s error: %v\n", modelDir, err)
		return
	}

	if *daoGenerate {
		err = os.MkdirAll(daoDir, 0777)
		if err != nil && !*overwrite {
			fmt.Printf("unable to create daoDir: %s error: %v\n", daoDir, err)
			return
		}
	}

	if *restAPIGenerate {
		err = os.MkdirAll(apiDir, 0777)
		if err != nil && !*overwrite {
			fmt.Printf("unable to create apiDir: %s error: %v\n", apiDir, err)
			return
		}
	}
	var ModelTmpl string
	var ModelBaseTmpl string
	var ControllerTmpl string
	var DaoTmpl string

	var DaoInitTmpl string
	var GoModuleTmpl string

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

	if GoModuleTmpl, err = loadTemplate("GoMod.tmpl"); err != nil {
		fmt.Printf("Error loading template %v\n", err)
		return
	}

	if ModelTmpl, err = loadTemplate("model.go.tmpl"); err != nil {
		fmt.Printf("Error loading template %v\n", err)
		return
	}
	if ModelBaseTmpl, err = loadTemplate("model_base.go.tmpl"); err != nil {
		fmt.Printf("Error loading template %v\n", err)
		return
	}

	*jsonNameFormat = strings.ToLower(*jsonNameFormat)

	// generate go files for each table
	for i, tableName := range tables {
		tableInfo, ok := tableInfos[tableName]
		if !ok {
			fmt.Printf("[%d] Table: %s - No tableInfo found\n", i, tableName)
			continue
		}

		if len(tableInfo.Fields) == 0 {
			if *verbose {
				fmt.Printf("[%d] Table: %s - No Fields Available\n", i, tableName)
			}

			continue
		}

		var modelInfo = map[string]interface{}{
			"StructName":      tableInfo.StructName,
			"TableName":       tableName,
			"ShortStructName": strings.ToLower(string(tableInfo.StructName[0])),
			"TableInfo":       tableInfo,
		}

		modelFile := filepath.Join(modelDir, CreateGoSrcFileName(tableName))
		writeTemplate("model", ModelTmpl, modelInfo, modelFile, *overwrite, true)

		if *restAPIGenerate {
			restFile := filepath.Join(apiDir, CreateGoSrcFileName(tableName))
			writeTemplate("rest", ControllerTmpl, modelInfo, restFile, *overwrite, true)
		}

		if *daoGenerate {
			//write dao
			outputFile := filepath.Join(daoDir, CreateGoSrcFileName(tableName))
			writeTemplate("dao", DaoTmpl, modelInfo, outputFile, *overwrite, true)
		}
	}

	data := map[string]interface{}{}

	if *restAPIGenerate {
		if err = generateRestBaseFiles(apiDir); err != nil {
			return
		}
	}

	if *daoGenerate {
		writeTemplate("daoBase", DaoInitTmpl, data, filepath.Join(daoDir, "dao_base.go"), *overwrite, true)
	}

	writeTemplate("modelBase", ModelBaseTmpl, data, filepath.Join(modelDir, "model_base.go"), *overwrite, true)

	if *modGenerate {
		writeTemplate("go.mod", GoModuleTmpl, data, filepath.Join(*outDir, "go.mod"), *overwrite, false)
	}

	if *makefileGenerate {
		if err = generateMakefile(); err != nil {
			return
		}
	}

	if *protobufAnnotation {
		if err = generateProtobufDefinitionFile(data); err != nil {
			return
		}
	}

	data = map[string]interface{}{
		"deps":        "go list -f '{{ join .Deps  \"\\n\"}}' .",
		"CommandLine": cmdLine,
	}

	if *projectGenerate {
		if err = generateProjectFiles(data); err != nil {
			return
		}
	}

	if *serverGenerate {
		if err = generateServerCode(); err != nil {
			return
		}
	}

	if *copyTemplates {
		if err = copyTemplatesToTarget(); err != nil {
			return
		}
	}
}

func generateRestBaseFiles(apiDir string) (err error) {

	data := map[string]interface{}{}
	var RouterTmpl string
	var HTTPUtilsTmpl string

	if HTTPUtilsTmpl, err = loadTemplate("http_utils.go.tmpl"); err != nil {
		fmt.Printf("Error loading template %v\n", err)
		return
	}
	if RouterTmpl, err = loadTemplate("router.go.tmpl"); err != nil {
		fmt.Printf("Error loading template %v\n", err)
		return
	}

	writeTemplate("router", RouterTmpl, data, filepath.Join(apiDir, "router.go"), *overwrite, true)
	writeTemplate("example server", HTTPUtilsTmpl, data, filepath.Join(apiDir, "http_utils.go"), *overwrite, true)
	return nil
}

func generateMakefile() (err error) {
	var MakefileTmpl string

	if MakefileTmpl, err = loadTemplate("Makefile.tmpl"); err != nil {
		fmt.Printf("Error loading template %v\n", err)
		return
	}

	data := map[string]interface{}{
		"deps":         "go list -f '{{ join .Deps  \"\\n\"}}' .",
		"RegenCmdLine": regenCmdLine(),
	}
	writeTemplate("makefile", MakefileTmpl, data, filepath.Join(*outDir, "Makefile"), *overwrite, false)
	return nil
}

func generateProtobufDefinitionFile(data map[string]interface{}) (err error) {
	var ProtobufTmpl string

	if ProtobufTmpl, err = loadTemplate("protobuf.tmpl"); err != nil {
		fmt.Printf("Error loading template %v\n", err)
		return err
	}

	protofile := fmt.Sprintf("%s.proto", *sqlDatabase)
	writeTemplate("protobuf", ProtobufTmpl, data, filepath.Join(*outDir, protofile), *overwrite, false)
	return nil
}

func generateProjectFiles(data map[string]interface{}) (err error) {

	var GitIgnoreTmpl string
	if GitIgnoreTmpl, err = loadTemplate("GitIgnore.tmpl"); err != nil {
		fmt.Printf("Error loading template %v\n", err)
		return
	}
	var ReadMeTmpl string
	if ReadMeTmpl, err = loadTemplate("README.md.tmpl"); err != nil {
		fmt.Printf("Error loading template %v\n", err)
		return
	}

	writeTemplate("gitignore", GitIgnoreTmpl, data, filepath.Join(*outDir, ".gitignore"), *overwrite, false)
	writeTemplate("readme", ReadMeTmpl, data, filepath.Join(*outDir, "README.md"), *overwrite, false)
	return nil
}

func generateServerCode() (err error) {
	data := map[string]interface{}{}
	var MainServerTmpl string
	if MainServerTmpl, err = loadTemplate("MainServer.go.tmpl"); err != nil {
		fmt.Printf("Error loading template %v\n", err)
		return
	}

	serverDir := filepath.Join(*outDir, "app/server")
	err = os.MkdirAll(serverDir, 0777)
	if err != nil {
		fmt.Printf("unable to create serverDir: %s error: %v\n", serverDir, err)
		return
	}
	writeTemplate("example server", MainServerTmpl, data, filepath.Join(serverDir, "main.go"), *overwrite, true)
	return nil
}

func copyTemplatesToTarget() (err error) {
	templatesDir := filepath.Join(*outDir, "templates")
	err = os.MkdirAll(templatesDir, 0777)
	if err != nil && !*overwrite {
		fmt.Printf("unable to create templatesDir: %s error: %v\n", templatesDir, err)
		return
	}

	fmt.Printf("Saving templates to %s\n", templatesDir)
	err = SaveAssets(templatesDir, baseTemplates)
	if err != nil {
		fmt.Printf("Error saving: %v\n", err)
	}
	return nil
}

func regenCmdLine() string {
	buf := bytes.Buffer{}

	buf.WriteString("gen")
	buf.WriteString(fmt.Sprintf(" --sqltype=%s", *sqlType))
	buf.WriteString(fmt.Sprintf(" --connstr=%s", *sqlConnStr))
	buf.WriteString(fmt.Sprintf(" --database=%s", *sqlDatabase))
	buf.WriteString(fmt.Sprintf(" --templateDir=%s", "./templates"))

	if *sqlTable != "" {
		buf.WriteString(fmt.Sprintf(" --table=%s", *sqlTable))
	}

	buf.WriteString(fmt.Sprintf(" --model=%s", *modelPackageName))
	buf.WriteString(fmt.Sprintf(" --dao=%s", *daoPackageName))
	buf.WriteString(fmt.Sprintf(" --api=%s", *apiPackageName))
	buf.WriteString(fmt.Sprintf(" --out=%s", "./"))
	buf.WriteString(fmt.Sprintf(" --module=%s", *module))
	if *jsonAnnotation {
		buf.WriteString(fmt.Sprintf(" --json"))
		buf.WriteString(fmt.Sprintf(" --json-fmt=%s", *jsonNameFormat))
	}
	if *gormAnnotation {
		buf.WriteString(fmt.Sprintf(" --gorm"))
	}
	if *protobufAnnotation {
		buf.WriteString(fmt.Sprintf(" --protobuf"))
	}
	if *dbAnnotation {
		buf.WriteString(fmt.Sprintf(" --db"))
	}
	if *gureguTypes {
		buf.WriteString(fmt.Sprintf(" --guregu"))
	}
	if *modGenerate {
		buf.WriteString(fmt.Sprintf(" --mod"))
	}
	if *makefileGenerate {
		buf.WriteString(fmt.Sprintf(" --makefile"))
	}
	if *serverGenerate {
		buf.WriteString(fmt.Sprintf(" --server"))
	}
	if *overwrite {
		buf.WriteString(fmt.Sprintf(" --overwrite"))
	}

	if *contextFileName != "" {
		buf.WriteString(fmt.Sprintf(" --context=%s", *contextFileName))
	}

	buf.WriteString(fmt.Sprintf(" --host=%s", *serverHost))
	buf.WriteString(fmt.Sprintf(" --port=%d", *serverPort))
	if *restAPIGenerate {
		buf.WriteString(fmt.Sprintf(" --rest"))
	}

	if *daoGenerate {
		buf.WriteString(fmt.Sprintf(" --generate-dao"))
	}
	if *projectGenerate {
		buf.WriteString(fmt.Sprintf(" --generate-proj"))
	}

	if *verbose {
		buf.WriteString(fmt.Sprintf(" --verbose"))
	}

	buf.WriteString(fmt.Sprintf(" --swagger_version=%s", *swaggerVersion))
	buf.WriteString(fmt.Sprintf(" --swagger_path=%s", *swaggerBasePath))
	buf.WriteString(fmt.Sprintf(" --swagger_tos=%s", *swaggerTos))
	buf.WriteString(fmt.Sprintf(" --swagger_contact_name=%s", *swaggerContactName))
	buf.WriteString(fmt.Sprintf(" --swagger_contact_url=%s", *swaggerContactURL))
	buf.WriteString(fmt.Sprintf(" --swagger_contact_email=%s", *swaggerContactEmail))

	regenCmdLine := buf.String()
	regenCmdLine = strings.Trim(regenCmdLine, " \t")
	return regenCmdLine
}

// GenerateTableFile generate file from template using specific table used within templates
func GenerateTableFile(tableName, templateFilename, outputDirectory, outputFileName string, formatOutput bool) string {
	buf := bytes.Buffer{}

	buf.WriteString(fmt.Sprintf("GenerateTableFile( %s, %s, %s, %s, %t)\n", tableName, templateFilename, outputDirectory, outputFileName, formatOutput))

	tableInfo, ok := tableInfos[tableName]
	if !ok {
		buf.WriteString(fmt.Sprintf("Table: %s - No tableInfo found\n", tableName))
		return buf.String()
	}

	if len(tableInfo.Fields) == 0 {
		buf.WriteString(fmt.Sprintf("able: %s - No Fields Available\n", tableName))
		return buf.String()
	}

	var data = map[string]interface{}{
		"StructName":      tableInfo.StructName,
		"TableName":       tableName,
		"ShortStructName": strings.ToLower(string(tableInfo.StructName[0])),
		"TableInfo":       tableInfo,
	}

	fileOutDir := filepath.Join(*outDir, outputDirectory)
	err := os.MkdirAll(fileOutDir, 0777)
	if err != nil && !*overwrite {
		buf.WriteString(fmt.Sprintf("unable to create fileOutDir: %s error: %v\n", fileOutDir, err))
		return buf.String()
	}

	var tpl string
	if tpl, err = loadTemplate(templateFilename); err != nil {
		buf.WriteString(fmt.Sprintf("Error loading template %v\n", err))
		return buf.String()
	}

	outputFile := filepath.Join(fileOutDir, outputFileName)
	buf.WriteString(fmt.Sprintf("Writing %s -> %s\n", templateFilename, outputFile))
	writeTemplate(templateFilename, tpl, data, outputFile, *overwrite, formatOutput)
	return buf.String()
}

// GenerateFile generate file from template, non table used within templates
func GenerateFile(templateFilename, outputDirectory, outputFileName string, formatOutput bool) string {
	buf := bytes.Buffer{}
	buf.WriteString(fmt.Sprintf("GenerateFile( %s, %s, %s)\n", templateFilename, outputDirectory, outputFileName))
	fileOutDir := filepath.Join(*outDir, outputDirectory)
	err := os.MkdirAll(fileOutDir, 0777)
	if err != nil && !*overwrite {
		buf.WriteString(fmt.Sprintf("unable to create fileOutDir: %s error: %v\n", fileOutDir, err))
		return buf.String()
	}

	data := map[string]interface{}{}

	var tpl string
	if tpl, err = loadTemplate(templateFilename); err != nil {
		buf.WriteString(fmt.Sprintf("Error loading template %v\n", err))
		return buf.String()
	}

	outputFile := filepath.Join(fileOutDir, outputFileName)
	buf.WriteString(fmt.Sprintf("Writing %s -> %s\n", templateFilename, outputFile))
	writeTemplate(templateFilename, tpl, data, outputFile, *overwrite, formatOutput)
	return buf.String()
}

func writeTemplate(name, templateStr string, data map[string]interface{}, outputFile string, overwrite, formatOutput bool) {
	if !overwrite && Exists(outputFile) {
		fmt.Printf("not overwriting %s\n", outputFile)
		return
	}

	for key, value := range contextMap {
		data[key] = value
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
	data["serverPort"] = *serverPort
	data["serverHost"] = *serverHost
	data["SwaggerInfo"] = swaggerInfo
	data["structs"] = structNames
	data["tableInfos"] = tableInfos
	data["tables"] = tables
	data["CommandLine"] = cmdLine
	data["outDir"] = *outDir

	rt, err := getTemplate(name, templateStr)
	if err != nil {
		fmt.Printf("Error in loading %s template, error: %v\n", name, err)
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
			formattedSource = buf.Bytes()
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

func getTemplate(name, t string) (*template.Template, error) {
	var funcMap = template.FuncMap{
		"FmtFieldName":      dbmeta.FmtFieldName,
		"singular":          inflection.Singular,
		"pluralize":         inflection.Plural,
		"title":             strings.Title,
		"toLower":           strings.ToLower,
		"toUpper":           strings.ToUpper,
		"toLowerCamelCase":  camelToLowerCamel,
		"toSnakeCase":       snaker.CamelToSnake,
		"markdownCodeBlock": markdownCodeBlock,
		"wrapBash":          wrapBash,
		"GenerateTableFile": GenerateTableFile,
		"GenerateFile":      GenerateFile,
		"ToJSON":            ToJSON,
	}

	tmpl, err := template.New(name).Funcs(funcMap).Parse(t)

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
	// fmt.Printf("%s - %s\n", contentType, content)
	return fmt.Sprintf("```%s\n%s\n```\n", contentType, content)
}

func wrapBash(content string) string {
	// fmt.Printf("wrapBash - %s\n",  content)
	parts := strings.Split(content, " ")
	return strings.Join(parts, " \\\n    ")
}

// SaveAssets will save the prepacked templates for local editing. File structure will be recreated under the output dir.
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

// ToJSON func to return json string representation of struct
func ToJSON(val interface{}, indent int) string {
	pad := fmt.Sprintf("%*s", indent, "")
	strB, _ := json.MarshalIndent(val, "", pad)

	response := string(strB)
	response = strings.Replace(response, "\n", "", -1)
	return response
}

func CreateGoSrcFileName(tableName string) string {
	name := inflection.Singular(tableName)
	if strings.HasSuffix(name, "_test") {
		name = name[0 : len(name)-5]
		name = name + "_tst"
	}
	return name + ".go"
}

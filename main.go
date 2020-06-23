package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"

	_ "github.com/denisenkom/go-mssqldb"
	"github.com/droundy/goopt"
	"github.com/gobuffalo/packd"
	"github.com/gobuffalo/packr/v2"
	"github.com/jimsmart/schema"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"

	"github.com/smallnest/gen/dbmeta"
)

var (
	sqlType          = goopt.String([]string{"--sqltype"}, "mysql", "sql database type such as [ mysql, mssql, postgres, sqlite, etc. ]")
	sqlConnStr       = goopt.String([]string{"-c", "--connstr"}, "nil", "database connection string")
	sqlDatabase      = goopt.String([]string{"-d", "--database"}, "nil", "Database to for connection")
	sqlTable         = goopt.String([]string{"-t", "--table"}, "", "Table to build struct from")
	excludeSqlTables = goopt.String([]string{"-x", "--exclude"}, "", "Table(s) to exclude")
	templateDir      = goopt.String([]string{"--templateDir"}, "", "Template Dir")
	saveTemplateDir  = goopt.String([]string{"--save"}, "", "Save templates to dir")

	modelPackageName    = goopt.String([]string{"--model"}, "model", "name to set for model package")
	modelNamingTemplate = goopt.String([]string{"--model_naming"}, "{{FmtFieldName .}}", "model naming template to name structs")
	fieldNamingTemplate = goopt.String([]string{"--field_naming"}, "{{FmtFieldName (stringifyFirstChar .) }}", "field naming template to name structs")
	fileNamingTemplate  = goopt.String([]string{"--file_naming"}, "{{.}}", "file_naming template to name files")

	daoPackageName      = goopt.String([]string{"--dao"}, "dao", "name to set for dao package")
	apiPackageName      = goopt.String([]string{"--api"}, "api", "name to set for api package")
	grpcPackageName     = goopt.String([]string{"--grpc"}, "grpc", "name to set for grpc package")
	outDir              = goopt.String([]string{"--out"}, ".", "output dir")
	module              = goopt.String([]string{"--module"}, "example.com/example", "module path")
	overwrite           = goopt.Flag([]string{"--overwrite"}, []string{"--no-overwrite"}, "Overwrite existing files (default)", "disable overwriting files")
	contextFileName     = goopt.String([]string{"--context"}, "", "context file (json) to populate context with")
	mappingFileName     = goopt.String([]string{"--mapping"}, "", "mapping file (json) to map sql types to golang/protobuf etc")
	execCustomScript    = goopt.String([]string{"--exec"}, "", "execute script for custom code generation")

	AddJSONAnnotation = goopt.Flag([]string{"--json"}, []string{"--no-json"}, "Add json annotations (default)", "Disable json annotations")
	jsonNameFormat    = goopt.String([]string{"--json-fmt"}, "snake", "json name format [snake | camel | lower_camel | none]")

	AddXMLAnnotation = goopt.Flag([]string{"--xml"}, []string{"--no-xml"}, "Add xml annotations (default)", "Disable xml annotations")
	xmlNameFormat    = goopt.String([]string{"--xml-fmt"}, "snake", "xml name format [snake | camel | lower_camel | none]")

	AddGormAnnotation     = goopt.Flag([]string{"--gorm"}, []string{}, "Add gorm annotations (tags)", "")
	AddProtobufAnnotation = goopt.Flag([]string{"--protobuf"}, []string{}, "Add protobuf annotations (tags)", "")
	protoNameFormat       = goopt.String([]string{"--proto-fmt"}, "snake", "proto name format [snake | camel | lower_camel | none]")
	gogoProtoImport       = goopt.String([]string{"--gogo-proto"}, "", "location of gogo import ")

	AddDBAnnotation = goopt.Flag([]string{"--db"}, []string{}, "Add db annotations (tags)", "")
	UseGureguTypes  = goopt.Flag([]string{"--guregu"}, []string{}, "Add guregu null types", "")

	copyTemplates    = goopt.Flag([]string{"--copy-templates"}, []string{}, "Copy regeneration templates to project directory", "")
	modGenerate      = goopt.Flag([]string{"--mod"}, []string{}, "Generate go.mod in output dir", "")
	makefileGenerate = goopt.Flag([]string{"--makefile"}, []string{}, "Generate Makefile in output dir", "")
	serverGenerate   = goopt.Flag([]string{"--server"}, []string{}, "Generate server app output dir", "")
	daoGenerate      = goopt.Flag([]string{"--generate-dao"}, []string{}, "Generate dao functions", "")
	projectGenerate  = goopt.Flag([]string{"--generate-proj"}, []string{}, "Generate project readme and gitignore", "")
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

	nameTest = goopt.String([]string{"--name_test"}, "", "perform name test using the --model_naming or --file_naming options")

	baseTemplates *packr.Box
	tableInfos    map[string]*dbmeta.ModelInfo
)

func init() {
	// Setup goopts
	goopt.Description = func() string {
		return "ORM and RESTful API generator for SQl databases"
	}
	goopt.Version = "v0.9.14 (06/23/2020)"
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

func loadContextMapping(conf *dbmeta.Config) {
	contextFile, err := os.Open(*contextFileName)
	if err != nil {
		fmt.Printf("Error loading context file %s error: %v\n", *contextFileName, err)
		return
	}

	defer contextFile.Close()
	jsonParser := json.NewDecoder(contextFile)

	err = jsonParser.Decode(&conf.ContextMap)
	if err != nil {
		fmt.Printf("Error loading context file %s error: %v\n", *contextFileName, err)
		return
	}

	fmt.Printf("Loaded Context from %s with %d defaults\n", *contextFileName, len(conf.ContextMap))
	for key, value := range conf.ContextMap {
		fmt.Printf("    Context:%s -> %s\n", key, value)
	}
}

func main() {
	//for i, arg := range os.Args {
	//	fmt.Printf("[%2d] %s\n", i, arg)
	//}


	baseTemplates = packr.New("gen", "./template")

	if *verbose {
		listTemplates()
	}

	if *saveTemplateDir != "" {
		saveTemplates()
		return
	}

	if *nameTest != "" {
		fmt.Printf("Running name test\n")
		fmt.Printf("table name: %s\n", *nameTest)

		fmt.Printf("modelNamingTemplate: %s\n", *modelNamingTemplate)
		result := dbmeta.Replace(*modelNamingTemplate, *nameTest)
		fmt.Printf("model: %s\n", result)

		fmt.Printf("fileNamingTemplate: %s\n", *fileNamingTemplate)
		result = dbmeta.Replace(*modelNamingTemplate, *nameTest)
		fmt.Printf("file: %s\n", result)

		fmt.Printf("fieldNamingTemplate: %s\n", *fieldNamingTemplate)
		result = dbmeta.Replace(*fieldNamingTemplate, *nameTest)
		fmt.Printf("field: %s\n", result)
		os.Exit(0)
		return
	}

	//fmt.Printf("fieldNamingTemplate: %s\n", *fieldNamingTemplate)
	//fmt.Printf("fileNamingTemplate: %s\n", *fileNamingTemplate)
	//fmt.Printf("modelNamingTemplate: %s\n", *modelNamingTemplate)

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
		fmt.Printf("Error in initializing db %v\n", err)
		os.Exit(1)
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
			fmt.Printf("Error in fetching tables information from %s information schema from %s\n", *sqlType, *sqlConnStr)
			os.Exit(1)
			return
		}
	}

	if strings.HasPrefix(*modelNamingTemplate, "'") && strings.HasSuffix(*modelNamingTemplate, "'") {
		*modelNamingTemplate = strings.TrimSuffix(*modelNamingTemplate, "'")
		*modelNamingTemplate = strings.TrimPrefix(*modelNamingTemplate, "'")
	}


	var excludeDbTables []string

	if *excludeSqlTables != "" {
		excludeDbTables = strings.Split(*excludeSqlTables, ",")
	}

	conf := dbmeta.NewConfig(LoadTemplate)
	initialize(conf)

	err = loadDefaultDBMappings(conf)
	if err != nil {
		fmt.Printf("Error processing default mapping file error: %v\n", err)
		os.Exit(1)
		return
	}

	if *mappingFileName != "" {
		err := dbmeta.LoadMappings(*mappingFileName, *verbose)
		if err != nil {
			fmt.Printf("Error loading mappings file %s error: %v\n", *mappingFileName, err)
			os.Exit(1)
			return
		}
	}

	if *contextFileName != "" {
		loadContextMapping(conf)
	}

	tableInfos = dbmeta.LoadTableInfo(db, dbTables, excludeDbTables, conf)

	if len(tableInfos) == 0 {
		fmt.Printf("No tables loaded\n")
		os.Exit(1)
	}

	fmt.Printf("Generating code for the following tables (%d)\n", len(tableInfos))
	i := 0
	for tableName := range tableInfos {
		fmt.Printf("[%d] %s\n", i, tableName)
		i++
	}

	conf.ContextMap["tableInfos"] = tableInfos

	if *execCustomScript != "" {
		err = executeCustomScript(conf)
		if err != nil {
			fmt.Printf("Error in executing custom script %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
		return
	}

	err = generate(conf)
	if err != nil {
		fmt.Printf("Error in executing generate %v\n", err)
		os.Exit(1)
	}

	os.Exit(0)
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

func initialize(conf *dbmeta.Config) {
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

	conf.SqlType = *sqlType
	conf.SqlDatabase = *sqlDatabase

	conf.AddJSONAnnotation = *AddJSONAnnotation
	conf.AddXMLAnnotation = *AddXMLAnnotation
	conf.AddGormAnnotation = *AddGormAnnotation
	conf.AddProtobufAnnotation = *AddProtobufAnnotation
	conf.AddDBAnnotation = *AddDBAnnotation
	conf.UseGureguTypes = *UseGureguTypes
	conf.JsonNameFormat = *jsonNameFormat
	conf.XMLNameFormat = *xmlNameFormat
	conf.ProtobufNameFormat = *protoNameFormat
	conf.Verbose = *verbose
	conf.OutDir = *outDir
	conf.Overwrite = *overwrite

	conf.SqlConnStr = *sqlConnStr
	conf.ServerPort = *serverPort
	conf.ServerHost = *serverHost
	conf.Overwrite = *overwrite

	conf.Module = *module
	conf.ModelPackageName = *modelPackageName
	conf.ModelFQPN = *module + "/" + *modelPackageName

	conf.DaoPackageName = *daoPackageName
	conf.DaoFQPN = *module + "/" + *daoPackageName

	conf.ApiPackageName = *apiPackageName
	conf.ApiFQPN = *module + "/" + *apiPackageName

	conf.GrpcPackageName = *grpcPackageName
	conf.GrpcFQPN = *module + "/" + *grpcPackageName

	conf.FileNamingTemplate = *fileNamingTemplate
	conf.ModelNamingTemplate = *modelNamingTemplate
	conf.FieldNamingTemplate = *fieldNamingTemplate

	conf.Swagger.Version = *swaggerVersion
	conf.Swagger.BasePath = *swaggerBasePath
	conf.Swagger.Title = fmt.Sprintf("Sample CRUD api for %s db", *sqlDatabase)
	conf.Swagger.Description = fmt.Sprintf("Sample CRUD api for %s db", *sqlDatabase)
	conf.Swagger.TOS = *swaggerTos
	conf.Swagger.ContactName = *swaggerContactName
	conf.Swagger.ContactURL = *swaggerContactURL
	conf.Swagger.ContactEmail = *swaggerContactEmail
	conf.Swagger.Host = fmt.Sprintf("%s:%d", *serverHost, *serverPort)

	conf.JsonNameFormat = strings.ToLower(conf.JsonNameFormat)
	conf.XMLNameFormat = strings.ToLower(conf.XMLNameFormat)
	conf.ProtobufNameFormat = strings.ToLower(conf.ProtobufNameFormat)
}

func loadDefaultDBMappings(conf *dbmeta.Config) error {
	var err error
	var content []byte
	content, err = baseTemplates.Find("mapping.json")
	if err != nil {
		return err
	}

	err = dbmeta.ProcessMappings("internal", content, conf.Verbose)
	if err != nil {
		return err
	}
	return nil
}

func executeCustomScript(conf *dbmeta.Config) error {
	fmt.Printf("Executing script %s\n", *execCustomScript)

	b, err := ioutil.ReadFile(*execCustomScript)
	if err != nil {
		fmt.Printf("Error Loading exec script: %s, error: %v\n", *execCustomScript, err)
		return err
	}
	content := string(b)
	data := map[string]interface{}{}
	err = execTemplate(conf, "exec", content, data)
	if err != nil {
		fmt.Printf("Error Loading exec script: %s, error: %v\n", *execCustomScript, err)
		return err
	}

	return nil
}

func execTemplate(conf *dbmeta.Config, name, templateStr string, data map[string]interface{}) error {

	data["DatabaseName"] = *sqlDatabase
	data["module"] = *module
	data["modelFQPN"] = conf.ModelFQPN
	data["daoFQPN"] = conf.DaoFQPN
	data["apiFQPN"] = conf.ApiFQPN
	data["modelPackageName"] = *modelPackageName
	data["daoPackageName"] = *daoPackageName
	data["apiPackageName"] = *apiPackageName
	data["sqlType"] = *sqlType
	data["sqlConnStr"] = *sqlConnStr
	data["serverPort"] = *serverPort
	data["serverHost"] = *serverHost
	data["SwaggerInfo"] = conf.Swagger
	data["tableInfos"] = tableInfos
	data["CommandLine"] = conf.CmdLine
	data["outDir"] = *outDir
	data["Config"] = conf

	rt, err := conf.GetTemplate(name, templateStr)
	if err != nil {
		fmt.Printf("Error in loading %s template, error: %v\n", name, err)
		return err
	}
	var buf bytes.Buffer
	err = rt.Execute(&buf, data)
	if err != nil {
		fmt.Printf("Error in rendering %s: %s\n", name, err.Error())
		return err
	}

	fmt.Printf("%s\n", buf.String())
	return nil
}

func generate(conf *dbmeta.Config) error {
	var err error

	*jsonNameFormat = strings.ToLower(*jsonNameFormat)
	*xmlNameFormat = strings.ToLower(*xmlNameFormat)
	modelDir := filepath.Join(*outDir, *modelPackageName)
	apiDir := filepath.Join(*outDir, *apiPackageName)
	daoDir := filepath.Join(*outDir, *daoPackageName)

	err = os.MkdirAll(*outDir, 0777)
	if err != nil && !*overwrite {
		fmt.Printf("unable to create outDir: %s error: %v\n", *outDir, err)
		return err
	}

	err = os.MkdirAll(modelDir, 0777)
	if err != nil && !*overwrite {
		fmt.Printf("unable to create modelDir: %s error: %v\n", modelDir, err)
		return err
	}

	if *daoGenerate {
		err = os.MkdirAll(daoDir, 0777)
		if err != nil && !*overwrite {
			fmt.Printf("unable to create daoDir: %s error: %v\n", daoDir, err)
			return err
		}
	}

	if *restAPIGenerate {
		err = os.MkdirAll(apiDir, 0777)
		if err != nil && !*overwrite {
			fmt.Printf("unable to create apiDir: %s error: %v\n", apiDir, err)
			return err
		}
	}
	var ModelTmpl string
	var ModelBaseTmpl string
	var ControllerTmpl string
	var DaoTmpl string
	var DaoFileName string

	var DaoInitTmpl string
	var GoModuleTmpl string

	if ControllerTmpl, err = LoadTemplate("api.go.tmpl"); err != nil {
		fmt.Printf("Error loading template %v\n", err)
		return err
	}

	if *AddGormAnnotation {
		DaoFileName = "dao_gorm.go.tmpl"
		if DaoTmpl, err = LoadTemplate(DaoFileName); err != nil {
			fmt.Printf("Error loading template %v\n", err)
			return err
		}
		if DaoInitTmpl, err = LoadTemplate("dao_gorm_init.go.tmpl"); err != nil {
			fmt.Printf("Error loading template %v\n", err)
			return err
		}
	} else {
		DaoFileName = "dao_sqlx.go.tmpl"
		if DaoTmpl, err = LoadTemplate(DaoFileName); err != nil {
			fmt.Printf("Error loading template %v\n", err)
			return err
		}
		if DaoInitTmpl, err = LoadTemplate("dao_sqlx_init.go.tmpl"); err != nil {
			fmt.Printf("Error loading template %v\n", err)
			return err
		}
	}

	if GoModuleTmpl, err = LoadTemplate("gomod.tmpl"); err != nil {
		fmt.Printf("Error loading template %v\n", err)
		return err
	}

	if ModelTmpl, err = LoadTemplate("model.go.tmpl"); err != nil {
		fmt.Printf("Error loading template %v\n", err)
		return err
	}
	if ModelBaseTmpl, err = LoadTemplate("model_base.go.tmpl"); err != nil {
		fmt.Printf("Error loading template %v\n", err)
		return err
	}

	*jsonNameFormat = strings.ToLower(*jsonNameFormat)
	*xmlNameFormat = strings.ToLower(*xmlNameFormat)

	// generate go files for each table
	for tableName, tableInfo := range tableInfos {

		if len(tableInfo.Fields) == 0 {
			if *verbose {
				fmt.Printf("[%d] Table: %s - No Fields Available\n", tableInfo.Index, tableName)
			}

			continue
		}

		modelInfo := conf.CreateContextForTableFile(tableInfo)

		modelFile := filepath.Join(modelDir, CreateGoSrcFileName(tableName))
		conf.WriteTemplate("model.go.tmpl", ModelTmpl, modelInfo, modelFile, true)

		if *restAPIGenerate {
			restFile := filepath.Join(apiDir, CreateGoSrcFileName(tableName))
			conf.WriteTemplate("api.go.tmpl", ControllerTmpl, modelInfo, restFile, true)
		}

		if *daoGenerate {
			//write dao
			outputFile := filepath.Join(daoDir, CreateGoSrcFileName(tableName))
			conf.WriteTemplate(DaoFileName, DaoTmpl, modelInfo, outputFile, true)
		}
	}

	data := map[string]interface{}{}

	if *restAPIGenerate {
		if err = generateRestBaseFiles(conf, apiDir); err != nil {
			return err
		}
	}

	if *daoGenerate {
		conf.WriteTemplate("daoBase", DaoInitTmpl, data, filepath.Join(daoDir, "dao_base.go"), true)
	}

	conf.WriteTemplate("modelBase", ModelBaseTmpl, data, filepath.Join(modelDir, "model_base.go"), true)

	if *modGenerate {
		conf.WriteTemplate("go.mod", GoModuleTmpl, data, filepath.Join(*outDir, "go.mod"), false)
	}

	if *makefileGenerate {
		if err = generateMakefile(conf); err != nil {
			return err
		}
	}

	if *AddProtobufAnnotation {
		if err = generateProtobufDefinitionFile(conf, data); err != nil {
			return err
		}
	}

	data = map[string]interface{}{
		"deps":        "go list -f '{{ join .Deps  \"\\n\"}}' .",
		"CommandLine": conf.CmdLine,
		"Config":      conf,
	}

	if *projectGenerate {
		if err = generateProjectFiles(conf, data); err != nil {
			return err
		}
	}

	if *serverGenerate {
		if err = generateServerCode(conf); err != nil {
			return err
		}
	}

	if *copyTemplates {
		if err = copyTemplatesToTarget(); err != nil {
			return err
		}
	}

	return nil
}

func generateRestBaseFiles(conf *dbmeta.Config, apiDir string) (err error) {

	data := map[string]interface{}{}
	var RouterTmpl string
	var HTTPUtilsTmpl string

	if HTTPUtilsTmpl, err = LoadTemplate("http_utils.go.tmpl"); err != nil {
		fmt.Printf("Error loading template %v\n", err)
		return
	}
	if RouterTmpl, err = LoadTemplate("router.go.tmpl"); err != nil {
		fmt.Printf("Error loading template %v\n", err)
		return
	}

	conf.WriteTemplate("router", RouterTmpl, data, filepath.Join(apiDir, "router.go"), true)
	conf.WriteTemplate("example server", HTTPUtilsTmpl, data, filepath.Join(apiDir, "http_utils.go"), true)
	return nil
}

func generateMakefile(conf *dbmeta.Config) (err error) {
	var MakefileTmpl string

	if MakefileTmpl, err = LoadTemplate("Makefile.tmpl"); err != nil {
		fmt.Printf("Error loading template %v\n", err)
		return
	}

	data := map[string]interface{}{
		"deps":             "go list -f '{{ join .Deps  \"\\n\"}}' .",
		"RegenCmdLineArgs": regenCmdLine(),
		"RegenCmdLine":     strings.Join(regenCmdLine(), " \\\n    "),
	}

	if *AddProtobufAnnotation {
		populateProtoCinContext(conf, data)
	}

	conf.WriteTemplate("makefile", MakefileTmpl, data, filepath.Join(*outDir, "Makefile"), false)
	return nil
}

func generateProtobufDefinitionFile(conf *dbmeta.Config, data map[string]interface{}) (err error) {

	moduleDir := filepath.Join(*outDir, conf.ModelPackageName)
	serverDir := filepath.Join(*outDir, conf.GrpcPackageName)
	err = os.MkdirAll(serverDir, 0777)
	if err != nil {
		fmt.Printf("unable to create serverDir: %s error: %v\n", serverDir, err)
		return
	}

	var ProtobufTmpl string

	if ProtobufTmpl, err = LoadTemplate("protobuf.tmpl"); err != nil {
		fmt.Printf("Error loading template %v\n", err)
		return err
	}

	protofile := fmt.Sprintf("%s.proto", *sqlDatabase)
	conf.WriteTemplate("protobuf", ProtobufTmpl, data, filepath.Join(*outDir, protofile), false)

	compileOutput, err := CompileProtoC(*outDir, moduleDir, filepath.Join(*outDir, protofile))
	if err != nil {
		fmt.Printf("Error compiling proto file %v\n", err)
		return err
	}
	fmt.Printf("----------------------------\n")
	fmt.Printf("protoc: %s\n", compileOutput)
	fmt.Printf("----------------------------\n")
	// protoc -I./  --go_out=plugins=grpc:./   ./dvdrental.proto

	if ProtobufTmpl, err = LoadTemplate("protomain.go.tmpl"); err != nil {
		fmt.Printf("Error loading template %v\n", err)
		return err
	}

	conf.WriteTemplate("protobuf", ProtobufTmpl, data, filepath.Join(serverDir, "main.go"), true)

	if ProtobufTmpl, err = LoadTemplate("protoserver.go.tmpl"); err != nil {
		fmt.Printf("Error loading template %v\n", err)
		return err
	}

	conf.WriteTemplate("protobuf", ProtobufTmpl, data, filepath.Join(serverDir, "protoserver.go"), true)

	return nil
}

func createProtocCmdLine(protoBufDir, protoBufOutDir, protoBufFile string) ([]string, error) {

	if *gogoProtoImport != "" {
		if !dbmeta.Exists(*gogoProtoImport) {
			fmt.Printf("%s does not exist on path - install with\ngo get -u github.com/gogo/protobuf/proto\n\n", *gogoProtoImport)
			return nil, fmt.Errorf("supplied gogo proto location does  not exist")
		}
	}

	usr, err := user.Current()
	if err == nil {
		dir := usr.HomeDir
		srcPath := filepath.Join(dir, "go/src")

		//srcDirExists := dbmeta.Exists(srcPath)

		gogoPath := filepath.Join(dir, "go/src/github.com/gogo/protobuf/gogoproto/gogo.proto")
		gogoImportExists := dbmeta.Exists(gogoPath)

		//fmt.Printf("path    : %s   srcDirExists: %t\n", srcPath, srcDirExists)
		//fmt.Printf("gogoPath: %s   gogoImportExists: %t\n", gogoPath, gogoImportExists)

		if !gogoImportExists {
			fmt.Printf("github.com/gogo/protobuf/gogoproto/gogo.proto does not exist on path - install with\ngo get -u github.com/gogo/protobuf/proto\n\n")
			return nil, fmt.Errorf("github.com/gogo/protobuf/gogoproto/gogo.proto does not exist")
		}

		*gogoProtoImport = srcPath
	}

	fmt.Printf("----------------------------\n")

	args := []string{fmt.Sprintf("-I%s", *gogoProtoImport),
		fmt.Sprintf("-I%s", protoBufDir),

		fmt.Sprintf("--gogo_out=plugins=grpc,Mgoogle/protobuf/timestamp.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/duration.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/empty.proto=github.com/gogo/protobuf/types,Mgoogle/api/annotations.proto=github.com/gogo/googleapis/google/api,Mmodel.proto:%s", protoBufOutDir),
		//fmt.Sprintf("--gogo_out=plugins=grpc:%s", protoBufOutDir),
		fmt.Sprintf("%s", protoBufFile)}

	return args, nil
}
func CompileProtoC(protoBufDir, protoBufOutDir, protoBufFile string) (string, error) {
	args, err := createProtocCmdLine(protoBufDir, protoBufOutDir, protoBufFile)
	if err != nil {
		return "", err
	}

	cmd := exec.Command("protoc", args...)

	cmdLineArgs := strings.Join(args, " ")
	fmt.Printf("protoc %s\n", cmdLineArgs)

	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("error calling protoc: %T %v\n", err, err)
		fmt.Printf("%s\n", stdoutStderr)
		return "", err
	}

	return string(stdoutStderr), nil
}

func generateProjectFiles(conf *dbmeta.Config, data map[string]interface{}) (err error) {

	var GitIgnoreTmpl string
	if GitIgnoreTmpl, err = LoadTemplate("gitignore.tmpl"); err != nil {
		fmt.Printf("Error loading template %v\n", err)
		return
	}
	var ReadMeTmpl string
	if ReadMeTmpl, err = LoadTemplate("README.md.tmpl"); err != nil {
		fmt.Printf("Error loading template %v\n", err)
		return
	}
	populateProtoCinContext(conf, data)
	conf.WriteTemplate("gitignore", GitIgnoreTmpl, data, filepath.Join(*outDir, ".gitignore"), false)
	conf.WriteTemplate("readme", ReadMeTmpl, data, filepath.Join(*outDir, "README.md"), false)
	return nil
}

func populateProtoCinContext(conf *dbmeta.Config, data map[string]interface{})  {
	protofile := fmt.Sprintf("%s.proto", *sqlDatabase)
	moduleDir := filepath.Join(*outDir, conf.ModelPackageName)
	protocCmdLineArgs, err := createProtocCmdLine(*outDir, moduleDir, filepath.Join(*outDir, protofile))
	if err != nil {
		protoC := []string{"gen"}
		protoC = append(protoC, protocCmdLineArgs...)

		data["ProtocCmdLineArgs"] = protoC
		data["ProtocCmdLine"] = strings.Join(protoC, " \\\n    ")
	}
}
func generateServerCode(conf *dbmeta.Config) (err error) {
	data := map[string]interface{}{}
	var MainServerTmpl string

	if *AddGormAnnotation {
		if MainServerTmpl, err = LoadTemplate("main_gorm.go.tmpl"); err != nil {
			fmt.Printf("Error loading template %v\n", err)
			return
		}
	} else {
		if MainServerTmpl, err = LoadTemplate("main_sqlx.go.tmpl"); err != nil {
			fmt.Printf("Error loading template %v\n", err)
			return
		}
	}

	serverDir := filepath.Join(*outDir, "app/server")
	err = os.MkdirAll(serverDir, 0777)
	if err != nil {
		fmt.Printf("unable to create serverDir: %s error: %v\n", serverDir, err)
		return
	}
	conf.WriteTemplate("example server", MainServerTmpl, data, filepath.Join(serverDir, "main.go"), true)
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

func regenCmdLine() []string {
	cmdLine := []string{"gen",
		fmt.Sprintf(" --sqltype=%s", *sqlType),
		fmt.Sprintf(" --connstr='%s'", *sqlConnStr),
		fmt.Sprintf(" --database=%s", *sqlDatabase),
		fmt.Sprintf(" --templateDir=%s", "./templates"),
	}

	if *sqlTable != "" {
		cmdLine = append(cmdLine, fmt.Sprintf(" --table=%s", *sqlTable))
	}

	if *excludeSqlTables != "" {
		cmdLine = append(cmdLine, fmt.Sprintf(" --exclude=%s", *excludeSqlTables))
	}

	cmdLine = append(cmdLine, fmt.Sprintf(" --model=%s", *modelPackageName))
	cmdLine = append(cmdLine, fmt.Sprintf(" --dao=%s", *daoPackageName))
	cmdLine = append(cmdLine, fmt.Sprintf(" --api=%s", *apiPackageName))
	cmdLine = append(cmdLine, fmt.Sprintf(" --out=%s", "./"))
	cmdLine = append(cmdLine, fmt.Sprintf(" --module=%s", *module))
	if *AddJSONAnnotation {
		cmdLine = append(cmdLine, fmt.Sprintf(" --json"))
		cmdLine = append(cmdLine, fmt.Sprintf(" --json-fmt=%s", *jsonNameFormat))
	}
	if *AddXMLAnnotation {
		cmdLine = append(cmdLine, fmt.Sprintf(" --xml"))
		cmdLine = append(cmdLine, fmt.Sprintf(" --xml-fmt=%s", *xmlNameFormat))
	}
	if *AddGormAnnotation {
		cmdLine = append(cmdLine, fmt.Sprintf(" --gorm"))
	}
	if *AddProtobufAnnotation {
		cmdLine = append(cmdLine, fmt.Sprintf(" --protobuf"))
		cmdLine = append(cmdLine, fmt.Sprintf(" --proto-fmt=%s", *protoNameFormat))
	}
	if *AddDBAnnotation {
		cmdLine = append(cmdLine, fmt.Sprintf(" --db"))
	}
	if *UseGureguTypes {
		cmdLine = append(cmdLine, fmt.Sprintf(" --guregu"))
	}
	if *modGenerate {
		cmdLine = append(cmdLine, fmt.Sprintf(" --mod"))
	}
	if *makefileGenerate {
		cmdLine = append(cmdLine, fmt.Sprintf(" --makefile"))
	}
	if *serverGenerate {
		cmdLine = append(cmdLine, fmt.Sprintf(" --server"))
	}
	if *overwrite {
		cmdLine = append(cmdLine, fmt.Sprintf(" --overwrite"))
	}

	if *contextFileName != "" {
		cmdLine = append(cmdLine, fmt.Sprintf(" --context=%s", *contextFileName))
	}

	cmdLine = append(cmdLine, fmt.Sprintf(" --host=%s", *serverHost))
	cmdLine = append(cmdLine, fmt.Sprintf(" --port=%d", *serverPort))
	if *restAPIGenerate {
		cmdLine = append(cmdLine, fmt.Sprintf(" --rest"))
	}

	if *daoGenerate {
		cmdLine = append(cmdLine, fmt.Sprintf(" --generate-dao"))
	}
	if *projectGenerate {
		cmdLine = append(cmdLine, fmt.Sprintf(" --generate-proj"))
	}

	cmdLine = append(cmdLine, fmt.Sprintf(" --file_naming='%s'", *fileNamingTemplate))
	cmdLine = append(cmdLine, fmt.Sprintf(" --model_naming='%s'", *modelNamingTemplate))

	if *verbose {
		cmdLine = append(cmdLine, fmt.Sprintf(" --verbose"))
	}

	cmdLine = append(cmdLine, fmt.Sprintf(" --swagger_version=%s", *swaggerVersion))
	cmdLine = append(cmdLine, fmt.Sprintf(" --swagger_path=%s", *swaggerBasePath))
	cmdLine = append(cmdLine, fmt.Sprintf(" --swagger_tos=%s", *swaggerTos))
	cmdLine = append(cmdLine, fmt.Sprintf(" --swagger_contact_name=%s", *swaggerContactName))
	cmdLine = append(cmdLine, fmt.Sprintf(" --swagger_contact_url=%s", *swaggerContactURL))
	cmdLine = append(cmdLine, fmt.Sprintf(" --swagger_contact_email=%s", *swaggerContactEmail))
	return cmdLine
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

// CreateGoSrcFileName ensures name doesnt clash with go naming conventions like _test.go
func CreateGoSrcFileName(tableName string) string {
	name := dbmeta.Replace(*fileNamingTemplate, tableName)
	// name := inflection.Singular(tableName)

	if strings.HasSuffix(name, "_test") {
		name = name[0 : len(name)-5]
		name = name + "_tst"
	}
	return name + ".go"
}

func LoadTemplate(filename string) (content string, err error) {
	if *templateDir != "" {
		fpath := filepath.Join(*templateDir, filename)
		var b []byte
		b, err = ioutil.ReadFile(fpath)
		if err == nil {
			// fmt.Printf("Loaded template from file: %s\n", fpath)
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

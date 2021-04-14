package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"go/build"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	_ "github.com/denisenkom/go-mssqldb"
	"github.com/droundy/goopt"
	"github.com/gobuffalo/packd"
	"github.com/gobuffalo/packr/v2"
	"github.com/jimsmart/schema"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/lib/pq"
	"github.com/logrusorgru/aurora"
	_ "github.com/mattn/go-sqlite3"

	"github.com/smallnest/gen/dbmeta"
)

var (
	sqlType          = goopt.String([]string{"--sqltype"}, "mysql", "sql database type such as [ mysql, mssql, postgres, sqlite, etc. ]")
	sqlConnStr       = goopt.String([]string{"-c", "--connstr"}, "nil", "database connection string")
	sqlDatabase      = goopt.String([]string{"-d", "--database"}, "nil", "Database to for connection")
	sqlTable         = goopt.String([]string{"-t", "--table"}, "", "Table to build struct from")
	excludeSQLTables = goopt.String([]string{"-x", "--exclude"}, "", "Table(s) to exclude")
	templateDir      = goopt.String([]string{"--templateDir"}, "", "Template Dir")
	fragmentsDir     = goopt.String([]string{"--fragmentsDir"}, "", "Code fragments Dir")
	saveTemplateDir  = goopt.String([]string{"--save"}, "", "Save templates to dir")

	modelPackageName    = goopt.String([]string{"--model"}, "model", "name to set for model package")
	modelNamingTemplate = goopt.String([]string{"--model_naming"}, "{{FmtFieldName .}}", "model naming template to name structs")
	fieldNamingTemplate = goopt.String([]string{"--field_naming"}, "{{FmtFieldName (stringifyFirstChar .) }}", "field naming template to name structs")
	fileNamingTemplate  = goopt.String([]string{"--file_naming"}, "{{.}}", "file_naming template to name files")

	daoPackageName  = goopt.String([]string{"--dao"}, "dao", "name to set for dao package")
	apiPackageName  = goopt.String([]string{"--api"}, "api", "name to set for api package")
	grpcPackageName = goopt.String([]string{"--grpc"}, "grpc", "name to set for grpc package")
	outDir          = goopt.String([]string{"--out"}, ".", "output dir")
	module          = goopt.String([]string{"--module"}, "example.com/example", "module path")
	overwrite       = goopt.Flag([]string{"--overwrite"}, []string{"--no-overwrite"}, "Overwrite existing files (default)", "disable overwriting files")
	windows         = goopt.Flag([]string{"--windows"}, []string{}, "use windows line endings in generated files", "")
	noColorOutput   = goopt.Flag([]string{"--no-color"}, []string{}, "disable color output", "")

	contextFileName  = goopt.String([]string{"--context"}, "", "context file (json) to populate context with")
	mappingFileName  = goopt.String([]string{"--mapping"}, "", "mapping file (json) to map sql types to golang/protobuf etc")
	execCustomScript = goopt.String([]string{"--exec"}, "", "execute script for custom code generation")

	addJSONAnnotation = goopt.Flag([]string{"--json"}, []string{"--no-json"}, "Add json annotations (default)", "Disable json annotations")
	jsonNameFormat    = goopt.String([]string{"--json-fmt"}, "snake", "json name format [snake | camel | lower_camel | none]")

	addXMLAnnotation = goopt.Flag([]string{"--xml"}, []string{"--no-xml"}, "Add xml annotations (default)", "Disable xml annotations")
	xmlNameFormat    = goopt.String([]string{"--xml-fmt"}, "snake", "xml name format [snake | camel | lower_camel | none]")

	addGormAnnotation     = goopt.Flag([]string{"--gorm"}, []string{}, "Add gorm annotations (tags)", "")
	addProtobufAnnotation = goopt.Flag([]string{"--protobuf"}, []string{}, "Add protobuf annotations (tags)", "")
	protoNameFormat       = goopt.String([]string{"--proto-fmt"}, "snake", "proto name format [snake | camel | lower_camel | none]")
	gogoProtoImport       = goopt.String([]string{"--gogo-proto"}, "", "location of gogo import ")

	addDBAnnotation = goopt.Flag([]string{"--db"}, []string{}, "Add db annotations (tags)", "")
	useGureguTypes  = goopt.Flag([]string{"--guregu"}, []string{}, "Add guregu null types", "")

	copyTemplates    = goopt.Flag([]string{"--copy-templates"}, []string{}, "Copy regeneration templates to project directory", "")
	modGenerate      = goopt.Flag([]string{"--mod"}, []string{}, "Generate go.mod in output dir", "")
	makefileGenerate = goopt.Flag([]string{"--makefile"}, []string{}, "Generate Makefile in output dir", "")
	serverGenerate   = goopt.Flag([]string{"--server"}, []string{}, "Generate server app output dir", "")
	daoGenerate      = goopt.Flag([]string{"--generate-dao"}, []string{}, "Generate dao functions", "")
	projectGenerate  = goopt.Flag([]string{"--generate-proj"}, []string{}, "Generate project readme and gitignore", "")
	restAPIGenerate  = goopt.Flag([]string{"--rest"}, []string{}, "Enable generating RESTful api", "")
	runGoFmt         = goopt.Flag([]string{"--run-gofmt"}, []string{}, "run gofmt on output dir", "")

	serverListen        = goopt.String([]string{"--listen"}, "", "listen address e.g. :8080")
	serverScheme        = goopt.String([]string{"--scheme"}, "http", "scheme for server url")
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
	au            aurora.Aurora
)

func init() {
	// Setup goopts
	goopt.Description = func() string {
		return "ORM and RESTful API generator for SQl databases"
	}
	goopt.Version = "v0.9.27 (08/04/2020)"
	goopt.Summary = `gen [-v] --sqltype=mysql --connstr "user:password@/dbname" --database <databaseName> --module=example.com/example [--json] [--gorm] [--guregu] [--generate-dao] [--generate-proj]
git fetch up
           sqltype - sql database type such as [ mysql, mssql, postgres, sqlite, etc. ]

`

	// Parse options
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

func loadContextMapping(conf *dbmeta.Config) error {
	contextFile, err := os.Open(*contextFileName)
	if err != nil {
		return err
	}

	defer contextFile.Close()
	jsonParser := json.NewDecoder(contextFile)

	err = jsonParser.Decode(&conf.ContextMap)
	if err != nil {
		return err
	}

	fmt.Printf("Loaded Context from %s with %d defaults\n", *contextFileName, len(conf.ContextMap))
	for key, value := range conf.ContextMap {
		fmt.Printf("    Context:%s -> %v\n", key, value)
	}
	return nil
}

func main() {
	//for i, arg := range os.Args {
	//	fmt.Printf("[%2d] %s\n", i, arg)
	//}
	au = aurora.NewAurora(!*noColorOutput)
	dbmeta.InitColorOutput(au)

	baseTemplates = packr.New("gen", "./template")

	if *saveTemplateDir != "" {
		saveTemplates()
		return
	}

	if *serverListen == "" {
		*serverListen = fmt.Sprintf(":%d", *serverPort)
	}
	if *serverScheme != "http" && *serverScheme != "https" {
		*serverScheme = "http"
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

	// fmt.Printf("fieldNamingTemplate: %s\n", *fieldNamingTemplate)
	// fmt.Printf("fileNamingTemplate: %s\n", *fileNamingTemplate)
	// fmt.Printf("modelNamingTemplate: %s\n", *modelNamingTemplate)

	// Username is required
	if sqlConnStr == nil || *sqlConnStr == "" || *sqlConnStr == "nil" {
		fmt.Print(au.Red("sql connection string is required! Add it with --connstr=s\n\n"))
		fmt.Println(goopt.Usage())
		return
	}

	if sqlDatabase == nil || *sqlDatabase == "" || *sqlDatabase == "nil" {
		fmt.Print(au.Red("Database can not be null\n\n"))
		fmt.Println(goopt.Usage())
		return
	}

	db, err := initializeDB()
	if err != nil {
		fmt.Print(au.Red(fmt.Sprintf("Error in initializing db %v\n", err)))
		os.Exit(1)
		return
	}

	defer db.Close()

	var dbTables []string
	// parse or read tables
	if *sqlTable != "" {
		dbTables = strings.Split(*sqlTable, ",")
	} else {
		schemaTables, err := schema.TableNames(db)
		if err != nil {
			fmt.Print(au.Red(fmt.Sprintf("Error in fetching tables information from %s information schema from %s\n", *sqlType, *sqlConnStr)))
			os.Exit(1)
			return
		}
		for _, st := range schemaTables {
			dbTables = append(dbTables, st[1]) // s[0] == sqlDatabase
		}
	}

	if strings.HasPrefix(*modelNamingTemplate, "'") && strings.HasSuffix(*modelNamingTemplate, "'") {
		*modelNamingTemplate = strings.TrimSuffix(*modelNamingTemplate, "'")
		*modelNamingTemplate = strings.TrimPrefix(*modelNamingTemplate, "'")
	}

	var excludeDbTables []string

	if *excludeSQLTables != "" {
		excludeDbTables = strings.Split(*excludeSQLTables, ",")
	}

	conf := dbmeta.NewConfig(LoadTemplate)
	initialize(conf)

	err = loadDefaultDBMappings(conf)
	if err != nil {
		fmt.Print(au.Red(fmt.Sprintf("Error processing default mapping file error: %v\n", err)))
		os.Exit(1)
		return
	}

	if *mappingFileName != "" {
		err := dbmeta.LoadMappings(*mappingFileName, *verbose)
		if err != nil {
			fmt.Print(au.Red(fmt.Sprintf("Error loading mappings file %s error: %v\n", *mappingFileName, err)))
			os.Exit(1)
			return
		}
	}

	if *contextFileName != "" {
		err = loadContextMapping(conf)
		if err != nil {
			fmt.Print(au.Red(fmt.Sprintf("Error loading context file %s error: %v\n", *contextFileName, err)))
			os.Exit(1)
			return
		}
	}

	tableInfos = dbmeta.LoadTableInfo(db, dbTables, excludeDbTables, conf)

	if len(tableInfos) == 0 {
		fmt.Print(au.Red(fmt.Sprintf("No tables loaded\n")))
		os.Exit(1)
	}

	fmt.Printf("Generating code for the following tables (%d)\n", len(tableInfos))
	i := 0
	for tableName := range tableInfos {
		fmt.Printf("[%d] %s\n", i, tableName)
		i++
	}

	conf.TableInfos = tableInfos
	conf.ContextMap["tableInfos"] = tableInfos

	if *execCustomScript != "" {
		err = executeCustomScript(conf)
		if err != nil {
			fmt.Print(au.Red(fmt.Sprintf("Error in executing custom script %v\n", err)))
			os.Exit(1)
		}
		os.Exit(0)
		return
	}

	if *verbose {
		listTemplates()
	}

	err = generate(conf)
	if err != nil {
		fmt.Print(au.Red(fmt.Sprintf("Error in executing generate %v\n", err)))
		os.Exit(1)
	}

	os.Exit(0)
}

func initializeDB() (db *sql.DB, err error) {
	db, err = sql.Open(*sqlType, *sqlConnStr)
	if err != nil {
		fmt.Print(au.Red(fmt.Sprintf("Error in open database: %v\n\n", err.Error())))
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		fmt.Print(au.Red(fmt.Sprintf("Error pinging database: %v\n\n", err.Error())))
		return
	}

	return
}

func initialize(conf *dbmeta.Config) {
	if outDir == nil || *outDir == "" {
		*outDir = "."
	}

	// load fragments if specified
	if fragmentsDir != nil && *fragmentsDir != "" {
		conf.LoadFragments(*fragmentsDir)
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

	conf.SQLType = *sqlType
	conf.SQLDatabase = *sqlDatabase

	conf.AddJSONAnnotation = *addJSONAnnotation
	conf.AddXMLAnnotation = *addXMLAnnotation
	conf.AddGormAnnotation = *addGormAnnotation
	conf.AddProtobufAnnotation = *addProtobufAnnotation
	conf.AddDBAnnotation = *addDBAnnotation
	conf.UseGureguTypes = *useGureguTypes
	conf.JSONNameFormat = *jsonNameFormat
	conf.XMLNameFormat = *xmlNameFormat
	conf.ProtobufNameFormat = *protoNameFormat
	conf.Verbose = *verbose
	conf.OutDir = *outDir
	conf.Overwrite = *overwrite
	conf.LineEndingCRLF = *windows

	conf.SQLConnStr = *sqlConnStr
	conf.ServerPort = *serverPort
	conf.ServerHost = *serverHost
	conf.ServerScheme = *serverScheme
	conf.ServerListen = *serverListen

	conf.Module = *module
	conf.ModelPackageName = *modelPackageName
	conf.ModelFQPN = *module + "/" + *modelPackageName

	conf.DaoPackageName = *daoPackageName
	conf.DaoFQPN = *module + "/" + *daoPackageName

	conf.APIPackageName = *apiPackageName
	conf.APIFQPN = *module + "/" + *apiPackageName

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
	if *serverPort == 80 {
		conf.Swagger.Host = *serverHost
	} else {
		conf.Swagger.Host = fmt.Sprintf("%s:%d", *serverHost, *serverPort)
	}

	conf.JSONNameFormat = strings.ToLower(conf.JSONNameFormat)
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
	data := map[string]interface{}{}

	absPath, err := filepath.Abs(*execCustomScript)
	if err != nil {
		absPath = *execCustomScript
	}

	tpl := &dbmeta.GenTemplate{Name: absPath, Content: string(b)}

	err = execTemplate(conf, tpl, data)
	if err != nil {
		fmt.Printf("Error Loading exec script: %s, error: %v\n", *execCustomScript, err)
		return err
	}

	return nil
}

func execTemplate(conf *dbmeta.Config, genTemplate *dbmeta.GenTemplate, data map[string]interface{}) error {
	data["DatabaseName"] = *sqlDatabase
	data["module"] = *module
	data["modelFQPN"] = conf.ModelFQPN
	data["daoFQPN"] = conf.DaoFQPN
	data["apiFQPN"] = conf.APIFQPN
	data["modelPackageName"] = *modelPackageName
	data["daoPackageName"] = *daoPackageName
	data["apiPackageName"] = *apiPackageName
	data["sqlType"] = *sqlType
	data["sqlConnStr"] = *sqlConnStr
	data["serverPort"] = *serverPort
	data["serverHost"] = *serverHost
	data["serverListen"] = *serverListen
	data["SwaggerInfo"] = conf.Swagger
	data["tableInfos"] = tableInfos
	data["CommandLine"] = conf.CmdLine
	data["outDir"] = *outDir
	data["Config"] = conf

	tables := make([]string, 0, len(tableInfos))
	for k := range tableInfos {
		tables = append(tables, k)
	}

	data["tables"] = tables

	rt, err := conf.GetTemplate(genTemplate)
	if err != nil {
		fmt.Print(au.Red(fmt.Sprintf("Error in loading %s template, error: %v\n", genTemplate.Name, err)))
		return err
	}
	var buf bytes.Buffer
	err = rt.Execute(&buf, data)
	if err != nil {
		fmt.Printf("Error in rendering %s: %s\n", genTemplate.Name, err.Error())
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
		fmt.Print(au.Red(fmt.Sprintf("unable to create outDir: %s error: %v\n", *outDir, err)))
		return err
	}

	err = os.MkdirAll(modelDir, 0777)
	if err != nil && !*overwrite {
		fmt.Print(au.Red(fmt.Sprintf("unable to create modelDir: %s error: %v\n", modelDir, err)))
		return err
	}

	if *daoGenerate {
		err = os.MkdirAll(daoDir, 0777)
		if err != nil && !*overwrite {
			fmt.Print(au.Red(fmt.Sprintf("unable to create daoDir: %s error: %v\n", daoDir, err)))
			return err
		}
	}

	if *restAPIGenerate {
		err = os.MkdirAll(apiDir, 0777)
		if err != nil && !*overwrite {
			fmt.Print(au.Red(fmt.Sprintf("unable to create apiDir: %s error: %v\n", apiDir, err)))
			return err
		}
	}
	var ModelTmpl *dbmeta.GenTemplate
	var ModelBaseTmpl *dbmeta.GenTemplate
	var ControllerTmpl *dbmeta.GenTemplate
	var DaoTmpl *dbmeta.GenTemplate

	var DaoInitTmpl *dbmeta.GenTemplate
	var GoModuleTmpl *dbmeta.GenTemplate

	if ControllerTmpl, err = LoadTemplate("api.go.tmpl"); err != nil {
		fmt.Print(au.Red(fmt.Sprintf("Error loading template %v\n", err)))
		return err
	}

	if *addGormAnnotation {
		if DaoTmpl, err = LoadTemplate("dao_gorm.go.tmpl"); err != nil {
			fmt.Print(au.Red(fmt.Sprintf("Error loading template %v\n", err)))
			return err
		}
		if DaoInitTmpl, err = LoadTemplate("dao_gorm_init.go.tmpl"); err != nil {
			fmt.Print(au.Red(fmt.Sprintf("Error loading template %v\n", err)))
			return err
		}
	} else {
		if DaoTmpl, err = LoadTemplate("dao_sqlx.go.tmpl"); err != nil {
			fmt.Print(au.Red(fmt.Sprintf("Error loading template %v\n", err)))
			return err
		}
		if DaoInitTmpl, err = LoadTemplate("dao_sqlx_init.go.tmpl"); err != nil {
			fmt.Print(au.Red(fmt.Sprintf("Error loading template %v\n", err)))
			return err
		}
	}

	if GoModuleTmpl, err = LoadTemplate("gomod.tmpl"); err != nil {
		fmt.Print(au.Red(fmt.Sprintf("Error loading template %v\n", err)))
		return err
	}

	if ModelTmpl, err = LoadTemplate("model.go.tmpl"); err != nil {
		fmt.Print(au.Red(fmt.Sprintf("Error loading template %v\n", err)))
		return err
	}
	if ModelBaseTmpl, err = LoadTemplate("model_base.go.tmpl"); err != nil {
		fmt.Print(au.Red(fmt.Sprintf("Error loading template %v\n", err)))
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
		err = conf.WriteTemplate(ModelTmpl, modelInfo, modelFile)
		if err != nil {
			fmt.Print(au.Red(fmt.Sprintf("Error writing file: %v\n", err)))
			os.Exit(1)
		}

		if *restAPIGenerate {
			restFile := filepath.Join(apiDir, CreateGoSrcFileName(tableName))
			err = conf.WriteTemplate(ControllerTmpl, modelInfo, restFile)
			if err != nil {
				fmt.Print(au.Red(fmt.Sprintf("Error writing file: %v\n", err)))
				os.Exit(1)
			}

		}

		if *daoGenerate {
			// write dao
			outputFile := filepath.Join(daoDir, CreateGoSrcFileName(tableName))
			err = conf.WriteTemplate(DaoTmpl, modelInfo, outputFile)
			if err != nil {
				fmt.Print(au.Red(fmt.Sprintf("Error writing file: %v\n", err)))
				os.Exit(1)
			}
		}
	}

	data := map[string]interface{}{}

	if *restAPIGenerate {
		if err = generateRestBaseFiles(conf, apiDir); err != nil {
			return err
		}
	}

	if *daoGenerate {
		err = conf.WriteTemplate(DaoInitTmpl, data, filepath.Join(daoDir, "dao_base.go"))
		if err != nil {
			fmt.Print(au.Red(fmt.Sprintf("Error writing file: %v\n", err)))
			os.Exit(1)
		}
	}

	err = conf.WriteTemplate(ModelBaseTmpl, data, filepath.Join(modelDir, "model_base.go"))
	if err != nil {
		fmt.Print(au.Red(fmt.Sprintf("Error writing file: %v\n", err)))
		os.Exit(1)
	}

	if *modGenerate {
		err = conf.WriteTemplate(GoModuleTmpl, data, filepath.Join(*outDir, "go.mod"))
		if err != nil {
			fmt.Print(au.Red(fmt.Sprintf("Error writing file: %v\n", err)))
			os.Exit(1)
		}
	}

	if *makefileGenerate {
		if err = generateMakefile(conf); err != nil {
			return err
		}
	}

	if *addProtobufAnnotation {
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

	if *runGoFmt {
		GoFmt(conf.OutDir)
	}

	return nil
}

func generateRestBaseFiles(conf *dbmeta.Config, apiDir string) (err error) {
	data := map[string]interface{}{}
	var RouterTmpl *dbmeta.GenTemplate
	var HTTPUtilsTmpl *dbmeta.GenTemplate

	if HTTPUtilsTmpl, err = LoadTemplate("http_utils.go.tmpl"); err != nil {
		fmt.Print(au.Red(fmt.Sprintf("Error loading template %v\n", err)))
		return
	}
	if RouterTmpl, err = LoadTemplate("router.go.tmpl"); err != nil {
		fmt.Print(au.Red(fmt.Sprintf("Error loading template %v\n", err)))
		return
	}

	err = conf.WriteTemplate(RouterTmpl, data, filepath.Join(apiDir, "router.go"))
	if err != nil {
		fmt.Print(au.Red(fmt.Sprintf("Error writing file: %v\n", err)))
		os.Exit(1)
	}

	err = conf.WriteTemplate(HTTPUtilsTmpl, data, filepath.Join(apiDir, "http_utils.go"))
	if err != nil {
		fmt.Print(au.Red(fmt.Sprintf("Error writing file: %v\n", err)))
		os.Exit(1)
	}

	return nil
}

func generateMakefile(conf *dbmeta.Config) (err error) {
	var MakefileTmpl *dbmeta.GenTemplate

	if MakefileTmpl, err = LoadTemplate("Makefile.tmpl"); err != nil {
		fmt.Print(au.Red(fmt.Sprintf("Error loading template %v\n", err)))
		return
	}

	data := map[string]interface{}{
		"deps":             "go list -f '{{ join .Deps  \"\\n\"}}' .",
		"RegenCmdLineArgs": regenCmdLine(),
		"RegenCmdLine":     strings.Join(regenCmdLine(), " \\\n    "),
	}

	if *addProtobufAnnotation {
		populateProtoCinContext(conf, data)
	}

	err = conf.WriteTemplate(MakefileTmpl, data, filepath.Join(*outDir, "Makefile"))
	if err != nil {
		fmt.Print(au.Red(fmt.Sprintf("Error writing file: %v\n", err)))
		os.Exit(1)
	}

	return nil
}

func generateProtobufDefinitionFile(conf *dbmeta.Config, data map[string]interface{}) (err error) {
	moduleDir := filepath.Join(*outDir, conf.ModelPackageName)
	serverDir := filepath.Join(*outDir, conf.GrpcPackageName)
	err = os.MkdirAll(serverDir, 0777)
	if err != nil {
		fmt.Print(au.Red(fmt.Sprintf("unable to create serverDir: %s error: %v\n", serverDir, err)))
		return
	}

	var ProtobufTmpl *dbmeta.GenTemplate

	if ProtobufTmpl, err = LoadTemplate("protobuf.tmpl"); err != nil {
		fmt.Print(au.Red(fmt.Sprintf("Error loading template %v\n", err)))
		return err
	}

	protofile := fmt.Sprintf("%s.proto", *sqlDatabase)
	err = conf.WriteTemplate(ProtobufTmpl, data, filepath.Join(*outDir, protofile))
	if err != nil {
		fmt.Print(au.Red(fmt.Sprintf("Error writing file: %v\n", err)))
		os.Exit(1)
	}

	compileOutput, err := CompileProtoC(*outDir, moduleDir, filepath.Join(*outDir, protofile))
	if err != nil {
		fmt.Print(au.Red(fmt.Sprintf("Error compiling proto file %v\n", err)))
		return err
	}
	fmt.Printf("----------------------------\n")
	fmt.Printf("protoc: %s\n", compileOutput)
	fmt.Printf("----------------------------\n")
	// protoc -I./  --go_out=plugins=grpc:./   ./dvdrental.proto

	if ProtobufTmpl, err = LoadTemplate("protomain.go.tmpl"); err != nil {
		fmt.Print(au.Red(fmt.Sprintf("Error loading template %v\n", err)))
		return err
	}

	err = conf.WriteTemplate(ProtobufTmpl, data, filepath.Join(serverDir, "main.go"))
	if err != nil {
		fmt.Print(au.Red(fmt.Sprintf("Error writing file: %v\n", err)))
		os.Exit(1)
	}

	if ProtobufTmpl, err = LoadTemplate("protoserver.go.tmpl"); err != nil {
		fmt.Print(au.Red(fmt.Sprintf("Error loading template %v\n", err)))
		return err
	}

	err = conf.WriteTemplate(ProtobufTmpl, data, filepath.Join(serverDir, "protoserver.go"))
	if err != nil {
		fmt.Print(au.Red(fmt.Sprintf("Error writing file: %v\n", err)))
		os.Exit(1)
	}

	return nil
}

func createProtocCmdLine(protoBufDir, protoBufOutDir, protoBufFile string) ([]string, error) {
	if *gogoProtoImport != "" {
		if !dbmeta.Exists(*gogoProtoImport) {
			fmt.Print(au.Red(fmt.Sprintf("%s does not exist on path - install with\ngo get -u github.com/gogo/protobuf/proto\n\n", *gogoProtoImport)))
			return nil, fmt.Errorf("supplied gogo proto location does  not exist")
		}
	}

	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		gopath = build.Default.GOPATH
	}

	srcPath := filepath.Join(gopath, "src")
	gogoPath := filepath.Join(gopath, "src/github.com/gogo/protobuf/gogoproto/gogo.proto")
	gogoImportExists := dbmeta.Exists(gogoPath)

	// fmt.Printf("path    : %s   srcDirExists: %t\n", srcPath, srcDirExists)
	// fmt.Printf("gogoPath: %s   gogoImportExists: %t\n", gogoPath, gogoImportExists)

	if !gogoImportExists {
		fmt.Print(au.Red("github.com/gogo/protobuf/gogoproto/gogo.proto does not exist on path - install with\ngo get -u github.com/gogo/protobuf/proto\n\n"))
		return nil, fmt.Errorf("github.com/gogo/protobuf/gogoproto/gogo.proto does not exist")
	}

	*gogoProtoImport = srcPath

	fmt.Printf("----------------------------\n")

	args := []string{
		fmt.Sprintf("-I%s", *gogoProtoImport),
		fmt.Sprintf("-I%s", protoBufDir),

		fmt.Sprintf("--gogo_out=plugins=grpc,Mgoogle/protobuf/timestamp.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/duration.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/empty.proto=github.com/gogo/protobuf/types,Mgoogle/api/annotations.proto=github.com/gogo/googleapis/google/api,Mmodel.proto:%s", protoBufOutDir),
		// fmt.Sprintf("--gogo_out=plugins=grpc:%s", protoBufOutDir),
		fmt.Sprintf("%s", protoBufFile),
	}

	return args, nil
}

// CompileProtoC exec protoc for proto file, returns stdout result or error
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
		fmt.Print(au.Red(fmt.Sprintf("error calling protoc: %T %v\n", err, err)))
		fmt.Print(au.Red(fmt.Sprintf("%s\n", stdoutStderr)))
		return "", err
	}

	return string(stdoutStderr), nil
}

// GoFmt exec gofmt for a code dir
func GoFmt(codeDir string) (string, error) {
	args := []string{"-s", "-d", "-w", "-l", codeDir}
	cmd := exec.Command("gofmt", args...)

	cmdLineArgs := strings.Join(args, " ")
	fmt.Printf("gofmt %s\n", cmdLineArgs)

	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Print(au.Red(fmt.Sprintf("error calling protoc: %T %v\n", err, err)))
		fmt.Print(au.Red(fmt.Sprintf("%s\n", stdoutStderr)))
		return "", err
	}

	return string(stdoutStderr), nil
}

func generateProjectFiles(conf *dbmeta.Config, data map[string]interface{}) (err error) {
	var GitIgnoreTmpl *dbmeta.GenTemplate
	if GitIgnoreTmpl, err = LoadTemplate("gitignore.tmpl"); err != nil {
		fmt.Print(au.Red(fmt.Sprintf("Error loading template %v\n", err)))
		return
	}
	var ReadMeTmpl *dbmeta.GenTemplate
	if ReadMeTmpl, err = LoadTemplate("README.md.tmpl"); err != nil {
		fmt.Print(au.Red(fmt.Sprintf("Error loading template %v\n", err)))
		return
	}
	populateProtoCinContext(conf, data)
	err = conf.WriteTemplate(GitIgnoreTmpl, data, filepath.Join(*outDir, ".gitignore"))
	if err != nil {
		fmt.Print(au.Red(fmt.Sprintf("Error writing file: %v\n", err)))
		os.Exit(1)
	}

	err = conf.WriteTemplate(ReadMeTmpl, data, filepath.Join(*outDir, "README.md"))
	if err != nil {
		fmt.Print(au.Red(fmt.Sprintf("Error writing file: %v\n", err)))
		os.Exit(1)
	}

	return nil
}

func populateProtoCinContext(conf *dbmeta.Config, data map[string]interface{}) {
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
	var MainServerTmpl *dbmeta.GenTemplate

	if *addGormAnnotation {
		if MainServerTmpl, err = LoadTemplate("main_gorm.go.tmpl"); err != nil {
			fmt.Print(au.Red(fmt.Sprintf("Error loading template %v\n", err)))
			return
		}
	} else {
		if MainServerTmpl, err = LoadTemplate("main_sqlx.go.tmpl"); err != nil {
			fmt.Print(au.Red(fmt.Sprintf("Error loading template %v\n", err)))
			return
		}
	}

	serverDir := filepath.Join(*outDir, "app/server")
	err = os.MkdirAll(serverDir, 0777)
	if err != nil {
		fmt.Print(au.Red(fmt.Sprintf("unable to create serverDir: %s error: %v\n", serverDir, err)))
		return
	}
	err = conf.WriteTemplate(MainServerTmpl, data, filepath.Join(serverDir, "main.go"))
	if err != nil {
		fmt.Print(au.Red(fmt.Sprintf("Error writing file: %v\n", err)))
		os.Exit(1)
	}

	return nil
}

func copyTemplatesToTarget() (err error) {
	templatesDir := filepath.Join(*outDir, "templates")
	err = os.MkdirAll(templatesDir, 0777)
	if err != nil && !*overwrite {
		fmt.Print(au.Red(fmt.Sprintf("unable to create templatesDir: %s error: %v\n", templatesDir, err)))
		return
	}

	fmt.Printf("Saving templates to %s\n", templatesDir)
	err = SaveAssets(templatesDir, baseTemplates)
	if err != nil {
		fmt.Print(au.Red(fmt.Sprintf("Error saving: %v\n", err)))
	}
	return nil
}

func regenCmdLine() []string {
	cmdLine := []string{
		"gen",
		fmt.Sprintf(" --sqltype=%s", *sqlType),
		fmt.Sprintf(" --connstr='%s'", *sqlConnStr),
		fmt.Sprintf(" --database=%s", *sqlDatabase),
		fmt.Sprintf(" --templateDir=%s", "./templates"),
	}

	if *sqlTable != "" {
		cmdLine = append(cmdLine, fmt.Sprintf(" --table=%s", *sqlTable))
	}

	if *excludeSQLTables != "" {
		cmdLine = append(cmdLine, fmt.Sprintf(" --exclude=%s", *excludeSQLTables))
	}

	cmdLine = append(cmdLine, fmt.Sprintf(" --model=%s", *modelPackageName))
	cmdLine = append(cmdLine, fmt.Sprintf(" --dao=%s", *daoPackageName))
	cmdLine = append(cmdLine, fmt.Sprintf(" --api=%s", *apiPackageName))
	cmdLine = append(cmdLine, fmt.Sprintf(" --out=%s", "./"))
	cmdLine = append(cmdLine, fmt.Sprintf(" --module=%s", *module))
	if *addJSONAnnotation {
		cmdLine = append(cmdLine, fmt.Sprintf(" --json"))
		cmdLine = append(cmdLine, fmt.Sprintf(" --json-fmt=%s", *jsonNameFormat))
	}
	if *addXMLAnnotation {
		cmdLine = append(cmdLine, fmt.Sprintf(" --xml"))
		cmdLine = append(cmdLine, fmt.Sprintf(" --xml-fmt=%s", *xmlNameFormat))
	}
	if *addGormAnnotation {
		cmdLine = append(cmdLine, fmt.Sprintf(" --gorm"))
	}
	if *addProtobufAnnotation {
		cmdLine = append(cmdLine, fmt.Sprintf(" --protobuf"))
		cmdLine = append(cmdLine, fmt.Sprintf(" --proto-fmt=%s", *protoNameFormat))
	}
	if *addDBAnnotation {
		cmdLine = append(cmdLine, fmt.Sprintf(" --db"))
	}
	if *useGureguTypes {
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

	cmdLine = append(cmdLine, fmt.Sprintf(" --listen=%s", *serverListen))
	cmdLine = append(cmdLine, fmt.Sprintf(" --scheme=%s", *serverScheme))

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

// LoadTemplate return template from template dir, falling back to the embedded templates
func LoadTemplate(filename string) (tpl *dbmeta.GenTemplate, err error) {
	baseName := filepath.Base(filename)
	// fmt.Printf("LoadTemplate: %s / %s\n", filename, baseName)

	if *templateDir != "" {
		fpath := filepath.Join(*templateDir, filename)
		var b []byte
		b, err = ioutil.ReadFile(fpath)
		if err == nil {

			absPath, err := filepath.Abs(fpath)
			if err != nil {
				absPath = fpath
			}
			// fmt.Printf("Loaded template from file: %s\n", fpath)
			tpl = &dbmeta.GenTemplate{Name: "file://" + absPath, Content: string(b)}
			return tpl, nil
		}
	}

	content, err := baseTemplates.FindString(baseName)
	if err != nil {
		return nil, fmt.Errorf("%s not found internally", baseName)
	}
	if *verbose {
		fmt.Printf("Loaded template from app: %s\n", filename)
	}

	tpl = &dbmeta.GenTemplate{Name: "internal://" + filename, Content: content}
	return tpl, nil
}

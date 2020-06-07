package dbmeta

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/format"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/jinzhu/inflection"
	"github.com/serenize/snaker"
)

type TemplateLoader func(filename string) (content string, err error)

func (c *Config) GetTemplate(name, t string) (*template.Template, error) {
	var funcMap = template.FuncMap{
		"FmtFieldName":      FmtFieldName,
		"singular":          inflection.Singular,
		"pluralize":         inflection.Plural,
		"title":             strings.Title,
		"toLower":           strings.ToLower,
		"toUpper":           strings.ToUpper,
		"toLowerCamelCase":  camelToLowerCamel,
		"toSnakeCase":       snaker.CamelToSnake,
		"markdownCodeBlock": markdownCodeBlock,
		"wrapBash":          wrapBash,
		"GenerateTableFile": c.GenerateTableFile,
		"GenerateFile":      c.GenerateFile,
		"ToJSON":            ToJSON,
		"StringsJoin":       strings.Join,
	}

	tmpl, err := template.New(name).Option("missingkey=error").Funcs(funcMap).Parse(t)
	if err != nil {
		return nil, err
	}

	if name == "api.go.tmpl" || name == "dao_gorm.go.tmpl" || name == "dao_sqlx.go.tmpl" || name == "code_dao_sqlx.md.tmpl" || name == "code_dao_gorm.md.tmpl" || name == "code_http.md.tmpl" {

		operations := []string{"add", "delete", "get", "getall", "update"}
		for _, op := range operations {
			var filename string
			if name == "api.go.tmpl" {
				filename = fmt.Sprintf("api_%s.go.tmpl", op)
			}
			if name == "dao_gorm.go.tmpl" {
				filename = fmt.Sprintf("dao_gorm_%s.go.tmpl", op)
			}
			if name == "dao_sqlx.go.tmpl" {
				filename = fmt.Sprintf("dao_sqlx_%s.go.tmpl", op)
			}

			if name == "code_dao_sqlx.md.tmpl" {
				filename = fmt.Sprintf("dao_sqlx_%s.go.tmpl", op)
			}
			if name == "code_dao_gorm.md.tmpl" {
				filename = fmt.Sprintf("dao_gorm_%s.go.tmpl", op)
			}
			if name == "code_http.md.tmpl" {
				filename = fmt.Sprintf("api_%s.go.tmpl", op)
			}

			var subTemplate string
			if subTemplate, err = c.TemplateLoader(filename); err != nil {
				fmt.Printf("Error loading template %v\n", err)
				return nil, err
			}

			fmt.Printf("loading sub template %v\n", filename)

			tmpl.Parse(subTemplate)
		}
	}

	return tmpl, nil
}

// ToJSON func to return json string representation of struct
func ToJSON(val interface{}, indent int) string {
	pad := fmt.Sprintf("%*s", indent, "")
	strB, _ := json.MarshalIndent(val, "", pad)

	response := string(strB)
	response = strings.Replace(response, "\n", "", -1)
	return response
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

// GenerateTableFile generate file from template using specific table used within templates
func (c *Config) GenerateTableFile(tableInfos map[string]*ModelInfo, tableName, templateFilename, outputDirectory, outputFileName string, formatOutput bool) string {
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

	data := c.CreateContextForTableFile(tableInfo)

	fileOutDir := filepath.Join(c.OutDir, outputDirectory)
	err := os.MkdirAll(fileOutDir, 0777)
	if err != nil && !c.Overwrite {
		buf.WriteString(fmt.Sprintf("unable to create fileOutDir: %s error: %v\n", fileOutDir, err))
		return buf.String()
	}

	var tpl string
	if tpl, err = c.TemplateLoader(templateFilename); err != nil {
		buf.WriteString(fmt.Sprintf("Error loading template %v\n", err))
		return buf.String()
	}

	outputFile := filepath.Join(fileOutDir, outputFileName)
	buf.WriteString(fmt.Sprintf("Writing %s -> %s\n", templateFilename, outputFile))
	c.WriteTemplate(templateFilename, tpl, data, outputFile, formatOutput)
	return buf.String()
}

func (c *Config) CreateContextForTableFile(tableInfo *ModelInfo) map[string]interface{} {
	var modelInfo = map[string]interface{}{
		"StructName":      tableInfo.StructName,
		"TableName":       tableInfo.DBMeta.TableName(),
		"ShortStructName": strings.ToLower(string(tableInfo.StructName[0])),
		"TableInfo":       tableInfo,
	}

	nonPrimaryKeys := NonPrimaryKeyNames(tableInfo.DBMeta)
	modelInfo["NonPrimaryKeyNamesList"] = nonPrimaryKeys
	modelInfo["NonPrimaryKeysJoined"] = strings.Join(nonPrimaryKeys, ",")

	primaryKeys := PrimaryKeyNames(tableInfo.DBMeta)
	modelInfo["PrimaryKeyNamesList"] = primaryKeys
	modelInfo["PrimaryKeysJoined"] = strings.Join(primaryKeys, ",")

	delSql, err := GenerateDeleteSql(tableInfo.DBMeta)
	if err == nil {
		modelInfo["delSql"] = delSql
	}

	updateSql, err := GenerateUpdateSql(tableInfo.DBMeta)
	if err == nil {
		modelInfo["updateSql"] = updateSql
	}

	insertSql, err := GenerateInsertSql(tableInfo.DBMeta)
	if err == nil {
		modelInfo["insertSql"] = insertSql
	}

	selectOneSql, err := GenerateSelectOneSql(tableInfo.DBMeta)
	if err == nil {
		modelInfo["selectOneSql"] = selectOneSql
	}

	selectMultiSql, err := GenerateSelectMultiSql(tableInfo.DBMeta)
	if err == nil {
		modelInfo["selectMultiSql"] = selectMultiSql
	}
	return modelInfo
}

func (c *Config) WriteTemplate(name, templateStr string, data map[string]interface{}, outputFile string, formatOutput bool) {
	if !c.Overwrite && Exists(outputFile) {
		fmt.Printf("not overwriting %s\n", outputFile)
		return
	}

	for key, value := range c.ContextMap {
		data[key] = value
	}

	data["DatabaseName"] = c.SqlDatabase
	data["module"] = c.Module

	data["modelFQPN"] = c.ModelFQPN
	data["modelPackageName"] = c.ModelPackageName

	data["daoFQPN"] = c.DaoFQPN
	data["daoPackageName"] = c.DaoPackageName

	data["apiFQPN"] = c.ApiFQPN
	data["apiPackageName"] = c.ApiPackageName

	data["sqlType"] = c.SqlType
	data["sqlConnStr"] = c.SqlConnStr
	data["serverPort"] = c.ServerPort
	data["serverHost"] = c.ServerHost
	data["SwaggerInfo"] = c.Swagger
	data["outDir"] = c.OutDir
	data["CommandLine"] = c.CmdLine
	data["Config"] = c

	rt, err := c.GetTemplate(name, templateStr)
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

	if c.Verbose {
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

// GenerateFile generate file from template, non table used within templates
func (c *Config) GenerateFile(templateFilename, outDir, outputDirectory, outputFileName string, formatOutput bool, overwrite bool) string {
	buf := bytes.Buffer{}
	buf.WriteString(fmt.Sprintf("GenerateFile( %s, %s, %s)\n", templateFilename, outputDirectory, outputFileName))
	fileOutDir := filepath.Join(outDir, outputDirectory)
	err := os.MkdirAll(fileOutDir, 0777)
	if err != nil && !overwrite {
		buf.WriteString(fmt.Sprintf("unable to create fileOutDir: %s error: %v\n", fileOutDir, err))
		return buf.String()
	}

	data := map[string]interface{}{}

	var tpl string
	if tpl, err = c.TemplateLoader(templateFilename); err != nil {
		buf.WriteString(fmt.Sprintf("Error loading template %v\n", err))
		return buf.String()
	}

	outputFile := filepath.Join(fileOutDir, outputFileName)
	buf.WriteString(fmt.Sprintf("Writing %s -> %s\n", templateFilename, outputFile))
	c.WriteTemplate(templateFilename, tpl, data, outputFile, formatOutput)
	return buf.String()
}

type SwaggerInfoDetails struct {
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

type Config struct {
	SqlType               string
	SqlConnStr            string
	SqlDatabase           string
	Module                string
	ModelPackageName      string
	ModelFQPN             string
	AddJSONAnnotation     bool
	AddGormAnnotation     bool
	AddProtobufAnnotation bool
	AddDBAnnotation       bool
	UseGureguTypes        bool
	JsonNameFormat        string
	ProtobufNameFormat    string
	DaoPackageName        string
	DaoFQPN               string
	ApiPackageName        string
	ApiFQPN               string
	Swagger               *SwaggerInfoDetails
	ServerPort            int
	ServerHost            string
	Verbose               bool
	OutDir                string
	Overwrite             bool
	CmdLine               string
	ContextMap            map[string]interface{}
	TemplateLoader        TemplateLoader
}

func NewConfig(	templateLoader  TemplateLoader) *Config {
	conf := &Config{
		Swagger: &SwaggerInfoDetails{
			Version:      "1.0",
			BasePath:     "/",
			Title:        "Swagger Example API",
			Description:  "This is a sample server Petstore server.",
			TOS:          "",
			ContactName:  "",
			ContactURL:   "",
			ContactEmail: "",
		},
		TemplateLoader: templateLoader,
	}
	conf.CmdLine = strings.Join(os.Args, " ")
	conf.ContextMap = make(map[string]interface{})
	return conf
}

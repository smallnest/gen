package dbmeta

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"path"
	"time"

	"github.com/smallnest/gen/utils"

	"go/format"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"text/template"

	"github.com/davecgh/go-spew/spew"
	"github.com/jinzhu/inflection"
	"github.com/serenize/snaker"
	"golang.org/x/tools/imports"
)

// GenTemplate template info struct
type GenTemplate struct {
	Name    string
	Content string
}

// TemplateLoader loader function to retrieve a template contents
type TemplateLoader func(filename string) (tpl *GenTemplate, err error)

var replaceFuncMap = template.FuncMap{
	"singular":           inflection.Singular,
	"pluralize":          inflection.Plural,
	"title":              strings.Title,
	"toLower":            strings.ToLower,
	"toUpper":            strings.ToUpper,
	"toLowerCamelCase":   camelToLowerCamel,
	"toUpperCamelCase":   camelToUpperCamel,
	"toSnakeCase":        snaker.CamelToSnake,
	"StringsJoin":        strings.Join,
	"replace":            replace,
	"stringifyFirstChar": stringifyFirstChar,
	"FmtFieldName":       FmtFieldName,
}

func replace(input, from, to string) string {
	return strings.Replace(input, from, to, -1)
}

// Replace takes a template based name format and will render a name using it
func Replace(nameFormat, name string) string {
	var tpl bytes.Buffer
	//fmt.Printf("Replace: %s\n",nameFormat)
	t := template.Must(template.New("t1").Funcs(replaceFuncMap).Parse(nameFormat))

	if err := t.Execute(&tpl, name); err != nil {
		//fmt.Printf("Error creating name format: %s error: %v\n", nameFormat, err)
		return name
	}
	result := tpl.String()

	result = strings.Trim(result, " \t")
	result = strings.Replace(result, " ", "_", -1)
	result = strings.Replace(result, "\t", "_", -1)

	//fmt.Printf("Replace( '%s' '%s')= %s\n",nameFormat, name, result)
	return result
}

// ReplaceFileNamingTemplate use the FileNamingTemplate to format a table name
func (c *Config) ReplaceFileNamingTemplate(name string) string {
	return Replace(c.FileNamingTemplate, name)
}

// ReplaceModelNamingTemplate use the ModelNamingTemplate to format a table name
func (c *Config) ReplaceModelNamingTemplate(name string) string {
	return Replace(c.ModelNamingTemplate, name)
}

// ReplaceFieldNamingTemplate use the FieldNamingTemplate to format a table name
func (c *Config) ReplaceFieldNamingTemplate(name string) string {
	return Replace(c.FieldNamingTemplate, name)
}

// GetTemplate return a Template based on a name and template contents
func (c *Config) GetTemplate(genTemplate *GenTemplate) (*template.Template, error) {
	var s State
	var funcMap = template.FuncMap{
		"ReplaceFileNamingTemplate":  c.ReplaceFileNamingTemplate,
		"ReplaceModelNamingTemplate": c.ReplaceModelNamingTemplate,
		"ReplaceFieldNamingTemplate": c.ReplaceFieldNamingTemplate,
		"stringifyFirstChar":         stringifyFirstChar,
		"singular":                   inflection.Singular,
		"pluralize":                  inflection.Plural,
		"title":                      strings.Title,
		"toLower":                    strings.ToLower,
		"toUpper":                    strings.ToUpper,
		"toLowerCamelCase":           camelToLowerCamel,
		"toUpperCamelCase":           camelToUpperCamel,
		"FormatSource":               FormatSource,
		"toSnakeCase":                snaker.CamelToSnake,
		"markdownCodeBlock":          markdownCodeBlock,
		"wrapBash":                   wrapBash,
		"escape":                     escape,
		"GenerateTableFile":          c.GenerateTableFile,
		"GenerateFile":               c.GenerateFile,
		"ToJSON":                     ToJSON,
		"spew":                       Spew,
		"set":                        s.Set,
		"inc":                        s.Inc,
		"StringsJoin":                strings.Join,
		"replace":                    replace,
		"hasField":                   hasField,
		"FmtFieldName":               FmtFieldName,
		"copy":                       c.FileSystemCopy,
		"mkdir":                      c.Mkdir,
		"touch":                      c.Touch,
		"pwd":                        Pwd,
		"config":                     c.DisplayConfig,
		"insertFragment":             c.insertFragment,
	}

	baseName := filepath.Base(genTemplate.Name)

	tmpl, err := template.New(baseName).Option("missingkey=error").Funcs(funcMap).Parse(genTemplate.Content)
	if err != nil {
		return nil, err
	}

	if baseName == "api.go.tmpl" ||
		baseName == "dao_gorm.go.tmpl" ||
		baseName == "dao_sqlx.go.tmpl" ||
		baseName == "code_dao_sqlx.md.tmpl" ||
		baseName == "code_dao_gorm.md.tmpl" ||
		baseName == "code_http.md.tmpl" {

		operations := []string{"add", "delete", "get", "getall", "update"}
		for _, op := range operations {
			var filename string
			if baseName == "api.go.tmpl" {
				filename = fmt.Sprintf("api_%s.go.tmpl", op)
			}
			if baseName == "dao_gorm.go.tmpl" {
				filename = fmt.Sprintf("dao_gorm_%s.go.tmpl", op)
			}
			if baseName == "dao_sqlx.go.tmpl" {
				filename = fmt.Sprintf("dao_sqlx_%s.go.tmpl", op)
			}

			if baseName == "code_dao_sqlx.md.tmpl" {
				filename = fmt.Sprintf("dao_sqlx_%s.go.tmpl", op)
			}
			if baseName == "code_dao_gorm.md.tmpl" {
				filename = fmt.Sprintf("dao_gorm_%s.go.tmpl", op)
			}
			if baseName == "code_http.md.tmpl" {
				filename = fmt.Sprintf("api_%s.go.tmpl", op)
			}

			var subTemplate *GenTemplate
			if subTemplate, err = c.TemplateLoader(filename); err != nil {
				fmt.Printf("Error loading template %v\n", err)
				return nil, err
			}

			// fmt.Printf("loading sub template %v\n", filename)
			tmpl.Parse(subTemplate.Content)
		}
	}

	return tmpl, nil
}

func hasField(v interface{}, name string) bool {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return false
	}
	return rv.FieldByName(name).IsValid()
}

// ToJSON func to return json string representation of struct
func ToJSON(val interface{}, indent int) string {
	pad := fmt.Sprintf("%*s", indent, "")
	strB, _ := json.MarshalIndent(val, "", pad)

	response := string(strB)
	response = strings.Replace(response, "\n", "", -1)
	return response
}

// Spew func to return spewed string representation of struct
func Spew(val interface{}) string {
	return spew.Sdump(val)
}

// State struct used for storing state in template parsing
type State struct {
	n int
}

// Set set state value in template parsing
func (s *State) Set(n int) int {
	s.n = n
	return n
}

// Inc increment state value in template parsing
func (s *State) Inc() int {
	s.n++
	return s.n
}

func camelToLowerCamel(s string) string {
	ss := strings.Split(s, "")
	ss[0] = strings.ToLower(ss[0])
	return strings.Join(ss, "")
}

func camelToUpperCamel(s string) string {
	ss := strings.Split(s, "")
	ss[0] = strings.ToUpper(ss[0])
	return strings.Join(ss, "")
}

// FormatSource format source code contents
func FormatSource(s string) string {
	formattedSource, err := format.Source([]byte(s))
	if err != nil {
		return fmt.Sprintf("Error in formatting source: %s\n", err.Error())
	}
	formattedSource, err = imports.Process("", formattedSource, nil)
	if err != nil {
		return fmt.Sprintf("Error in formatting source: %s\n", err.Error())
	}
	return string(formattedSource)
}

func markdownCodeBlock(contentType, content string) string {
	// fmt.Printf("%s - %s\n", contentType, content)
	return fmt.Sprintf("```%s\n%s\n```\n", contentType, content)
}

func wrapBash(content string) string {

	r := csv.NewReader(strings.NewReader(content))
	r.Comma = ' '
	record, err := r.Read()
	if err != nil {
		return content
	}

	fmt.Printf("[%s]\n", content)

	for i, j := range record {
		fmt.Printf("wrapBash [%d] %s\n", i, j)
	}

	out := strings.Join(record, " \\\n    ")
	return out

	//
	//
	//r := regexp.MustCompile(`[^\s"']+|"([^"]*)"|'([^']*)`)
	//arr := r.FindAllString(content, -1)
	//return strings.Join(arr, " \\\n    ")
	//
	//

	//splitter := "[^\\s\"']+|\"[^\"]*\"|'[^']*'"
	//result := RegSplit(content, splitter)
	//return strings.Join(result, " \\\n    ")

	//
	//result, err := parseCommandLine(content)
	//if err != nil {
	//	return content
	//}
	//return strings.Join(result, " \\\n    ")
}

// RegSplit split text based on regex
func RegSplit(text string, delimeter string) []string {
	reg := regexp.MustCompile(delimeter)
	indexes := reg.FindAllStringIndex(text, -1)
	laststart := 0
	result := make([]string, len(indexes)+1)
	for i, element := range indexes {
		result[i] = text[laststart:element[0]]
		laststart = element[1]
	}
	result[len(indexes)] = text[laststart:]
	return result
}

func parseCommandLine(command string) ([]string, error) {
	var args []string
	state := "start"
	current := ""
	quote := "\""
	escapeNext := true
	for i := 0; i < len(command); i++ {
		c := command[i]

		if state == "quotes" {
			if string(c) != quote {
				current += string(c)
			} else {
				args = append(args, current)
				current = ""
				state = "start"
			}
			continue
		}

		if escapeNext {
			current += string(c)
			escapeNext = false
			continue
		}

		if c == '\\' {
			escapeNext = true
			continue
		}

		if c == '"' || c == '\'' {
			state = "quotes"
			quote = string(c)
			continue
		}

		if state == "arg" {
			if c == ' ' || c == '\t' {
				args = append(args, current)
				current = ""
				state = "start"
			} else {
				current += string(c)
			}
			continue
		}

		if c != ' ' && c != '\t' {
			state = "arg"
			current += string(c)
		}
	}

	if state == "quotes" {
		return []string{}, fmt.Errorf("unclosed quote in command line: %s", command)
	}

	if current != "" {
		args = append(args, current)
	}

	return args, nil
}

func escape(content string) string {
	content = strings.Replace(content, "\"", "\\\"", -1)
	content = strings.Replace(content, "'", "\\'", -1)
	return content
}

// LoadFragments read all filed inside dirname to `fragments`
func (c *Config) LoadFragments(dirname string) error {
	c.FragmentsDir = dirname
	files, err := ioutil.ReadDir(dirname)
	if err != nil {
		return err
	}

	c.fragments = &bytes.Buffer{}
	for _, file := range files {
		if !file.IsDir() {
			filename := path.Join(dirname, file.Name())
			content, err := ioutil.ReadFile(filename)
			if err != nil {
				return err
			}
			c.fragments.Write(content)
			c.fragments.WriteRune('\n')
		}
	}
	return nil
}

// Insert code fragment from file.
// Within a single file, fragment can be marked as // fragment: <name> ... // end
func (c *Config) insertFragment(name, defValue string) (string, error) {
	if c.fragments == nil || c.fragments.Len() == 0 {
		return defValue, nil
	}
	// pointer to fragment content
	content := c.fragments.Bytes()

	// if fragment name is not defined,
	// return whole content
	if name == "" {
		return string(content), nil
	}

	// Use regex to extract fragment name from given file.
	// Below is an example of code fragment.

	/*
		//fragment: <name>
		func (w *Model) MyCode() {

		}
		// end
	*/
	begExp := fmt.Sprintf(`//[\s]{0,}fragment[\s]{0,}:[\s]{0,}%s[\s]{1,}`, name)
	reBeg, err := regexp.Compile(begExp)
	if err != nil {
		return defValue, nil
	}
	begLoc := reBeg.FindIndex(content)
	if begLoc == nil || begLoc[1] == len(content) {
		// not found or no content specified after // fragment: <name>
		return defValue, nil
	}

	fromIdx := begLoc[1]
	subContent := content[fromIdx:]
	reEnd := regexp.MustCompile(`//[\s]end[\s]{0,}`)
	endLoc := reEnd.FindIndex(subContent)
	if endLoc == nil {
		// not found, return remaining content
		return string(subContent), nil
	}
	return string(subContent[:endLoc[0]]), nil
}

// JSONFieldName convert name to appropriate case
func (c *Config) JSONFieldName(name string) string {
	return formatFieldName(c.JSONNameFormat, name)
}

// JSONTag converts name to `json:"name"` respecting json-fmt option
func (c *Config) JSONTag(name string) string {
	return fmt.Sprintf("`json:\"%s\"`", c.JSONFieldName(name))
}

// JSONTagOmitEmpty converts name to JSON tag with omitempty
func (c *Config) JSONTagOmitEmpty(name string) string {
	return fmt.Sprintf("`json:\"%s,omitempty\"`", c.JSONFieldName(name))
}

// GenerateTableFile generate file from template using specific table used within templates
func (c *Config) GenerateTableFile(tableName, templateFilename, outputDirectory, outputFileName string) string {
	buf := bytes.Buffer{}

	buf.WriteString(fmt.Sprintf("GenerateTableFile( %s, %s, %s, %s)\n", tableName, templateFilename, outputDirectory, outputFileName))

	tableInfo, ok := c.TableInfos[tableName]
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

	var tpl *GenTemplate
	if tpl, err = c.TemplateLoader(templateFilename); err != nil {
		buf.WriteString(fmt.Sprintf("Error loading template %v\n", err))
		return buf.String()
	}

	outputFile := filepath.Join(fileOutDir, outputFileName)
	buf.WriteString(fmt.Sprintf("Writing %s -> %s\n", templateFilename, outputFile))
	err = c.WriteTemplate(tpl, data, outputFile)
	return buf.String()
}

// CreateContextForTableFile create map context for a db table
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

	delSQL, err := GenerateDeleteSQL(tableInfo.DBMeta)
	if err == nil {
		modelInfo["delSql"] = delSQL
	}

	updateSQL, err := GenerateUpdateSQL(tableInfo.DBMeta)
	if err == nil {
		modelInfo["updateSql"] = updateSQL
	}

	insertSQL, err := GenerateInsertSQL(tableInfo.DBMeta)
	if err == nil {
		modelInfo["insertSql"] = insertSQL
	}

	selectOneSQL, err := GenerateSelectOneSQL(tableInfo.DBMeta)
	if err == nil {
		modelInfo["selectOneSql"] = selectOneSQL
	}

	selectMultiSQL, err := GenerateSelectMultiSQL(tableInfo.DBMeta)
	if err == nil {
		modelInfo["selectMultiSql"] = selectMultiSQL
	}
	return modelInfo
}

// WriteTemplate write a template out
func (c *Config) WriteTemplate(genTemplate *GenTemplate, data map[string]interface{}, outputFile string) error {
	//fmt.Printf("WriteTemplate %s\n", outputFile)

	if !c.Overwrite && Exists(outputFile) {
		fmt.Printf("not overwriting %s\n", outputFile)
		return nil
	}

	for key, value := range c.ContextMap {
		data[key] = value
	}

	dir := filepath.Dir(outputFile)
	parent := filepath.Base(dir)

	data["File"] = outputFile
	data["Dir"] = dir
	data["Parent"] = parent

	data["DatabaseName"] = c.SQLDatabase
	data["module"] = c.Module

	data["modelFQPN"] = c.ModelFQPN
	data["modelPackageName"] = c.ModelPackageName

	data["daoFQPN"] = c.DaoFQPN
	data["daoPackageName"] = c.DaoPackageName

	data["UseGuregu"] = c.UseGureguTypes

	data["apiFQPN"] = c.APIFQPN
	data["apiPackageName"] = c.APIPackageName

	data["sqlType"] = c.SQLType
	data["sqlConnStr"] = c.SQLConnStr
	data["serverPort"] = c.ServerPort
	data["serverHost"] = c.ServerHost
	data["serverScheme"] = c.ServerScheme
	data["serverListen"] = c.ServerListen
	data["SwaggerInfo"] = c.Swagger
	data["outDir"] = c.OutDir
	data["Config"] = c

	rt, err := c.GetTemplate(genTemplate)
	if err != nil {
		return fmt.Errorf("error in loading %s template, error: %v", genTemplate.Name, err)
	}
	var buf bytes.Buffer
	err = rt.Execute(&buf, data)
	if err != nil {
		return fmt.Errorf("error in rendering %s: %s", genTemplate.Name, err.Error())
	}

	fileContents, err := c.format(genTemplate, buf.Bytes(), outputFile)
	if err != nil {
		return fmt.Errorf("error writing %s - error: %v", outputFile, err)
	}

	err = ioutil.WriteFile(outputFile, fileContents, 0777)
	if err != nil {
		return fmt.Errorf("error writing %s - error: %v", outputFile, err)
	}

	if c.Verbose {
		fmt.Printf("writing %s\n", outputFile)
	}
	return nil
}

func (c *Config) format(genTemplate *GenTemplate, content []byte, outputFile string) ([]byte, error) {
	extension := filepath.Ext(outputFile)
	if extension == ".go" {
		formattedSource, err := format.Source([]byte(content))
		if err != nil {
			return nil, fmt.Errorf("error in formatting template: %s outputfile: %s source: %s", genTemplate.Name, outputFile, err.Error())
		}

		fileContents := NormalizeNewlines(formattedSource)
		if c.LineEndingCRLF {
			fileContents = CRLFNewlines(formattedSource)
		}
		return fileContents, nil
	}

	fileContents := NormalizeNewlines([]byte(content))
	if c.LineEndingCRLF {
		fileContents = CRLFNewlines(fileContents)
	}
	return fileContents, nil
}

// NormalizeNewlines normalizes \r\n (windows) and \r (mac)
// into \n (unix)
func NormalizeNewlines(d []byte) []byte {
	// replace CR LF \r\n (windows) with LF \n (unix)
	d = bytes.Replace(d, []byte{13, 10}, []byte{10}, -1)
	// replace CF \r (mac) with LF \n (unix)
	d = bytes.Replace(d, []byte{13}, []byte{10}, -1)
	return d
}

// CRLFNewlines transforms \n to \r\n (windows)
func CRLFNewlines(d []byte) []byte {
	// replace LF (unix) with CR LF \r\n (windows)
	d = bytes.Replace(d, []byte{10}, []byte{13, 10}, -1)
	return d
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
func (c *Config) GenerateFile(templateFilename, outputDirectory, outputFileName string, overwrite bool) string {
	buf := bytes.Buffer{}
	buf.WriteString(fmt.Sprintf("GenerateFile( %s, %s, %s)\n", templateFilename, outputDirectory, outputFileName))
	fileOutDir := outputDirectory
	err := os.MkdirAll(fileOutDir, 0777)
	if err != nil && !overwrite {
		buf.WriteString(fmt.Sprintf("unable to create fileOutDir: %s error: %v\n", fileOutDir, err))
		return buf.String()
	}

	data := map[string]interface{}{}

	var tpl *GenTemplate
	if tpl, err = c.TemplateLoader(templateFilename); err != nil {
		buf.WriteString(fmt.Sprintf("Error loading template %v\n", err))
		return buf.String()
	}

	outputFile := filepath.Join(fileOutDir, outputFileName)
	buf.WriteString(fmt.Sprintf("Writing %s -> %s\n", templateFilename, outputFile))
	err = c.WriteTemplate(tpl, data, outputFile)
	if err != nil {
		buf.WriteString(fmt.Sprintf("Error calling WriteTemplate %s -> %v\n", templateFilename, err))
	}
	return buf.String()
}

// DisplayConfig display config info
func (c *Config) DisplayConfig() string {

	info := fmt.Sprintf(
		`DisplayConfig
  SQLType      : %s
  SQLConnStr   : %s
  SQLDatabase  : %s
  Module       : %s
  OutDir       : %s
`, c.SQLType, c.SQLConnStr, c.SQLDatabase, c.Module, c.OutDir)

	return info
}

type copyRules struct {
	result  bool
	pattern string
	r       *regexp.Regexp
}

// FileSystemCopy template command to copy files, directories and to pass --include XXX and --exclude YYY regular expressions. Files ending in .tmpl will be processed as a template.
// Files ending in .table.tmpl will be processed as a template iterating through all the tables
func (c *Config) FileSystemCopy(src, dst string, options ...string) string {
	dstDir := filepath.Join(c.OutDir, dst)

	patterns := make([]*copyRules, 0)

	for _, o := range options {

		if strings.HasPrefix(o, "--exclude ") {
			pattern := o[len("--exclude "):]
			r, _ := regexp.Compile(pattern)
			if r != nil {
				patterns = append(patterns, &copyRules{result: false, r: r, pattern: pattern})
			}
		}

		if strings.HasPrefix(o, "--include ") {
			pattern := o[len("--include "):]
			r, _ := regexp.Compile(pattern)
			if r != nil {
				patterns = append(patterns, &copyRules{result: true, r: r, pattern: pattern})
			}
		}
	}

	opt := utils.DefaultCopyOptions()

	opt.ShouldCopy = func(info os.FileInfo) bool {
		name := info.Name()

		for _, r := range patterns {
			if r.r.Match([]byte(name)) {
				//fmt.Printf("copy ShouldCopy %s  pattern: [%s]  result: %t\n", name, r.pattern, r.result)
				return r.result
			}
		}

		return true
	}

	opt.FileHandler = func(src, dest string, info os.FileInfo) utils.FileHandlerFunc {

		if !opt.ShouldCopy(info) {
			return func(src, dest string, info os.FileInfo, opt utils.Options, results *utils.Results) (err error) {
				results.Info.WriteString(fmt.Sprintf("CopyFile Skipping %s\n", src))
				return nil
			}
		}

		if strings.HasSuffix(src, ".table.tmpl") {
			//fmt.Printf("@@ HandleTableTemplateFile: src: %s  dest: %s Name: %s\n", src, dest, info.Name())
			return c.tableFileHandlerFunc
		}
		if strings.HasSuffix(src, ".tmpl") {
			//fmt.Printf("@@ HandleTemplateFile: src: %s  dest: %s Name: %s\n", src, dest, info.Name())
			return c.fileHandlerFunc
		}

		return func(src, dest string, info os.FileInfo, opt utils.Options, results *utils.Results) (err error) {
			results.Info.WriteString(fmt.Sprintf("CopyFile %s\n", dest))
			return utils.DefaultFileCopy(src, dest, info, opt, results)
		}
	}

	result, err := utils.Copy(src, dstDir, opt)
	if err != nil {
		return fmt.Sprintf("copy returned an error %v", err)
	}
	return fmt.Sprintf("copy %s %s\n%s\n", src, dstDir, result.String())
}

// Mkdir template command to mkdir under the output directory
func (c *Config) Mkdir(dst string) string {
	dstDir := filepath.Join(c.OutDir, dst)

	err := os.MkdirAll(dstDir, os.ModePerm)
	if err != nil {
		return fmt.Sprintf("mkdir returned an error %v", err)

	}
	return fmt.Sprintf("mkdir %s", dstDir)
}

// Touch template command to touch a file under the output directory
func (c *Config) Touch(dst string) string {
	dstDir := filepath.Join(c.OutDir, dst)

	_, err := os.Stat(dstDir)
	if os.IsNotExist(err) {
		file, err := os.Create(dstDir)
		if err != nil {
			return fmt.Sprintf("touch returned an error %v", err)
		}
		defer file.Close()
	} else {
		currentTime := time.Now().Local()
		err = os.Chtimes(dstDir, currentTime, currentTime)
		if err != nil {
			return fmt.Sprintf("touch returned an error %v", err)
		}
	}
	return fmt.Sprintf("touch %s", dstDir)
}

// ".tmpl"
func (c *Config) fileHandlerFunc(src, dest string, info os.FileInfo, opt utils.Options, results *utils.Results) (err error) {
	genTemplate := &GenTemplate{
		Name:    info.Name(),
		Content: loadFile(src),
	}

	data := make(map[string]interface{})

	outputFile := dest[0 : len(dest)-5]
	results.Info.WriteString(fmt.Sprintf("WriteTemplate %s\n", outputFile))
	return c.WriteTemplate(genTemplate, data, outputFile)
}

// ".table.tmpl"
func (c *Config) tableFileHandlerFunc(src, dest string, info os.FileInfo, opt utils.Options, results *utils.Results) (err error) {
	genTemplate := &GenTemplate{
		Name:    info.Name(),
		Content: loadFile(src),
	}

	outputFile := dest[0 : len(dest)-11]

	dir := filepath.Dir(outputFile)
	tmplateName := filepath.Base(outputFile)
	// parent := filepath.Base(dir)
	results.Info.WriteString(fmt.Sprintf("WriteTableTemplate %s\n", src))

	for tableName, tableInfo := range c.TableInfos {
		data := c.CreateContextForTableFile(tableInfo)
		// fileName := filepath.Join(dir, tableName+name)
		name := c.ReplaceFileNamingTemplate(tableName) + filepath.Ext(tmplateName)
		fileName := filepath.Join(dir, name)
		results.Info.WriteString(fmt.Sprintf("    table: %-25s  %s\n", tableName, fileName))
		c.WriteTemplate(genTemplate, data, fileName)
	}
	return nil
}

func loadFile(src string) string {
	// Read entire file content, giving us little control but
	// making it very simple. No need to close the file.
	content, err := ioutil.ReadFile(src)
	if err != nil {
		return fmt.Sprintf("error loading %s error: %v", src, err)
	}

	// Convert []byte to string and print to screen
	text := string(content)
	return text
}

// SwaggerInfoDetails swagger details
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

// Config for generating code
type Config struct {
	SQLType               string
	SQLConnStr            string
	SQLDatabase           string
	Module                string
	ModelPackageName      string
	ModelFQPN             string
	AddJSONAnnotation     bool
	AddGormAnnotation     bool
	AddProtobufAnnotation bool
	AddXMLAnnotation      bool
	AddDBAnnotation       bool
	UseGureguTypes        bool
	JSONNameFormat        string
	XMLNameFormat         string
	ProtobufNameFormat    string
	DaoPackageName        string
	DaoFQPN               string
	APIPackageName        string
	APIFQPN               string
	GrpcPackageName       string
	GrpcFQPN              string
	Swagger               *SwaggerInfoDetails
	ServerPort            int
	ServerHost            string
	ServerScheme          string
	ServerListen          string
	Verbose               bool
	OutDir                string
	Overwrite             bool
	LineEndingCRLF        bool
	CmdLine               string
	CmdLineWrapped        string
	CmdLineArgs           []string
	FileNamingTemplate    string
	ModelNamingTemplate   string
	FieldNamingTemplate   string
	ContextMap            map[string]interface{}
	TemplateLoader        TemplateLoader
	TableInfos            map[string]*ModelInfo
	FragmentsDir          string
	fragments             *bytes.Buffer
}

// NewConfig create a new code config
func NewConfig(templateLoader TemplateLoader) *Config {
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
	conf.CmdLineArgs = os.Args
	conf.CmdLineWrapped = strings.Join(os.Args, " \\\n    ")
	conf.CmdLine = strings.Join(os.Args, " ")

	conf.ContextMap = make(map[string]interface{})

	conf.FileNamingTemplate = "{{.}}"
	conf.ModelNamingTemplate = "{{FmtFieldName .}}"
	conf.FieldNamingTemplate = "{{FmtFieldName (stringifyFirstChar .) }}"

	outDir := "."
	module := "github.com/alexj212/test"
	modelPackageName := "model"
	daoPackageName := "dao"
	apiPackageName := "api"

	conf.ModelPackageName = modelPackageName
	conf.DaoPackageName = daoPackageName
	conf.APIPackageName = apiPackageName

	conf.AddJSONAnnotation = true
	conf.AddXMLAnnotation = true
	conf.AddGormAnnotation = true
	conf.AddProtobufAnnotation = true
	conf.AddDBAnnotation = true
	conf.UseGureguTypes = false
	conf.JSONNameFormat = "snake"
	conf.XMLNameFormat = "snake"
	conf.ProtobufNameFormat = "snake"
	conf.Verbose = false
	conf.OutDir = outDir
	conf.Overwrite = true

	conf.ServerPort = 8080
	conf.ServerHost = "127.0.0.1"
	conf.ServerScheme = "http"
	conf.ServerListen = ":8080"
	conf.Overwrite = true

	conf.Module = module
	conf.ModelFQPN = module + "/" + modelPackageName
	conf.DaoFQPN = module + "/" + daoPackageName
	conf.APIFQPN = module + "/" + apiPackageName

	if conf.ServerPort == 80 {
		conf.Swagger.Host = conf.ServerHost
	} else {
		conf.Swagger.Host = fmt.Sprintf("%s:%d", conf.ServerHost, conf.ServerPort)
	}

	return conf
}

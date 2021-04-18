[comment]: <> (This is a generated file please edit source in ./templates)
[comment]: <> (All modification will be lost, you have been warned)
[comment]: <> ()
## gen

[![License](https://img.shields.io/badge/License-Apache%203.0-blue.svg)](https://opensource.org/licenses/Apache-3.0) [![GoDoc](https://godoc.org/github.com/smallnest/gen?status.png)](http://godoc.org/github.com/smallnest/gen)  [![travis](https://travis-ci.org/smallnest/gen.svg?branch=master)](https://travis-ci.org/smallnest/gen) [![Go Report Card](https://goreportcard.com/badge/github.com/smallnest/gen)](https://goreportcard.com/report/github.com/smallnest/gen)

The gen tool produces a CRUD (Create, read, update and delete) REST api project template from a given database. The gen tool will
connect to the db connection string analyze the database and generate the code based on the flags provided.

By reading details from the database about the column structure, gen generates a go compatible struct type
with the required column names, data types, and annotations.

It supports [gorm](https://github.com/jinzhu/gorm) tags and implements some usable methods. Generated data types include support for nullable columns [sql.NullX types](https://golang.org/pkg/database/sql/#NullBool) or [guregu null.X types](https://github.com/guregu/null)
and the expected basic built in go types.

`gen` is based / inspired by the work of Seth Shelnutt's [db2struct](https://github.com/Shelnutt2/db2struct), and Db2Struct is based/inspired by the work of ChimeraCoder's gojson package [gojson](https://github.com/ChimeraCoder/gojson).



## CRUD Generation
This is a sample table contained within the ./example/sample.db Sqlite3 database. Using `gen` will generate the following struct.
```sql
CREATE TABLE "albums"
(
    [AlbumId]  INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    [Title]    NVARCHAR(160) NOT NULL,
    [ArtistId] INTEGER NOT NULL,
    FOREIGN KEY ([ArtistId]) REFERENCES "artists" ([ArtistId])
		ON DELETE NO ACTION ON UPDATE NO ACTION
)
```
#### Transforms into
```go
type Album struct {
	//[ 0] AlbumId                                        integer              null: false  primary: true   auto: true   col: integer         len: -1      default: []
	AlbumID int `gorm:"primary_key;AUTO_INCREMENT;column:AlbumId;type:INTEGER;" json:"album_id" db:"AlbumId" protobuf:"int32,0,opt,name=album_id"`
	//[ 1] Title                                          nvarchar(160)        null: false  primary: false  auto: false  col: nvarchar        len: 160     default: []
	Title string `gorm:"column:Title;type:NVARCHAR(160);size:160;" json:"title" db:"Title" protobuf:"string,1,opt,name=title"`
	//[ 2] ArtistId                                       integer              null: false  primary: false  auto: false  col: integer         len: -1      default: []
	ArtistID int `gorm:"column:ArtistId;type:INTEGER;" json:"artist_id" db:"ArtistId" protobuf:"int32,2,opt,name=artist_id"`
}
```
Code generation for a complete CRUD rest project is possible with DAO crud functions, http handlers, makefile, sample server are available. Check out some of the [code generated samples](#Generated-Samples).



## Binary Installation
```BASH
## install gen tool (should be installed to ~/go/bin, make sure ~/go/bin is in your path.
$ go get -u github.com/smallnest/gen

## download sample sqlite database
$ wget https://github.com/smallnest/gen/raw/master/example/sample.db

## generate code based on the sqlite database (project will be contained within the ./example dir)
$ gen --sqltype=sqlite3 \
   	--connstr "./sample.db" \
   	--database main  \
   	--json \
   	--gorm \
   	--guregu \
   	--rest \
   	--out ./example \
   	--module example.com/rest/example \
   	--mod \
   	--server \
   	--makefile \
   	--json-fmt=snake \
   	--generate-dao \
   	--generate-proj \
   	--overwrite

## build example code (build process will install packr2 if not installed)
$ cd ./example
$ make example

## binary will be located at ./bin/example
## when launching make sure that the SQLite file sample.db is located in the same dir as the binary
$ cp ../../sample.db  .
$ ./example


## Open a browser to http://127.0.0.1:8080/swagger/index.html

## Use wget/curl/httpie to fetch via command line
http http://localhost:8080/albums
curl http://localhost:8080/artists

```


## Usage
```console
Usage of gen:
	gen [-v] --sqltype=mysql --connstr "user:password@/dbname" --database <databaseName> --module=example.com/example [--json] [--gorm] [--guregu] [--generate-dao] [--generate-proj]
git fetch up
           sqltype - sql database type such as [ mysql, mssql, postgres, sqlite, etc. ]


Options:
  --sqltype=mysql                                          sql database type such as [ mysql, mssql, postgres, sqlite, etc. ]
  -c, --connstr=nil                                        database connection string
  -d, --database=nil                                       Database to for connection
  -t, --table=                                             Table to build struct from
  -x, --exclude=                                           Table(s) to exclude
  --templateDir=                                           Template Dir
  --fragmentsDir=                                          Code fragments Dir
  --save=                                                  Save templates to dir
  --model=model                                            name to set for model package
  --model_naming={{FmtFieldName .}}                        model naming template to name structs
  --field_naming={{FmtFieldName (stringifyFirstChar .) }}  field naming template to name structs
  --file_naming={{.}}                                      file_naming template to name files
  --dao=dao                                                name to set for dao package
  --api=api                                                name to set for api package
  --grpc=grpc                                              name to set for grpc package
  --out=.                                                  output dir
  --module=example.com/example                             module path
  --overwrite                                              Overwrite existing files (default)
  --no-overwrite                                           disable overwriting files
  --windows                                                use windows line endings in generated files
  --no-color                                               disable color output
  --context=                                               context file (json) to populate context with
  --mapping=                                               mapping file (json) to map sql types to golang/protobuf etc
  --exec=                                                  execute script for custom code generation
  --json                                                   Add json annotations (default)
  --no-json                                                Disable json annotations
  --json-fmt=snake                                         json name format [snake | camel | lower_camel | none]
  --xml                                                    Add xml annotations (default)
  --no-xml                                                 Disable xml annotations
  --xml-fmt=snake                                          xml name format [snake | camel | lower_camel | none]
  --gorm                                                   Add gorm annotations (tags)
  --protobuf                                               Add protobuf annotations (tags)
  --proto-fmt=snake                                        proto name format [snake | camel | lower_camel | none]
  --gogo-proto=                                            location of gogo import 
  --db                                                     Add db annotations (tags)
  --guregu                                                 Add guregu null types
  --copy-templates                                         Copy regeneration templates to project directory
  --mod                                                    Generate go.mod in output dir
  --makefile                                               Generate Makefile in output dir
  --server                                                 Generate server app output dir
  --generate-dao                                           Generate dao functions
  --generate-proj                                          Generate project readme and gitignore
  --rest                                                   Enable generating RESTful api
  --run-gofmt                                              run gofmt on output dir
  --listen=                                                listen address e.g. :8080
  --scheme=http                                            scheme for server url
  --host=localhost                                         host for server
  --port=8080                                              port for server
  --swagger_version=1.0                                    swagger version
  --swagger_path=/                                         swagger base path
  --swagger_tos=                                           swagger tos url
  --swagger_contact_name=Me                                swagger contact name
  --swagger_contact_url=http://me.com/terms.html           swagger contact url
  --swagger_contact_email=me@me.com                        swagger contact email
  -v, --verbose                                            Enable verbose output
  --name_test=                                             perform name test using the --model_naming or --file_naming options
  -h, --help                                               Show usage message
  --version                                                Show version


```

## Building
The project contains a makefile for easy building and common tasks.
* `go get` - get the relevant dependencies as a "go" software
* `make help` - list available targets
* `make build` - generate the binary `./gen`
* `make example` - run the gen process on the example SqlLite db located in ./examples place the sources in ./example
Other targets exist for dev tasks.

## Example
The project provides a sample SQLite database in the `./example` directory. From the project `Makefile` can be used to generate the example code.
```.bash
make example
```

The generated project will contain the following code under the `./example` directory.
* Makefile
  * useful Makefile for installing tools building project etc. Issue `make` to display help output.
* .gitignore
  * git ignore for go project
* go.mod
  * go module setup, pass `--module` flag for setting the project module default `example.com/example`
* README.md
  * Project readme
* app/server/main.go
  * Sample Gin Server, with swagger init and comments
* api/<table name>.go
  * REST crud controllers
* dao/<table name>.go
  * DAO functions providing CRUD access to database
* model/<table name>.go
  * Structs representing a row for each database table


#### Generated Samples
* [GORM DAO CRUD Functions](./code_dao_gorm.md)
* [SQLX DAO CRUD Functions](./code_dao_sqlx.md)
* [Http CRUD Handlers](./code_http.md)
* [Model](./code_model.md)
* [Protobuf Definition](./code_protobuf.md)


The REST api server utilizes the Gin framework, GORM db api and Swag for providing swagger documentation
* [Gin](https://github.com/gin-gonic/gin)
* [Swaggo](https://github.com/swaggo/swag)
* [Gorm](https://github.com/jinzhu/gorm)
* [packr2](https://github.com/gobuffalo/packr)



## Supported Databases
Currently Supported,
- MariaDB
- MySQL
- PostgreSQL
- Microsoft SQL Server
- SQLite

Planned Support
- Oracle

## Supported Data Types

Most data types are supported, for Mysql, Postgres, SQLite and MS SQL. `gen` uses a mapping json file that can be used to add mapping types. By default, the internal mapping file is loaded and processed. If can be overwritten or additional types added by using the `--mapping=extra.json` command line option.

The default `mapping.json` file is located within the ./templates dir. Use `gen --save=./templates` to save the contents of the templates to `./templates`.
Below is a portion of the mapping file, showing the mapping for `varchar`.

```json
    {
      "sql_type": "varchar",
      "go_type": "string",
      "protobuf_type": "bytes",
      "guregu_type": "null.String",
      "go_nullable_type": "sql.NullString"
    }
```


## Advanced
The `gen` tool provides functionality to layout your own project format. Users have 2 options.
* Provide local templates with the `--templateDir=` option - this will generate code using the local templates. Templates can either be exported from `gen`
via the command `gen --save ./mytemplates`. This will save the embedded templates for local editing. Then you would specify the `--templateDir=` option when generating a project.

* Passing `--exec=../sample.gen` on the command line will load the `sample.gen` script and execute it. The script has access to the table information and other info passed to `gen`. This allows developers to customize the generation of code. You could loop through the list of tables and invoke
`GenerateTableFile` or  `GenerateFile`. You can also perform operations such as mkdir, copy, touch, pwd.


### Example - generate files from a template looping thru a map of tables.
Loop thru map of tables, key is the table name and value is ModelInfo. Creating a file using the table ModelInfo.

`tableInfos := map[string]*ModelInfo`

`GenerateTableFile(tableInfos map[string]*ModelInfo, tableName, templateFilename, outputDirectory, outputFileName string, formatOutput bool)`

```

{{ range $tableName , $table := .tableInfos }}
   {{$i := inc }}
   {{$name := toUpper $table.TableName -}}
   {{$filename  := printf "My%s.go" $name -}}

   {{ GenerateTableFile $.tableInfos $table.TableName  "custom.go.tmpl" "test" $filename true}}{{- end }}

```

### Example - generate file from a template.
`GenerateFile(templateFilename, outputDirectory, outputFileName string, formatOutput bool, overwrite bool)`
```

{{ GenerateFile "custom.md.tmpl" "test" "custom.md" false false }}

```

### Example - make a directory.
```

{{ mkdir "test/alex/test/mkdir" }}

```

### Example - touch a file.
```

{{ touch "test/alex/test/mkdir/alex.txt" }}

```

### Example - display working directory.
```

{{ pwd }}

```

### Example - copy a file or directory from source to a target directory.
copy function updated to provide --include and --exclude patterns. Patterns are processed in order, an include preceeding an exclude will take precedence. Multiple include and excludes can be specified. Files ending with .table.tmpl will be processed for each table. Output filenames will be stored in the proper directory, with a name of the table with the suffix of the template extension. Files ending with .tmpl will be processed as a template and the filename will be the name of the template stripped with the .tmpl suffix.

```

{{ copy "../_test" "test" }}

{{ copy "./backend" "test/backend" "--exclude .idea|commands" "--exclude go.sum"  "--include .*" }}

{{ copy "./backend" "test/backend" "--include backend" "--include go.mod" "--exclude .*"  }}


```


You can also populate the context used by templates with extra data by passing the `--context=<json file>` option. The json file will be used to populate the context used when parsing templates.

### File Generation
```gotemplate
// Loop through tables and print out table name and various forms of the table name

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


// GenerateTableFile(tableInfos map[string]*ModelInfo, tableName, templateFilename, outputDirectory, outputFileName string, formatOutput bool)
// GenerateFile(templateFilename, outputDirectory, outputFileName string, formatOutput bool, overwrite bool)

The following info is available within use of the exec template.


   "AdvancesSample"            string                         "\n{{ range $i, $table := .tables }}\n    {{$singular   := singular $table -}}\n    {{$plural     := pluralize $table -}}\n    {{$title      := title $table -}}\n    {{$lower      := toLower $table -}}\n    {{$lowerCamel := toLowerCamelCase $table -}}\n    {{$snakeCase  := toSnakeCase $table -}}\n    {{ printf \"[%-2d] %-20s %-20s %-20s %-20s %-20s %-20s %-20s\" $i $table $singular $plural $title $lower $lowerCamel $snakeCase}}{{- end }}\n\n\n{{ range $i, $table := .tables }}\n   {{$name := toUpper $table -}}\n   {{$filename  := printf \"My%s\" $name -}}\n   {{ printf \"[%-2d] %-20s %-20s\" $i $table $filename}}\n   {{ GenerateTableFile $table  \"custom.go.tmpl\" \"test\" $filename true}}\n{{- end }}\n"
   "Config"                    *dbmeta.Config                 &dbmeta.Config{SQLType:"sqlite3", SQLConnStr:"./example/sample.db", SQLDatabase:"main", Module:"github.com/alexj212/test", ModelPackageName:"model", ModelFQPN:"github.com/alexj212/test/model", AddJSONAnnotation:true, AddGormAnnotation:true, AddProtobufAnnotation:true, AddXMLAnnotation:true, AddDBAnnotation:true, UseGureguTypes:false, JSONNameFormat:"snake", XMLNameFormat:"snake", ProtobufNameFormat:"", DaoPackageName:"dao", DaoFQPN:"github.com/alexj212/test/dao", APIPackageName:"api", APIFQPN:"github.com/alexj212/test/api", GrpcPackageName:"", GrpcFQPN:"", Swagger:(*dbmeta.SwaggerInfoDetails)(0xc000ad0510), ServerPort:8080, ServerHost:"127.0.0.1", ServerScheme:"http", ServerListen:":8080", Verbose:false, OutDir:".", Overwrite:true, LineEndingCRLF:false, CmdLine:"/tmp/go-build271698611/b001/exe/readme --sqltype=sqlite3 --connstr ./example/sample.db --database main --table invoices", CmdLineWrapped:"/tmp/go-build271698611/b001/exe/readme \\\n    --sqltype=sqlite3 \\\n    --connstr \\\n    ./example/sample.db \\\n    --database \\\n    main \\\n    --table \\\n    invoices", CmdLineArgs:[]string{"/tmp/go-build271698611/b001/exe/readme", "--sqltype=sqlite3", "--connstr", "./example/sample.db", "--database", "main", "--table", "invoices"}, FileNamingTemplate:"{{.}}", ModelNamingTemplate:"{{FmtFieldName .}}", FieldNamingTemplate:"{{FmtFieldName (stringifyFirstChar .) }}", ContextMap:map[string]interface {}{"GenHelp":"Usage of gen:\n\tgen [-v] --sqltype=mysql --connstr \"user:password@/dbname\" --database <databaseName> --module=example.com/example [--json] [--gorm] [--guregu] [--generate-dao] [--generate-proj]\ngit fetch up\n           sqltype - sql database type such as [ mysql, mssql, postgres, sqlite, etc. ]\n\n\nOptions:\n  --sqltype=mysql                                          sql database type such as [ mysql, mssql, postgres, sqlite, etc. ]\n  -c, --connstr=nil                                        database connection string\n  -d, --database=nil                                       Database to for connection\n  -t, --table=                                             Table to build struct from\n  -x, --exclude=                                           Table(s) to exclude\n  --templateDir=                                           Template Dir\n  --save=                                                  Save templates to dir\n  --model=model                                            name to set for model package\n  --model_naming={{FmtFieldName .}}                        model naming template to name structs\n  --field_naming={{FmtFieldName (stringifyFirstChar .) }}  field naming template to name structs\n  --file_naming={{.}}                                      file_naming template to name files\n  --dao=dao                                                name to set for dao package\n  --api=api                                                name to set for api package\n  --grpc=grpc                                              name to set for grpc package\n  --out=.                                                  output dir\n  --module=example.com/example                             module path\n  --overwrite                                              Overwrite existing files (default)\n  --no-overwrite                                           disable overwriting files\n  --windows                                                use windows line endings in generated files\n  --no-color                                               disable color output\n  --context=                                               context file (json) to populate context with\n  --mapping=                                               mapping file (json) to map sql types to golang/protobuf etc\n  --exec=                                                  execute script for custom code generation\n  --json                                                   Add json annotations (default)\n  --no-json                                                Disable json annotations\n  --json-fmt=snake                                         json name format [snake | camel | lower_camel | none]\n  --xml                                                    Add xml annotations (default)\n  --no-xml                                                 Disable xml annotations\n  --xml-fmt=snake                                          xml name format [snake | camel | lower_camel | none]\n  --gorm                                                   Add gorm annotations (tags)\n  --protobuf                                               Add protobuf annotations (tags)\n  --proto-fmt=snake                                        proto name format [snake | camel | lower_camel | none]\n  --gogo-proto=                                            location of gogo import \n  --db                                                     Add db annotations (tags)\n  --guregu                                                 Add guregu null types\n  --copy-templates                                         Copy regeneration templates to project directory\n  --mod                                                    Generate go.mod in output dir\n  --makefile                                               Generate Makefile in output dir\n  --server                                                 Generate server app output dir\n  --generate-dao                                           Generate dao functions\n  --generate-proj                                          Generate project readme and gitignore\n  --rest                                                   Enable generating RESTful api\n  --run-gofmt                                              run gofmt on output dir\n  --listen=                                                listen address e.g. :8080\n  --scheme=http                                            scheme for server url\n  --host=localhost                                         host for server\n  --port=8080                                              port for server\n  --swagger_version=1.0                                    swagger version\n  --swagger_path=/                                         swagger base path\n  --swagger_tos=                                           swagger tos url\n  --swagger_contact_name=Me                                swagger contact name\n  --swagger_contact_url=http://me.com/terms.html           swagger contact url\n  --swagger_contact_email=me@me.com                        swagger contact email\n  -v, --verbose                                            Enable verbose output\n  --name_test=                                             perform name test using the --model_naming or --file_naming options\n  -h, --help                                               Show usage message\n  --version                                                Show version\n\n", "tableInfos":map[string]*dbmeta.ModelInfo{"invoices":(*dbmeta.ModelInfo)(0xc0001e94a0)}}, TemplateLoader:(dbmeta.TemplateLoader)(0x8a7e40), TableInfos:map[string]*dbmeta.ModelInfo(nil)}
   "DatabaseName"              string                         "main"
   "Dir"                       string                         "."
   "File"                      string                         "./README.md"
   "GenHelp"                   string                         "Usage of gen:\n\tgen [-v] --sqltype=mysql --connstr \"user:password@/dbname\" --database <databaseName> --module=example.com/example [--json] [--gorm] [--guregu] [--generate-dao] [--generate-proj]\ngit fetch up\n           sqltype - sql database type such as [ mysql, mssql, postgres, sqlite, etc. ]\n\n\nOptions:\n  --sqltype=mysql                                          sql database type such as [ mysql, mssql, postgres, sqlite, etc. ]\n  -c, --connstr=nil                                        database connection string\n  -d, --database=nil                                       Database to for connection\n  -t, --table=                                             Table to build struct from\n  -x, --exclude=                                           Table(s) to exclude\n  --templateDir=                                           Template Dir\n  --save=                                                  Save templates to dir\n  --model=model                                            name to set for model package\n  --model_naming={{FmtFieldName .}}                        model naming template to name structs\n  --field_naming={{FmtFieldName (stringifyFirstChar .) }}  field naming template to name structs\n  --file_naming={{.}}                                      file_naming template to name files\n  --dao=dao                                                name to set for dao package\n  --api=api                                                name to set for api package\n  --grpc=grpc                                              name to set for grpc package\n  --out=.                                                  output dir\n  --module=example.com/example                             module path\n  --overwrite                                              Overwrite existing files (default)\n  --no-overwrite                                           disable overwriting files\n  --windows                                                use windows line endings in generated files\n  --no-color                                               disable color output\n  --context=                                               context file (json) to populate context with\n  --mapping=                                               mapping file (json) to map sql types to golang/protobuf etc\n  --exec=                                                  execute script for custom code generation\n  --json                                                   Add json annotations (default)\n  --no-json                                                Disable json annotations\n  --json-fmt=snake                                         json name format [snake | camel | lower_camel | none]\n  --xml                                                    Add xml annotations (default)\n  --no-xml                                                 Disable xml annotations\n  --xml-fmt=snake                                          xml name format [snake | camel | lower_camel | none]\n  --gorm                                                   Add gorm annotations (tags)\n  --protobuf                                               Add protobuf annotations (tags)\n  --proto-fmt=snake                                        proto name format [snake | camel | lower_camel | none]\n  --gogo-proto=                                            location of gogo import \n  --db                                                     Add db annotations (tags)\n  --guregu                                                 Add guregu null types\n  --copy-templates                                         Copy regeneration templates to project directory\n  --mod                                                    Generate go.mod in output dir\n  --makefile                                               Generate Makefile in output dir\n  --server                                                 Generate server app output dir\n  --generate-dao                                           Generate dao functions\n  --generate-proj                                          Generate project readme and gitignore\n  --rest                                                   Enable generating RESTful api\n  --run-gofmt                                              run gofmt on output dir\n  --listen=                                                listen address e.g. :8080\n  --scheme=http                                            scheme for server url\n  --host=localhost                                         host for server\n  --port=8080                                              port for server\n  --swagger_version=1.0                                    swagger version\n  --swagger_path=/                                         swagger base path\n  --swagger_tos=                                           swagger tos url\n  --swagger_contact_name=Me                                swagger contact name\n  --swagger_contact_url=http://me.com/terms.html           swagger contact url\n  --swagger_contact_email=me@me.com                        swagger contact email\n  -v, --verbose                                            Enable verbose output\n  --name_test=                                             perform name test using the --model_naming or --file_naming options\n  -h, --help                                               Show usage message\n  --version                                                Show version\n\n"
   "NonPrimaryKeyNamesList"    []string                       []string{"CustomerId", "InvoiceDate", "BillingAddress", "BillingCity", "BillingState", "BillingCountry", "BillingPostalCode", "Total"}
   "NonPrimaryKeysJoined"      string                         "CustomerId,InvoiceDate,BillingAddress,BillingCity,BillingState,BillingCountry,BillingPostalCode,Total"
   "Parent"                    string                         "."
   "PrimaryKeyNamesList"       []string                       []string{"InvoiceId"}
   "PrimaryKeysJoined"         string                         "InvoiceId"
   "ReleaseHistory"            string                         "- v0.9.27 (08/04/2020)\n    - Updated '--exec' mode to provide various functions for processing\n    - copy function updated to provide --include and --exclude patterns. Patterns are processed in order, an include preceeding an exclude will take precedence. Multiple include and excludes can be specified. Files ending with .table.tmpl will be processed for each table. Output filenames will be stored in the proper directory, with a name of the table with the suffix of the template extension. Files ending with .tmpl will be processed as a template and the filename will be the name of the template stripped with the .tmpl suffix. \n    - When processing templates, files generated with a .go extension will be formatted with the go fmt.\n- v0.9.26 (07/31/2020)\n    - Release scripting\n    - Added custom script functions to copy, mkdir, touch, pwd\n    - Fixed custom script exec example\n- v0.9.25 (07/26/2020)\n    - Adhere json-fmt flag for all JSON response so when camel or lower_camel is specified, fields name in GetAll variant and DDL info will also have the same name format\n    - Fix: Build information embedded through linker in Makefile is not consistent with the variable defined in main file.\n    - Added --scheme and --listen options. This allows compiled binary to be used behind reverse proxy.\n    - In addition, template for generating URL was fixed, i.e. when PORT is 80, then PORT is omitted from URL segment.\n- v0.9.24 (07/13/2020)\n    - Fixed array bounds issue parsing mysql db meta\n- v0.9.23 (07/10/2020)\n    - Added postgres types: bigserial, serial, smallserial, bigserial, float4 to mapping.json\n- v0.9.22 (07/08/2020)\n    - Modified gogo.proto check to use GOPATH not hardcoded.\n    - Updated gen to error exit on first error encountered\n    - Added color output for error\n    - Added --no-color option for non colorized output\n- v0.9.21 (07/07/2020)\n    - Repacking templates, update version number in info.\n- v0.9.20 (07/07/2020)\n    - Fixed render error in router.go.tmpl\n    - upgraded project to use go.mod 1.14\n- v0.9.19 (07/07/2020)\n    - Added --windows flag to write files with CRLF windows line endings, otherwise they are all unix based LF line endings\n- v0.9.18 (06/30/2020)\n    - Fixed naming in templates away from hard coded model package.\n- v0.9.17 (06/30/2020)\n    - Refactored template loading, to better report error in template\n    - Added option to run gofmt on output directory\n- v0.9.16 (06/29/2020)\n    - Fixes to router.go.tmpl from calvinchengx\n    - Added postgres db support for inet and timestamptz\n- v0.9.15 (06/23/2020)\n    - Code cleanup using gofmt name suggestions.\n    - Template updates for generated code cleanup using gofmt name suggestions.\n- v0.9.14 (06/23/2020)\n    - Added model comment on field line if available from database.\n    - Added exposing TableInfo via api call.\n- v0.9.13 (06/22/2020)\n    - fixed closing of connections via defer\n    - bug fixes in sqlx generated code\n- v0.9.12 (06/14/2020)\n    - SQLX changed MustExec to Exec and checking/returning error\n    - Updated field renaming if duplicated, need more elegant renaming solution.\n    - Added exclude to test.sh\n- v0.9.11 (06/13/2020)\n    - Added ability to pass field, model and file naming format\n    - updated test scripts\n    - Fixed sqlx sql query placeholders\n- v0.9.10 (06/11/2020)\n    - Bug fix with retrieving varchar length from mysql\n    - Added support for mysql unsigned decimal - maps to float\n- v0.9.9 (06/11/2020)\n    - Fixed issue with mysql and table named `order`\n    - Fixed internals in GetAll generation in gorm and sqlx.\n- v0.9.8 (06/10/2020)\n    - Added ability to set file naming convention for models, dao, apis and grpc  `--file_naming={{.}}`\n    - Added ability to set struct naming convention `--model_naming={{.}}`\n    - Fixed bug with Makefile generation removing quoted conn string in `make regen`\n- v0.9.7 (06/09/2020)\n    - Added grpc server generation - WIP (looking for code improvements)\n    - Added ability to exclude tables\n    - Added support for unsigned from mysql ddl.\n- v0.9.6 (06/08/2020)\n    - Updated SQLX codegen\n    - Updated templates to split code gen functions into seperate files\n    - Added code_dao_gorm, code_dao_sqlx to be generated from templates\n- v0.9.5 (05/16/2020)\n    - Added SQLX codegen by default, split dao templates.\n    - Renamed templates\n- v0.9.4 (05/15/2020)\n    - Documentation updates, samples etc.\n- v0.9.3 (05/14/2020)\n    - Template bug fixes, when using custom api, dao and model package.\n    - Set primary key if not set to the first column\n    - Skip code gen if primary key column is not int or string\n    - validated codegen for mysql, mssql, postgres and sqlite3\n    - Fixed file naming if table ends with _test.go renames to _tst.go\n    - Fix for duplicate field names in struct due to renaming\n    - Added Notes for columns and tables for situations where a primary key is set since not defined in db\n    - Fixed issue when model contained field that had were named the same as funcs within model.\n- v0.9.2 (05/12/2020)\n    - Code cleanup gofmt, etc.\n- v0.9.1 (05/12/2020)\n- v0.9 (05/12/2020)\n    - updated db meta data loading fetching default values\n    - added default value to GORM tags\n    - Added protobuf .proto generation\n    - Added test app to display meta data\n    - Cleanup DDL generation\n    - Added support for varchar2, datetime2, float8, USER_DEFINED\n- v0.5\n"
   "ShortStructName"           string                         "i"
   "StructName"                string                         "Invoices"
   "SwaggerInfo"               *dbmeta.SwaggerInfoDetails     &dbmeta.SwaggerInfoDetails{Version:"1.0.0", Host:"127.0.0.1:8080", BasePath:"/", Title:"Sample CRUD api for main db", Description:"Sample CRUD api for main db", TOS:"My Custom TOS", ContactName:"", ContactURL:"", ContactEmail:""}
   "TableInfo"                 *dbmeta.ModelInfo              &dbmeta.ModelInfo{Index:0, IndexPlus1:1, PackageName:"model", StructName:"Invoices", ShortStructName:"i", TableName:"invoices", Fields:[]string{"//[ 0] InvoiceId                                      integer              null: false  primary: true   isArray: false  auto: true   col: integer         len: -1      default: []\n    InvoiceID int32 `gorm:\"primary_key;AUTO_INCREMENT;column:InvoiceId;type:integer;\" json:\"invoice_id\" xml:\"invoice_id\" db:\"InvoiceId\" protobuf:\"int32,0,opt,name=InvoiceId\"`", "//[ 1] CustomerId                                     integer              null: false  primary: false  isArray: false  auto: false  col: integer         len: -1      default: []\n    CustomerID int32 `gorm:\"column:CustomerId;type:integer;\" json:\"customer_id\" xml:\"customer_id\" db:\"CustomerId\" protobuf:\"int32,1,opt,name=CustomerId\"`", "//[ 2] InvoiceDate                                    datetime             null: false  primary: false  isArray: false  auto: false  col: datetime        len: -1      default: []\n    InvoiceDate time.Time `gorm:\"column:InvoiceDate;type:datetime;\" json:\"invoice_date\" xml:\"invoice_date\" db:\"InvoiceDate\" protobuf:\"google.protobuf.Timestamp,2,opt,name=InvoiceDate\"`", "//[ 3] BillingAddress                                 nvarchar(70)         null: true   primary: false  isArray: false  auto: false  col: nvarchar        len: 70      default: []\n    BillingAddress sql.NullString `gorm:\"column:BillingAddress;type:nvarchar;size:70;\" json:\"billing_address\" xml:\"billing_address\" db:\"BillingAddress\" protobuf:\"string,3,opt,name=BillingAddress\"`", "//[ 4] BillingCity                                    nvarchar(40)         null: true   primary: false  isArray: false  auto: false  col: nvarchar        len: 40      default: []\n    BillingCity sql.NullString `gorm:\"column:BillingCity;type:nvarchar;size:40;\" json:\"billing_city\" xml:\"billing_city\" db:\"BillingCity\" protobuf:\"string,4,opt,name=BillingCity\"`", "//[ 5] BillingState                                   nvarchar(40)         null: true   primary: false  isArray: false  auto: false  col: nvarchar        len: 40      default: []\n    BillingState sql.NullString `gorm:\"column:BillingState;type:nvarchar;size:40;\" json:\"billing_state\" xml:\"billing_state\" db:\"BillingState\" protobuf:\"string,5,opt,name=BillingState\"`", "//[ 6] BillingCountry                                 nvarchar(40)         null: true   primary: false  isArray: false  auto: false  col: nvarchar        len: 40      default: []\n    BillingCountry sql.NullString `gorm:\"column:BillingCountry;type:nvarchar;size:40;\" json:\"billing_country\" xml:\"billing_country\" db:\"BillingCountry\" protobuf:\"string,6,opt,name=BillingCountry\"`", "//[ 7] BillingPostalCode                              nvarchar(10)         null: true   primary: false  isArray: false  auto: false  col: nvarchar        len: 10      default: []\n    BillingPostalCode sql.NullString `gorm:\"column:BillingPostalCode;type:nvarchar;size:10;\" json:\"billing_postal_code\" xml:\"billing_postal_code\" db:\"BillingPostalCode\" protobuf:\"string,7,opt,name=BillingPostalCode\"`", "//[ 8] Total                                          numeric              null: false  primary: false  isArray: false  auto: false  col: numeric         len: -1      default: []\n    Total float64 `gorm:\"column:Total;type:numeric;\" json:\"total\" xml:\"total\" db:\"Total\" protobuf:\"float,8,opt,name=Total\"`"}, DBMeta:(*dbmeta.dbTableMeta)(0xc000133b00), Instance:(*struct { BillingState string "json:\"billing_state\""; BillingCountry string "json:\"billing_country\""; BillingPostalCode string "json:\"billing_postal_code\""; CustomerID int "json:\"customer_id\""; InvoiceDate time.Time "json:\"invoice_date\""; BillingAddress string "json:\"billing_address\""; BillingCity string "json:\"billing_city\""; Total float64 "json:\"total\""; InvoiceID int "json:\"invoice_id\"" })(0xc000a85700), CodeFields:[]*dbmeta.FieldInfo{(*dbmeta.FieldInfo)(0xc00025c640), (*dbmeta.FieldInfo)(0xc00025c780), (*dbmeta.FieldInfo)(0xc00025c8c0), (*dbmeta.FieldInfo)(0xc00025ca00), (*dbmeta.FieldInfo)(0xc00025cb40), (*dbmeta.FieldInfo)(0xc00025cc80), (*dbmeta.FieldInfo)(0xc00025cdc0), (*dbmeta.FieldInfo)(0xc00025cf00), (*dbmeta.FieldInfo)(0xc00025d040)}}
   "TableName"                 string                         "invoices"
   "apiFQPN"                   string                         "github.com/alexj212/test/api"
   "apiPackageName"            string                         "api"
   "daoFQPN"                   string                         "github.com/alexj212/test/dao"
   "daoPackageName"            string                         "dao"
   "delSql"                    string                         "DELETE FROM `invoices` where InvoiceId = ?"
   "insertSql"                 string                         "INSERT INTO `invoices` ( CustomerId,  InvoiceDate,  BillingAddress,  BillingCity,  BillingState,  BillingCountry,  BillingPostalCode,  Total) values ( ?, ?, ?, ?, ?, ?, ?, ? )"
   "modelFQPN"                 string                         "github.com/alexj212/test/model"
   "modelPackageName"          string                         "model"
   "module"                    string                         "github.com/alexj212/test"
   "outDir"                    string                         "."
   "selectMultiSql"            string                         "SELECT * FROM `invoices`"
   "selectOneSql"              string                         "SELECT * FROM `invoices` WHERE InvoiceId = ?"
   "serverHost"                string                         "127.0.0.1"
   "serverListen"              string                         ":8080"
   "serverPort"                int                            8080
   "serverScheme"              string                         "http"
   "sqlConnStr"                string                         "./example/sample.db"
   "sqlType"                   string                         "sqlite3"
   "tableInfos"                map[string]*dbmeta.ModelInfo   map[string]*dbmeta.ModelInfo{"invoices":(*dbmeta.ModelInfo)(0xc0001e94a0)}
   "updateSql"                 string                         "UPDATE `invoices` set CustomerId = ?, InvoiceDate = ?, BillingAddress = ?, BillingCity = ?, BillingState = ?, BillingCountry = ?, BillingPostalCode = ?, Total = ? WHERE InvoiceId = ?"


```

### Struct naming
The ability exists to set a template that will be used for generating a struct name. By passing the flag `--model_naming={{.}}`
The struct will be named the table name. Various functions can be used in the template to modify the name such as

You can use the argument `--name_test=user` in conjunction with the `--model_naming` or `--file_name`, to view what the naming would be.

| Function   | Table Name  | Output
|---|---|---|
|singular   |Users   | `User`  |
|pluralize   |Users   | `Users`  |
|title   |Users   | `Users`  |
|toLower   |Users   | `users`  |
|toUpper   |Users   | `USERS`  |
|toLowerCamelCase   |Users   | `users`  |
|toUpperCamelCase   |Users   | `Users`  |
|toSnakeCase   |Users   | `users`  |




### Struct naming Examples
Table Name: registration_source

| Model Naming Format  | Generated Struct Name
|---|---|
|`{{.}}`   | registration_source   |
|`Struct{{.}}`   | Structregistration_source   |
|`Struct{{ singular .}}`   | Structregistration_source   |
|`Struct{{ toLowerCamelCase .}}`   | Structregistration_source   |
|`Struct{{ toUpperCamelCase .}}`   | StructRegistration_source   |
|`Struct{{ toSnakeCase .}}`   | Structregistration_source   |
|`Struct{{ toLowerCamelCase .}}`   | Structtable_registration_source   |
|`Struct{{ toUpperCamelCase .}}`   | StructTable_registration_source   |
|`Struct{{ toSnakeCase .}}`   | Structtable_registration_source   |
|`Struct{{ toSnakeCase ( replace . "table_" "") }}`   | Structregistration_source   |


## Notes
- MySql, Mssql, Postgres and Sqlite have a database metadata fetcher that will query the db, and update the auto increment, primary key and nullable info for the gorm annotation.
- Tables that have a non-standard primary key (NON integer based or String) the table will be ignored.

## DB Meta Data Loading
| DB   | Type  | Nullable  | Primary Key  | Auto Increment  | Column Len | default Value| create ddl
|---|---|---|---|---|---|---|---|
|sqlite   |y   | y  | y  | y  | y | y| y
|postgres   |y   | y  | y  | y  | y | y| n
|mysql   |y   | y  | y  | y  | y | y| y
|ms sql   |y   | y  | y  | y  | y | y| n

## Version History
- v0.9.27 (08/04/2020)
    - Updated '--exec' mode to provide various functions for processing
    - copy function updated to provide --include and --exclude patterns. Patterns are processed in order, an include preceeding an exclude will take precedence. Multiple include and excludes can be specified. Files ending with .table.tmpl will be processed for each table. Output filenames will be stored in the proper directory, with a name of the table with the suffix of the template extension. Files ending with .tmpl will be processed as a template and the filename will be the name of the template stripped with the .tmpl suffix. 
    - When processing templates, files generated with a .go extension will be formatted with the go fmt.
- v0.9.26 (07/31/2020)
    - Release scripting
    - Added custom script functions to copy, mkdir, touch, pwd
    - Fixed custom script exec example
- v0.9.25 (07/26/2020)
    - Adhere json-fmt flag for all JSON response so when camel or lower_camel is specified, fields name in GetAll variant and DDL info will also have the same name format
    - Fix: Build information embedded through linker in Makefile is not consistent with the variable defined in main file.
    - Added --scheme and --listen options. This allows compiled binary to be used behind reverse proxy.
    - In addition, template for generating URL was fixed, i.e. when PORT is 80, then PORT is omitted from URL segment.
- v0.9.24 (07/13/2020)
    - Fixed array bounds issue parsing mysql db meta
- v0.9.23 (07/10/2020)
    - Added postgres types: bigserial, serial, smallserial, bigserial, float4 to mapping.json
- v0.9.22 (07/08/2020)
    - Modified gogo.proto check to use GOPATH not hardcoded.
    - Updated gen to error exit on first error encountered
    - Added color output for error
    - Added --no-color option for non colorized output
- v0.9.21 (07/07/2020)
    - Repacking templates, update version number in info.
- v0.9.20 (07/07/2020)
    - Fixed render error in router.go.tmpl
    - upgraded project to use go.mod 1.14
- v0.9.19 (07/07/2020)
    - Added --windows flag to write files with CRLF windows line endings, otherwise they are all unix based LF line endings
- v0.9.18 (06/30/2020)
    - Fixed naming in templates away from hard coded model package.
- v0.9.17 (06/30/2020)
    - Refactored template loading, to better report error in template
    - Added option to run gofmt on output directory
- v0.9.16 (06/29/2020)
    - Fixes to router.go.tmpl from calvinchengx
    - Added postgres db support for inet and timestamptz
- v0.9.15 (06/23/2020)
    - Code cleanup using gofmt name suggestions.
    - Template updates for generated code cleanup using gofmt name suggestions.
- v0.9.14 (06/23/2020)
    - Added model comment on field line if available from database.
    - Added exposing TableInfo via api call.
- v0.9.13 (06/22/2020)
    - fixed closing of connections via defer
    - bug fixes in sqlx generated code
- v0.9.12 (06/14/2020)
    - SQLX changed MustExec to Exec and checking/returning error
    - Updated field renaming if duplicated, need more elegant renaming solution.
    - Added exclude to test.sh
- v0.9.11 (06/13/2020)
    - Added ability to pass field, model and file naming format
    - updated test scripts
    - Fixed sqlx sql query placeholders
- v0.9.10 (06/11/2020)
    - Bug fix with retrieving varchar length from mysql
    - Added support for mysql unsigned decimal - maps to float
- v0.9.9 (06/11/2020)
    - Fixed issue with mysql and table named `order`
    - Fixed internals in GetAll generation in gorm and sqlx.
- v0.9.8 (06/10/2020)
    - Added ability to set file naming convention for models, dao, apis and grpc  `--file_naming={{.}}`
    - Added ability to set struct naming convention `--model_naming={{.}}`
    - Fixed bug with Makefile generation removing quoted conn string in `make regen`
- v0.9.7 (06/09/2020)
    - Added grpc server generation - WIP (looking for code improvements)
    - Added ability to exclude tables
    - Added support for unsigned from mysql ddl.
- v0.9.6 (06/08/2020)
    - Updated SQLX codegen
    - Updated templates to split code gen functions into seperate files
    - Added code_dao_gorm, code_dao_sqlx to be generated from templates
- v0.9.5 (05/16/2020)
    - Added SQLX codegen by default, split dao templates.
    - Renamed templates
- v0.9.4 (05/15/2020)
    - Documentation updates, samples etc.
- v0.9.3 (05/14/2020)
    - Template bug fixes, when using custom api, dao and model package.
    - Set primary key if not set to the first column
    - Skip code gen if primary key column is not int or string
    - validated codegen for mysql, mssql, postgres and sqlite3
    - Fixed file naming if table ends with _test.go renames to _tst.go
    - Fix for duplicate field names in struct due to renaming
    - Added Notes for columns and tables for situations where a primary key is set since not defined in db
    - Fixed issue when model contained field that had were named the same as funcs within model.
- v0.9.2 (05/12/2020)
    - Code cleanup gofmt, etc.
- v0.9.1 (05/12/2020)
- v0.9 (05/12/2020)
    - updated db meta data loading fetching default values
    - added default value to GORM tags
    - Added protobuf .proto generation
    - Added test app to display meta data
    - Cleanup DDL generation
    - Added support for varchar2, datetime2, float8, USER_DEFINED
- v0.5



## Contributors
- [alexj212](https://github.com/alexj212) -  a big thanks to alexj212 for his contributions

See more contributors: [contributors](https://github.com/smallnest/gen/graphs/contributors)

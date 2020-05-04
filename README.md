### gen 

[![License](https://img.shields.io/badge/License-Apache%203.0-blue.svg)](https://opensource.org/licenses/Apache-3.0) [![GoDoc](https://godoc.org/github.com/smallnest/gen?status.png)](http://godoc.org/github.com/smallnest/gen)  [![travis](https://travis-ci.org/smallnest/gen.svg?branch=master)](https://travis-ci.org/smallnest/gen) [![Go Report Card](https://goreportcard.com/badge/github.com/smallnest/gen)](https://goreportcard.com/report/github.com/smallnest/gen)

The gen tool produces a CRUD (Create, read, update and delete) REST api project template from a given database. The gen tool will
connect to the db connection string analyze the database and generate the code based on the flags provided. 

By reading details from the database about the column structure, gen generates a go compatible struct type
with the required column names, data types, and annotations.

It supports [gorm](https://github.com/jinzhu/gorm) tags and implements some usable methods. Generated data types include support for nullable columns [sql.NullX types](https://golang.org/pkg/database/sql/#NullBool) or [guregu null.X types](https://github.com/guregu/null)
and the expected basic built in go types.

`gen` is based / inspired by the work of Seth Shelnutt's [db2struct](https://github.com/Shelnutt2/db2struct), and Db2Struct is based/inspired by the work of ChimeraCoder's gojson package [gojson](https://github.com/ChimeraCoder/gojson).


## Binary Installation
```BASH
## install gen tool (should be installed to ~/go/bin, make sure ~/go/bin is in your path.
$ go get -u github.com/smallnest/gen

## download sample sqlite database
$ wget https://github.com/smallnest/gen/raw/master/example/sample.db

## generate code based on the sqlite database (project will be containied within the ./example dir)
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


## Open a browser to http://localhost:8080/swagger/index.html

## Use wget/curl/httpie to fetch via command line
http http://localhost:8080/albums
curl http://localhost:8080/artists

```


## Usage
```BASH
$ ./gen --help
Usage of gen:
        gen [-v] --connstr "user:password@/dbname" --package pkgName --database databaseName --table tableName [--json] [--gorm] [--guregu]
Options:
  --sqltype=mysql                                 sql database type such as mysql, postgres, etc.
  -c, --connstr=nil                               database connection string
  -d, --database=nil                              Database to for connection
  -t, --table=                                    Table to build struct from
  --templateDir=                                  Template Dir
  --save=                                         Save templates to dir
  --model=model                                   name to set for model package
  --dao=dao                                       name to set for dao package
  --api=api                                       name to set for api package
  --out=.                                         output dir
  --module=example.com/example                    module path
  --overwrite                                     Overwrite existing files (default)
  --no-overwrite                                  disable overwriting files
  --exec=                                         execute script for custom code generation
  --json                                          Add json annotations (default)
  --no-json                                       Disable json annotations
  --json-fmt=snake                                json name format [snake | camel | lower_camel | none]
  --gorm                                          Add gorm annotations (tags)
  --protobuf                                      Add protobuf annotations (tags)
  --db                                            Add db annotations (tags)
  --guregu                                        Add guregu null types
  --mod                                           Generate go.mod in output dir
  --makefile                                      Generate Makefile in output dir
  --server                                        Generate server app output dir
  --generate-dao                                  Generate dao functions
  --generate-proj                                 Generate project README.md and .gitignore
  --context=                                      context file (json) to populate context with
  --mapping=                                      mapping file (json) to map sql types to golang/protobuf etc
  --host=localhost                                host for server
  --port=8080                                     port for server
  --rest                                          Enable generating RESTful api
  --copy-templates                                Copy regeneration templates to project directory
  --swagger_version=1.0                           swagger version
  --swagger_path=/                                swagger base path
  --swagger_tos=                                  swagger tos url
  --swagger_contact_name=Me                       swagger contact name
  --swagger_contact_url=http://me.com/terms.html  swagger contact url
  --swagger_contact_email=me@me.com               swagger contact email
  -v, --verbose                                   Enable verbose output
  -h, --help                                      Show usage message
  --version                                       Show version

```

## Building
The project contains a makefile for easy building and common tasks.
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
* api/*.go
  * REST crud controllers
* dao/*.go
  * DAO functions providing CRUD access to database
* model/*.go
  * Structs representing a row for each database table


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

#### Supported Data Types

Most datatypes are supported, for Mysql, Postgres, SQLite and MS SQL. `gen` uses a mapping json file that can be used to add mapping types. By default the internal mapping file is loaded and processed. If can be overwritten or additional types added by using the `--mapping=extra.json` command line option.

The default `mapping.json` file is located within the templates dir. Use `gen --save=./templates` to save the contents of the templates to `./templates`. 
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
 

### Advanced
The `gen` tool provides functionality to layout your own project format. Users have 2 options.
* Provide local templates with the `--templateDir=` option - this will generate code using the local templates. Templates can either be exported from `gen`
via the command `gen --save ./mytemplates`. This will save the embedded templates for local editing. Then you would specify the `--templateDir=` option when generating a project.

* Passing `--exec=../sample.gen` on the command line will load the `sample.gen` script and execute it. The script has access to the table information and other info passed to `gen`. This allows developers to customize the generation of code. You could loop through the list of tables and invoke  
`GenerateTableFile` or  `GenerateFile`. 

You can also populate the context used by templates with extra data by passing the `--contect=<json file>` option. The json file will be used to populate the context used when parsing templates.  



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

// GenerateTableFile(tableName, templateFilename, outputDirectory, outputFileName string, formatOutput bool)
// GenerateFile(templateFilename, outputDirectory, outputFileName string, formatOutput bool) string

The following info is available within use of the exec template.

   "CommandLine"          [string] "/tmp/go-build728666148/b001/exe/gen --sqltype=sqlite3 --connstr ./sample.db --database main --module github.com/alexj212/generated --verbose --overwrite --out ./ --exec=../sample.gen"
   "DatabaseName"         [string] "main"
   "apiFQPN"              [string] "github.com/alexj212/generated/api"
   "apiPackageName"       [string] "api"
   "daoFQPN"              [string] "github.com/alexj212/generated/dao"
   "daoPackageName"       [string] "dao"
   "modelFQPN"            [string] "github.com/alexj212/generated/model"
   "modelPackageName"     [string] "model"
   "module"               [string] "github.com/alexj212/generated"
   "outDir"               [string] "./"
   "serverHost"           [string] "localhost"
   "serverPort"           [int] 8080
   "sqlConnStr"           [string] "./sample.db"
   "sqlType"              [string] "sqlite3"
   "structs"              [[]string] []string{"Album", "Artist", "Customer", "Employee", "Genre", "Invoice", "InvoiceItem", "MediaType", "Playlist", "PlaylistTrack", "Track"}
   "SwaggerInfo"          [*main.swaggerInfo] &main.swaggerInfo{Version:"1.0", Host:"localhost:8080", BasePath:"/", Title:"Sample CRUD api for main db", Description:"Sample CRUD api for main db", TOS:"", ContactName:"Me", ContactUrl:"http://me.com/terms.html", ContactEmail:"me@me.com"}
   "tableInfos"           [map[string]*dbmeta.ModelInfo] map[string]*dbmeta.ModelInfo{"albums":(*dbmeta.ModelInfo)(0xc00031fa40), "artists":(*dbmeta.ModelInfo)(0xc00031fb90), "customers":(*dbmeta.ModelInfo)(0xc00031fce0), "employees":(*dbmeta.ModelInfo)(0xc00031fd50), "genres":(*dbmeta.ModelInfo)(0xc00031fdc0), "invoice_items":(*dbmeta.ModelInfo)(0xc00031fea0), "invoices":(*dbmeta.ModelInfo)(0xc00031fe30), "media_types":(*dbmeta.ModelInfo)(0xc00031ff10), "playlist_track":(*dbmeta.ModelInfo)(0xc0001121c0), "playlists":(*dbmeta.ModelInfo)(0xc00031ff80), "tracks":(*dbmeta.ModelInfo)(0xc000112230)}
   "tables"               [[]string] []string{"albums", "artists", "customers", "employees", "genres", "invoices", "invoice_items", "media_types", "playlists", "playlist_track", "tracks"}
```





## Notes
- MySql, Mssql, Postgres and Sqlite have a database metadata fetcher that will query the db, and update the auto increment, primary key and nullable info for the gorm annotation.


 

### gen 

[![License](https://img.shields.io/badge/License-Apache%203.0-blue.svg)](https://opensource.org/licenses/Apache-3.0) [![GoDoc](https://godoc.org/github.com/smallnest/gen?status.png)](http://godoc.org/github.com/smallnest/gen)  [![travis](https://travis-ci.org/smallnest/gen.svg?branch=master)](https://travis-ci.org/smallnest/gen) [![Go Report Card](https://goreportcard.com/badge/github.com/smallnest/gen)](https://goreportcard.com/report/github.com/smallnest/gen)

The gen tool produces a CRUD (Create, read, update and delete) REST api project template from a given database. The gen tool will
connect to the db connection string analyze the database and generate the code based on the flags provided. 

By reading details from the database about the column structure, gen generates a go compatible struct type
with the required column names, data types, and annotations.

It supports [gorm](https://github.com/jinzhu/gorm) tags and implements some usable methods. Generated datatypes include support for nullable columns [sql.NullX types](https://golang.org/pkg/database/sql/#NullBool) or [guregu null.X types](https://github.com/guregu/null)
and the expected basic built in go types.

`gen` is based / inspired by the work of Seth Shelnutt's [db2struct](https://github.com/Shelnutt2/db2struct), and Db2Struct is based/inspired by the work of ChimeraCoder's gojson package [gojson](https://github.com/ChimeraCoder/gojson).


## Binary Installation
```BASH
go get -u github.com/smallnest/gen

```

## Usage
```BASH
gen --sqltype=sqlite3 \
   	--connstr "./example/sample.db" \
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
   	--overwrite
```


## Options
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
  --json                                          Add json annotations (default)
  --no-json                                       Disable json annotations
  --json-fmt=snake                                json name format [snake | camel | lower_camel | none
  --gorm                                          Add gorm annotations (tags)
  --guregu                                        Add guregu null types
  --mod                                           Generate go.mod in output dir
  --makefile                                      Generate Makefile in output dir
  --server                                        Generate server app output dir
  --overwrite                                     Overwrite existing files (default)
  --no-overwrite                                  disable overwriting files
  --host=localhost                                host for server
  --port=8080                                     port for server
  --rest                                          Enable generating RESTful api
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
* `make example` - run the gen process on the example sqlite db located in ./examples place the sources in ./example
Other targets exist for dev tasks.

## Example
The projects provides a sample SQLite database in the `./example` directory. From the project `Makefile` can be used to generate the example code. 
```.bash
make example
``` 

The generated project will contain the following code under the `./example` directory.
* Makefile
  * useful Makefile for installing tools building project etc. Issue `make` to display help
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

Currently Supported
- MariaDB
- MySQL
- PostgreSQL
- Microsoft SQL Server
- SQLite

Planned Support
- Oracle


### MariaDB/MySQL

Structures are created by querying the INFORMATION_SCHEMA.Columns table and then formatting the types, column names,
and metadata to create a usable go compatible struct type.


#### Supported Datatypes

Currently only a limited number of datatypes are supported. Initial support includes:
-  tinyint (sql.NullInt64 or null.Int)
-  int      (sql.NullInt64 or null.Int)
-  smallint      (sql.NullInt64 or null.Int)
-  mediumint      (sql.NullInt64 or null.Int)
-  bigint (sql.NullInt64 or null.Int)
-  decimal (sql.NullFloat64 or null.Float)
-  float (sql.NullFloat64 or null.Float)
-  double (sql.NullFloat64 or null.Float)
-  datetime (null.Time)
-  time  (null.Time)
-  date (null.Time)
-  timestamp (null.Time)
-  var (sql.String or null.String)
-  enum (sql.String or null.String)
-  varchar (sql.String or null.String)
-  longtext (sql.String or null.String)
-  mediumtext (sql.String or null.String)
-  text (sql.String or null.String)
-  tinytext (sql.String or null.String)
-  binary
-  blob
-  longblob
-  mediumblob
-  varbinary


## Issues

- Postgres and SQLite driver support for sql.ColumnType.Nullable() ([#3](https://github.com/smallnest/gen/issues/3))
- Can not distinguish primay key of tables. Only set the first field as primay key. So you need to change it in some cases.

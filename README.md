### gen 

[![License](https://img.shields.io/:license-apache-3.0-blue.svg)](https://opensource.org/licenses/Apache-3.0) [![GoDoc](https://godoc.org/github.com/smallnest/gen?status.png)](http://godoc.org/github.com/smallnest/gen)  [![travis](https://travis-ci.org/smallnest/gen.svg?branch=master)](https://travis-ci.org/smallnest/gen) [![Go Report Card](https://goreportcard.com/badge/github.com/smallnest/gen)](https://goreportcard.com/report/github.com/smallnest/gen)

The gen tool produces golang structs from a given database for use in a .go file.
It supports [gorm](https://github.com/jinzhu/gorm) tags and implements some usable methods.
It can also generate RESTful api for those structs.

By reading details from the database about the column structure, gen generates a go compatible struct type
with the required column names, data types, and annotations.

Generated datatypes include support for nullable columns [sql.NullX types](https://golang.org/pkg/database/sql/#NullBool) or [guregu null.X types](https://github.com/guregu/null)
and the expected basic built in go types.

gen is based/inspired by the work of Seth Shelnutt's [db2struct](https://github.com/Shelnutt2/db2struct), and Db2Struct is based/inspired by the work of ChimeraCoder's gojson package [gojson](https://github.com/ChimeraCoder/gojson).


## Usage

```BASH
go get github.com/smallnest/gen
gen --connstr "root@tcp(127.0.0.1:3306)/employees?&parseTime=True" --database employees  --json --gorm --guregu --rest
```

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
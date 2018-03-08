package dbmeta

// import (
// 	"database/sql"
// 	"errors"
// 	"fmt"
// 	"strings"
// )

// // GetColumnsFromMysqlTable Select column details from information schema and return map of map
// func GetColumnsFromMysqlTable(db *sql.DB, mysqlDatabase string, mysqlTable string) (*map[string]map[string]string, error) {
// 	var err error
// 	// Check for error in db, note this does not check connectivity but does check uri
// 	if err != nil {
// 		fmt.Println("Error opening mysql db: " + err.Error())
// 		return nil, err
// 	}

// 	// Store colum as map of maps
// 	columnDataTypes := make(map[string]map[string]string)
// 	// Select columnd data from INFORMATION_SCHEMA
// 	columnDataTypeQuery := "SELECT COLUMN_NAME, DATA_TYPE, IS_NULLABLE FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = ? AND table_name = ?"

// 	rows, err := db.Query(columnDataTypeQuery, mysqlDatabase, mysqlTable)

// 	if err != nil {
// 		fmt.Println("Error selecting from db: " + err.Error())
// 		return nil, err
// 	}
// 	if rows != nil {
// 		defer rows.Close()
// 	} else {
// 		return nil, errors.New("No results returned for table")
// 	}

// 	for rows.Next() {
// 		var column string
// 		var dataType string
// 		var nullable string
// 		rows.Scan(&column, &dataType, &nullable)

// 		columnDataTypes[column] = map[string]string{"value": dataType, "nullable": nullable}
// 	}

// 	return &columnDataTypes, err
// }

// // Generate go struct entries for a map[string]interface{} structure
// func generateMysqlTypes(db *sql.DB, obj map[string]map[string]string, depth int, jsonAnnotation bool, gormAnnotation bool, gureguTypes bool) []string {
// 	keys := make([]string, 0, len(obj))
// 	for key := range obj {
// 		keys = append(keys, key)
// 	}

// 	//sort.Strings(keys)

// 	var fields []string
// 	var field = ""
// 	for _, key := range keys {
// 		mysqlType := obj[key]
// 		nullable := false
// 		if mysqlType["nullable"] == "YES" {
// 			nullable = true
// 		}

// 		// Get the corresponding go value type for this mysql type
// 		var valueType string
// 		// If the guregu (https://github.com/guregu/null) CLI option is passed use its types, otherwise use go's sql.NullX

// 		valueType = mysqlTypeToGoType(mysqlType["value"], nullable, gureguTypes)

// 		fieldName := FmtFieldName(stringifyFirstChar(key))
// 		var annotations []string
// 		if gormAnnotation == true {
// 			annotations = append(annotations, fmt.Sprintf("gorm:\"column:%s\"", key))
// 		}
// 		if jsonAnnotation == true {
// 			annotations = append(annotations, fmt.Sprintf("json:\"%s\"", key))
// 		}
// 		if len(annotations) > 0 {
// 			field = fmt.Sprintf("%s %s `%s`",
// 				fieldName,
// 				valueType,
// 				strings.Join(annotations, " "))

// 		} else {
// 			field = fmt.Sprintf("%s %s",
// 				fieldName,
// 				valueType)
// 		}

// 		fields = append(fields, field)
// 	}
// 	return fields
// }

// // mysqlTypeToGoType converts the mysql types to go compatible sql.Nullable (https://golang.org/pkg/database/sql/) types
// func mysqlTypeToGoType(mysqlType string, nullable bool, gureguTypes bool) string {
// 	switch mysqlType {
// 	case "tinyint", "int", "smallint", "mediumint":
// 		if nullable {
// 			if gureguTypes {
// 				return gureguNullInt
// 			}
// 			return sqlNullInt
// 		}
// 		return golangInt
// 	case "bigint":
// 		if nullable {
// 			if gureguTypes {
// 				return gureguNullInt
// 			}
// 			return sqlNullInt
// 		}
// 		return golangInt64
// 	case "char", "enum", "varchar", "longtext", "mediumtext", "text", "tinytext":
// 		if nullable {
// 			if gureguTypes {
// 				return gureguNullString
// 			}
// 			return sqlNullString
// 		}
// 		return "string"
// 	case "date", "datetime", "time", "timestamp":
// 		if nullable && gureguTypes {
// 			return gureguNullTime
// 		}
// 		return golangTime
// 	case "decimal", "double":
// 		if nullable {
// 			if gureguTypes {
// 				return gureguNullFloat
// 			}
// 			return sqlNullFloat
// 		}
// 		return golangFloat64
// 	case "float":
// 		if nullable {
// 			if gureguTypes {
// 				return gureguNullFloat
// 			}
// 			return sqlNullFloat
// 		}
// 		return golangFloat32
// 	case "binary", "blob", "longblob", "mediumblob", "varbinary":
// 		return golangByteArray
// 	}
// 	return ""
// }

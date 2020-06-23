package dbmeta

import (
	"bytes"
	"fmt"
)

// PrimaryKeyCount return the number of primary keys in table
func PrimaryKeyCount(dbTable DbTableMeta) int {
	primaryKeys := 0
	for _, col := range dbTable.Columns() {
		if col.IsPrimaryKey() {
			primaryKeys++
		}
	}
	return primaryKeys
}

// PrimaryKeyNames return the list of primary key names
func PrimaryKeyNames(dbTable DbTableMeta) []string {
	primaryKeyNames := make([]string, 0)
	for _, col := range dbTable.Columns() {
		if col.IsPrimaryKey() {
			primaryKeyNames = append(primaryKeyNames, col.Name())
		}
	}
	return primaryKeyNames
}

// NonPrimaryKeyNames return the list of primary key names
func NonPrimaryKeyNames(dbTable DbTableMeta) []string {
	primaryKeyNames := make([]string, 0)
	for _, col := range dbTable.Columns() {
		if !col.IsPrimaryKey() {
			primaryKeyNames = append(primaryKeyNames, col.Name())
		}
	}
	return primaryKeyNames
}

// GenerateDeleteSQL generate sql for a delete
func GenerateDeleteSQL(dbTable DbTableMeta) (string, error) {
	primaryCnt := PrimaryKeyCount(dbTable)

	if primaryCnt == 0 {
		return "", fmt.Errorf("table %s does not have a primary key, cannot generate sql", dbTable.TableName())
	}

	buf := bytes.Buffer{}
	buf.WriteString(fmt.Sprintf("DELETE FROM `%s` where", dbTable.TableName()))

	addedKey := 1
	for _, col := range dbTable.Columns() {
		if col.IsPrimaryKey() {
			buf.WriteString(fmt.Sprintf(" %s = ?", col.Name()))
			addedKey++

			if addedKey < primaryCnt {
				buf.WriteString(" AND")
			}
		}
	}

	return buf.String(), nil
}

// GenerateUpdateSQL generate sql for a update
func GenerateUpdateSQL(dbTable DbTableMeta) (string, error) {
	primaryCnt := PrimaryKeyCount(dbTable)
	// nonPrimaryCnt := len(dbTable.Columns()) - primaryCnt

	if primaryCnt == 0 {
		return "", fmt.Errorf("table %s does not have a primary key, cannot generate sql", dbTable.TableName())
	}

	buf := bytes.Buffer{}
	buf.WriteString(fmt.Sprintf("UPDATE `%s` set", dbTable.TableName()))

	setCol := 1
	for _, col := range dbTable.Columns() {
		if !col.IsPrimaryKey() {
			if setCol != 1 {
				buf.WriteString(",")
			}

			buf.WriteString(fmt.Sprintf(" %s = ?", col.Name()))
			setCol++
		}
	}

	buf.WriteString(" WHERE")
	addedKey := 0
	for _, col := range dbTable.Columns() {
		if col.IsPrimaryKey() {
			buf.WriteString(fmt.Sprintf(" %s = ?", col.Name()))

			setCol++
			addedKey++

			if addedKey < primaryCnt {
				buf.WriteString(" AND")
			}
		}
	}

	return buf.String(), nil
}

// GenerateInsertSQL generate sql for a insert
func GenerateInsertSQL(dbTable DbTableMeta) (string, error) {
	primaryCnt := PrimaryKeyCount(dbTable)

	if primaryCnt == 0 {
		return "", fmt.Errorf("table %s does not have a primary key, cannot generate sql", dbTable.TableName())
	}

	buf := bytes.Buffer{}
	buf.WriteString(fmt.Sprintf("INSERT INTO `%s` (", dbTable.TableName()))

	pastFirst := false
	for _, col := range dbTable.Columns() {
		if !col.IsAutoIncrement() {
			if pastFirst {
				buf.WriteString(", ")
			}

			buf.WriteString(fmt.Sprintf(" %s", col.Name()))
			pastFirst = true
		}
	}
	buf.WriteString(") values ( ")

	pastFirst = false
	pos := 1
	for _, col := range dbTable.Columns() {
		if !col.IsAutoIncrement() {
			if pastFirst {
				buf.WriteString(", ")
			}

			buf.WriteString(fmt.Sprintf("?"))
			pos++
			pastFirst = true
		}
	}

	buf.WriteString(" )")
	return buf.String(), nil
}

// GenerateSelectOneSQL generate sql for selecting one record
func GenerateSelectOneSQL(dbTable DbTableMeta) (string, error) {
	primaryCnt := PrimaryKeyCount(dbTable)

	if primaryCnt == 0 {
		return "", fmt.Errorf("table %s does not have a primary key, cannot generate sql", dbTable.TableName())
	}

	buf := bytes.Buffer{}
	buf.WriteString(fmt.Sprintf("SELECT * FROM `%s` WHERE ", dbTable.TableName()))

	pastFirst := false
	pos := 1
	for _, col := range dbTable.Columns() {
		if col.IsPrimaryKey() {
			if pastFirst {
				buf.WriteString(" AND ")
			}

			buf.WriteString(fmt.Sprintf("%s = ?", col.Name()))
			pos++
			pastFirst = true
		}
	}
	return buf.String(), nil
}

// GenerateSelectMultiSQL generate sql for selecting multiple records
func GenerateSelectMultiSQL(dbTable DbTableMeta) (string, error) {
	primaryCnt := PrimaryKeyCount(dbTable)

	if primaryCnt == 0 {
		return "", fmt.Errorf("table %s does not have a primary key, cannot generate sql", dbTable.TableName())
	}

	buf := bytes.Buffer{}
	buf.WriteString(fmt.Sprintf("SELECT * FROM `%s`", dbTable.TableName()))
	return buf.String(), nil
}

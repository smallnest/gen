package dbmeta

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"unicode"
)

// commonInitialisms is a set of common initialisms.
// Only add entries that are highly unlikely to be non-initialisms.
// For instance, "ID" is fine (Freudian code is rare), but "AND" is not.
var commonInitialisms = map[string]bool{
	"API":   true,
	"ASCII": true,
	"CPU":   true,
	"CSS":   true,
	"DNS":   true,
	"EOF":   true,
	"GUID":  true,
	"HTML":  true,
	"HTTP":  true,
	"HTTPS": true,
	"ID":    true,
	"IP":    true,
	"JSON":  true,
	"LHS":   true,
	"QPS":   true,
	"RAM":   true,
	"RHS":   true,
	"RPC":   true,
	"SLA":   true,
	"SMTP":  true,
	"SSH":   true,
	"TLS":   true,
	"TTL":   true,
	"UI":    true,
	"UID":   true,
	"UUID":  true,
	"URI":   true,
	"URL":   true,
	"UTF8":  true,
	"VM":    true,
	"XML":   true,
	"ACL":   true,
}

var intToWordMap = []string{
	"zero",
	"one",
	"two",
	"three",
	"four",
	"five",
	"six",
	"seven",
	"eight",
	"nine",
}

var parsePrimaryKeys = map[string]string{
	"uint8":     "parseUint8",
	"uint16":    "parseUint16",
	"uint32":    "parseUint32",
	"uint64":    "parseUint64",
	"int":       "parseInt",
	"int8":      "parseInt8",
	"int16":     "parseInt16",
	"int32":     "parseInt32",
	"int64":     "parseInt64",
	"string":    "parseString",
	"uuid.UUID": "parseUUID",
	"time.Time": "parseTime",
	"varbinary": "parseBytes",
}

var reservedFieldNames = map[string]bool{
	"TableName":  true,
	"BeforeSave": true,
	"Prepare":    true,
	"Validate":   true,
	"type":       true,
}

// RenameReservedName renames a reserved word
func RenameReservedName(s string) string {
	_, match := reservedFieldNames[s]
	if match {
		return fmt.Sprintf("%s_", s)
	}

	return s
}

// FmtFieldName formats a string as a struct key
//
// Example:
// 	fmtFieldName("foo_id")
// Output: FooID
func FmtFieldName(s string) string {
	name := lintFieldName(s)
	runes := []rune(name)
	for i, c := range runes {
		ok := unicode.IsLetter(c) || unicode.IsDigit(c)
		if i == 0 {
			ok = unicode.IsLetter(c)
		}
		if !ok {
			runes[i] = '_'
		}
	}
	fieldName := string(runes)
	fieldName = RenameReservedName(fieldName)
	// fmt.Printf("FmtFieldName:%s=%s\n", s, fieldName)
	return fieldName
}

func isAllLower(name string) (allLower bool) {
	allLower = true
	for _, r := range name {
		if !unicode.IsLower(r) {
			allLower = false
			break
		}
	}
	return
}

func lintAllLowerFieldName(name string) string {
	runes := []rune(name)
	if u := strings.ToUpper(name); commonInitialisms[u] {
		copy(runes[0:], []rune(u))
	} else {
		runes[0] = unicode.ToUpper(runes[0])
	}
	return string(runes)
}

func lintFieldName(name string) string {
	// Fast path for simple cases: "_" and all lowercase.
	if name == "_" {
		return name
	}

	for len(name) > 0 && name[0] == '_' {
		name = name[1:]
	}

	allLower := isAllLower(name)

	if allLower {
		return lintAllLowerFieldName(name)
	}

	return lintMixedFieldName(name)
}

func lintMixedFieldName(name string) string {
	// Split camelCase at any lower->upper transition, and split on underscores.
	// Check each word for common initialisms.
	runes := []rune(name)
	w, i := 0, 0 // index of start of word, scan

	for i+1 <= len(runes) {
		eow := false // whether we hit the end of a word

		if i+1 == len(runes) {
			eow = true
		} else if runes[i+1] == '_' {
			// underscore; shift the remainder forward over any run of underscores
			eow = true
			n := 1
			for i+n+1 < len(runes) && runes[i+n+1] == '_' {
				n++
			}

			// Leave at most one underscore if the underscore is between two digits
			if i+n+1 < len(runes) && unicode.IsDigit(runes[i]) && unicode.IsDigit(runes[i+n+1]) {
				n--
			}

			copy(runes[i+1:], runes[i+n+1:])
			runes = runes[:len(runes)-n]
		} else if unicode.IsLower(runes[i]) && !unicode.IsLower(runes[i+1]) {
			// lower->non-lower
			eow = true
		}
		i++
		if !eow {
			continue
		}

		// [w,i) is a word.
		word := string(runes[w:i])
		if u := strings.ToUpper(word); commonInitialisms[u] {
			// All the common initialisms are ASCII,
			// so we can replace the bytes exactly.
			copy(runes[w:], []rune(u))
		} else if strings.ToLower(word) == word {
			// already all lowercase, and not the first word, so uppercase the first character.
			runes[w] = unicode.ToUpper(runes[w])
		}
		w = i
	}
	return string(runes)
}

// convert first character ints to strings
func stringifyFirstChar(str string) string {
	first := str[:1]

	i, err := strconv.ParseInt(first, 10, 8)
	if err != nil {
		return str
	}

	return intToWordMap[i] + "_" + str[1:]
}

// Copy a src struct into a destination struct
func Copy(dst interface{}, src interface{}) error {
	dstV := reflect.Indirect(reflect.ValueOf(dst))
	srcV := reflect.Indirect(reflect.ValueOf(src))

	if !dstV.CanAddr() {
		return errors.New("copy to value is unaddressable")
	}

	if srcV.Type() != dstV.Type() {
		return errors.New("different types can be copied")
	}

	for i := 0; i < dstV.NumField(); i++ {
		f := srcV.Field(i)
		if !isZeroOfUnderlyingType(f.Interface()) {
			dstV.Field(i).Set(f)
		}
	}

	return nil
}

func isZeroOfUnderlyingType(x interface{}) bool {
	return x == nil || reflect.DeepEqual(x, reflect.Zero(reflect.TypeOf(x)).Interface())
}

// Pwd template command to return the current working directory
func Pwd() string {
	currentWorkingDirectory, err := os.Getwd()
	if err != nil {
		return fmt.Sprintf("pwd returned an error %v", err)
	}
	return currentWorkingDirectory
}

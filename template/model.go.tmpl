package {{.modelPackageName}}

import (
    "database/sql"
    "time"

    "github.com/google/uuid"
    {{if .UseGuregu}} "github.com/guregu/null" {{end}}
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

var (
	_ = datatypes.JSON{}
)

/*
DB Table Details
-------------------------------------
{{ $ddl := .TableInfo.DBMeta.DDL }}
{{if $ddl }}
{{$ddl}}
{{- end}}

JSON Sample
-------------------------------------
{{ToJSON .TableInfo.Instance 4}}

{{if .TableInfo.Notes }}
Comments
-------------------------------------
{{ .TableInfo.Notes}}
{{end}}

*/

 {{if not .Config.AddProtobufAnnotation }}

// {{.StructName}} struct is a row record of the {{.TableName}} table in the {{.DatabaseName}} database
type {{.StructName}} struct {
    {{range .TableInfo.Fields}}{{.}}
    {{end}}
}
{{else}}

// {{.StructName}} struct is a row record of the {{.TableName}} table in the {{.DatabaseName}} database
/*
type {{.StructName}} struct {
    {{range .TableInfo.Fields}}{{.}}
    {{end}}
}
*/

{{end}}

var {{.TableName}}TableInfo = &TableInfo {
	Name: "{{.TableName}}",
	Columns: []*ColumnInfo{
        {{range .TableInfo.CodeFields}}
        {
        	Index: {{.ColumnMeta.Index}},
        	Name: "{{.ColumnMeta.Name}}",
        	Comment: `{{.ColumnMeta.Comment}}`,
        	Notes: `{{.ColumnMeta.Notes}}`,
        	Nullable: {{.ColumnMeta.Nullable}},
        	DatabaseTypeName: "{{.ColumnMeta.DatabaseTypeName}}",
        	DatabaseTypePretty: "{{.ColumnMeta.DatabaseTypePretty}}",
        	IsPrimaryKey: {{.ColumnMeta.IsPrimaryKey}},
        	IsAutoIncrement: {{.ColumnMeta.IsAutoIncrement}},
        	IsArray: {{.ColumnMeta.IsArray}},
        	ColumnType: "{{.ColumnMeta.ColumnType}}",
        	ColumnLength: {{.ColumnMeta.ColumnLength}},
        	GoFieldName: "{{.GoFieldName}}",
        	GoFieldType: "{{.GoFieldType}}",
        	JSONFieldName: "{{.JSONFieldName}}",
        	ProtobufFieldName: "{{.ProtobufFieldName}}",
        	ProtobufType: "{{.ProtobufType}}",
        	ProtobufPos: {{.ProtobufPos}},
        },
        {{end}}
	},
}



// TableName sets the insert table name for this struct type
func ({{.ShortStructName}} *{{.StructName}}) TableName() string {
	return "{{.TableName}}"
}

// BeforeSave invoked before saving, return an error if field is not populated.
func ({{.ShortStructName}} *{{.StructName}}) BeforeSave(tx *gorm.DB) error {
	return nil
}

// Prepare invoked before saving, can be used to populate fields etc.
func ({{.ShortStructName}} *{{.StructName}}) Prepare() {
}

// Validate invoked before performing action, return an error if field is not populated.
func ({{.ShortStructName}} *{{.StructName}}) Validate(action Action) error {
    return nil
}

// TableInfo return table meta data
func ({{.ShortStructName}} *{{.StructName}}) TableInfo() *TableInfo {
	return {{.TableName}}TableInfo
}

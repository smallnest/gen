package template

var ModelTmpl = `package {{.PackageName}}

import (
    "database/sql"
    "time"

    "github.com/guregu/null"
	"github.com/jinzhu/gorm"
)

var _ = time.Second
var _ = sql.LevelDefault
var _ = null.Bool{}

type {{.StructName}} struct {
    {{range .Fields}}{{.}}
    {{end}}
}

// TableName sets the insert table name for this struct type
func ({{.ShortStructName}} *{{.StructName}}) TableName() string {
	return "{{.TableName}}"
}

func ({{.ShortStructName}} *{{.StructName}}) Create{{.StructName}}(db *gorm.DB) error {
    return db.Create({{.ShortStructName}}).Error
}
`

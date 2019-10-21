package model

import (
	"database/sql"
	"time"

	"github.com/guregu/null"
)

var (
	_ = time.Second
	_ = sql.LevelDefault
	_ = null.Bool{}
)

type Example struct {
	ID int         `gorm:"column:id;primary_key" json:"id"`
	C1 null.String `gorm:"column:c1" json:"c1"`
	C2 null.String `gorm:"column:c2" json:"c2"`
	C3 null.Int    `gorm:"column:c3" json:"c3"`
	C4 null.Time   `gorm:"column:c4" json:"c4"`
	C5 null.String `gorm:"column:c5" json:"c5"`
}

// TableName sets the insert table name for this struct type
func (e *Example) TableName() string {
	return "example"
}

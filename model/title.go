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

type Title struct {
	EmpNo    int       `gorm:"column:emp_no;primary_key" json:"emp_no"`
	Title    string    `gorm:"column:title" json:"title"`
	FromDate time.Time `gorm:"column:from_date" json:"from_date"`
	ToDate   null.Time `gorm:"column:to_date" json:"to_date"`
}

// TableName sets the insert table name for this struct type
func (t *Title) TableName() string {
	return "titles"
}

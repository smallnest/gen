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

type Department struct {
	DeptNo   string `gorm:"column:dept_no;primary_key" json:"dept_no"`
	DeptName string `gorm:"column:dept_name" json:"dept_name"`
}

// TableName sets the insert table name for this struct type
func (d *Department) TableName() string {
	return "departments"
}

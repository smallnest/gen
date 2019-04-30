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

type Salary struct {
	EmpNo    int       `gorm:"column:emp_no;primary_key" json:"emp_no"`
	Salary   int       `gorm:"column:salary" json:"salary"`
	FromDate time.Time `gorm:"column:from_date" json:"from_date"`
	ToDate   time.Time `gorm:"column:to_date" json:"to_date"`
}

// TableName sets the insert table name for this struct type
func (s *Salary) TableName() string {
	return "salaries"
}

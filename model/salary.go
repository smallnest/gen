package model

import (
	"time"

	"github.com/guregu/null"
	"github.com/jinzhu/gorm"
)

type Salary struct {
	EmpNo    int       `gorm:"column:emp_no" json:"emp_no"`
	Salary   int       `gorm:"column:salary" json:"salary"`
	FromDate time.Time `gorm:"column:from_date" json:"from_date"`
	ToDate   time.Time `gorm:"column:to_date" json:"to_date"`
}

// TableName sets the insert table name for this struct type
func (s *Salary) TableName() string {
	return "salaries"
}

func (s *Salary) CreateSalary(db *gorm.DB) error {
	return db.Create(s).Error
}

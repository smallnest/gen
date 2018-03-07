package model

import (
	"time"

	"github.com/jinzhu/gorm"
)

type DeptEmp struct {
	EmpNo    int       `gorm:"column:emp_no" json:"emp_no"`
	DeptNo   string    `gorm:"column:dept_no" json:"dept_no"`
	FromDate time.Time `gorm:"column:from_date" json:"from_date"`
	ToDate   time.Time `gorm:"column:to_date" json:"to_date"`
}

// TableName sets the insert table name for this struct type
func (d *DeptEmp) TableName() string {
	return "dept_emp"
}

func (d *DeptEmp) CreateDeptEmp(db *gorm.DB) error {
	return db.Create(d).Error
}

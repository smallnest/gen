package model

import (
	"time"

	"github.com/guregu/null"
	"github.com/jinzhu/gorm"
)

type Department struct {
	DeptNo   string `gorm:"column:dept_no" json:"dept_no"`
	DeptName string `gorm:"column:dept_name" json:"dept_name"`
}

// TableName sets the insert table name for this struct type
func (d *Department) TableName() string {
	return "departments"
}

func (d *Department) CreateDepartment(db *gorm.DB) error {
	return db.Create(d).Error
}

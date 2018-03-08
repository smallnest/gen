package model

import (
	"time"

	"github.com/guregu/null"
	"github.com/jinzhu/gorm"
)

type Employee struct {
	EmpNo     int       `gorm:"column:emp_no" json:"emp_no"`
	BirthDate time.Time `gorm:"column:birth_date" json:"birth_date"`
	FirstName string    `gorm:"column:first_name" json:"first_name"`
	LastName  string    `gorm:"column:last_name" json:"last_name"`
	Gender    string    `gorm:"column:gender" json:"gender"`
	HireDate  time.Time `gorm:"column:hire_date" json:"hire_date"`
}

// TableName sets the insert table name for this struct type
func (e *Employee) TableName() string {
	return "employees"
}

func (e *Employee) CreateEmployee(db *gorm.DB) error {
	return db.Create(e).Error
}

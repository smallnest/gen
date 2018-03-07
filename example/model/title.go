package model

import (
	"time"

	"github.com/guregu/null"
	"github.com/jinzhu/gorm"
)

type Title struct {
	ToDate   null.Time `gorm:"column:to_date" json:"to_date"`
	EmpNo    int       `gorm:"column:emp_no" json:"emp_no"`
	Title    string    `gorm:"column:title" json:"title"`
	FromDate time.Time `gorm:"column:from_date" json:"from_date"`
}

// TableName sets the insert table name for this struct type
func (t *Title) TableName() string {
	return "titles"
}

func (t *Title) CreateTitle(db *gorm.DB) error {
	return db.Create(t).Error
}

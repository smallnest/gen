package models

import (
	"database/sql"
	"time"
)

var (
	_ = time.Second
)

type Rfq struct {
	ID              int        `gorm:"column:id;primary_key" json:"id"`
	TransactionID   int64      `gorm:"column:transaction_id" json:"transaction_id"`
	RfqNo           string     `gorm:"column:rfq_no" json:"rfq_no"`
	ReferenceNo     string     `gorm:"column:reference_no" json:"reference_no"`
	CompanyBuyerID  int64      `gorm:"column:company_buyer_id" json:"company_buyer_id"`
	CompanySellerID int64      `gorm:"column:company_seller_id" json:"company_seller_id"`
	Notes           string     `gorm:"column:notes" json:"notes"`
	Status          int        `gorm:"column:status" json:"status"`
	StatusReason    string     `gorm:"column:status_reason" json:"status_reason"`
	SubTotal        float64    `gorm:"column:sub_total" json:"sub_total"`
	TaxBasis        float64    `gorm:"column:tax_basis" json:"tax_basis"`
	Ppn             float64    `gorm:"column:ppn" json:"ppn"`
	Pph             float64    `gorm:"column:pph" json:"pph"`
	Rounding        float64    `gorm:"column:rounding" json:"rounding"`
	GrandTotal      float64    `gorm:"column:grand_total" json:"grand_total"`
	ExpiredAt       *time.Time `gorm:"column:expired_at" json:"expired_at"`
	RequestedBy     int64      `gorm:"column:requested_by" json:"requested_by"`
	CreatedBy       int64      `gorm:"column:created_by" json:"created_by"`
	CreatedAt       *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt       *time.Time `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt       *time.Time `gorm:"column:deleted_at" json:"deleted_at"`
}

// TableName sets the insert table name for this struct type
func (r *Rfq) TableName() string {
	return "rfqs"
}

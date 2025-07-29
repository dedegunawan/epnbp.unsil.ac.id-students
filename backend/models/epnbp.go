package models

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"time"
)

type PayUrl struct {
	ID            uint      `gorm:"primaryKey"`
	StudentBillID uint      `gorm:"column:student_bill_id" json:"student_bill_id"`
	PayUrl        string    `gorm:"column:pay_url" json:"pay_url"`
	InvoiceID     uint      `gorm:"column:invoice_id" json:"invoice_id"`
	Nominal       uint64    `gorm:"column:nominal" json:"nominal"`
	ExpiredAt     time.Time `gorm:"column:expired_at" json:"expired_at"`

	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

type PaymentCallback struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	StudentBillID *uint          `gorm:"column:student_bill_id" json:"student_bill_id"`
	Request       datatypes.JSON `gorm:"type:json" json:"request"`  // seluruh isi request (body, header, url, dsb)
	Response      datatypes.JSON `gorm:"type:json" json:"response"` // response kita ke provider
}

func MigrateEpnbp(db *gorm.DB) {
	db.AutoMigrate(&PayUrl{},
		&PaymentCallback{},
	)
}

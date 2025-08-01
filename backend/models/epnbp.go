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
type PaymentConfirmation struct {
	ID            uint      `gorm:"primaryKey"`
	StudentBillID uint      `gorm:"column:student_bill_id" json:"student_bill_id"`
	VaNumber      string    `gorm:"column:va_number" json:"va_number"`
	PaymentDate   string    `gorm:"column:payment_date" json:"payment_date"`
	ObjectName    string    `gorm:"column:object_name" json:"object_name"`
	Message       string    `gorm:"column:message;type:text" json:"message"`
	CreatedAt     time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at" json:"updated_at"`
}

type PaymentCallback struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	StudentBillID *uint          `gorm:"column:student_bill_id" json:"student_bill_id"`
	Status        string         `gorm:"column:status" json:"status"`
	TryCount      uint           `gorm:"column:try_count;default:0" json:"try_count"`
	Request       datatypes.JSON `gorm:"type:json" json:"request"`       // seluruh isi request (body, header, url, dsb)
	Response      datatypes.JSON `gorm:"type:json" json:"response"`      // response kita ke provider
	ResponseFrom  datatypes.JSON `gorm:"type:json" json:"response_from"` // response dari callback
	LastError     string         `gorm:"column:last_error;type:text" json:"last_error"`
	CreatedAt     time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"column:updated_at" json:"updated_at"`
	LastUpdatedAt time.Time      `gorm:"column:last_updated_at" json:"last_updated_at"`
}

func MigrateEpnbp(db *gorm.DB) {
	db.AutoMigrate(&PayUrl{},
		&PaymentCallback{},
		&PaymentConfirmation{},
		&PaymentCallback{},
	)
}

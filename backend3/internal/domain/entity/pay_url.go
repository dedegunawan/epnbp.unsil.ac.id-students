package entity

import "time"

// PayUrl represents a payment URL generated for a student bill
type PayUrl struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	StudentBillID uint      `gorm:"column:student_bill_id;index" json:"student_bill_id"`
	PayUrl        string    `gorm:"column:pay_url;type:text" json:"pay_url"`
	InvoiceID     uint      `gorm:"column:invoice_id" json:"invoice_id"`
	Nominal       uint64    `gorm:"column:nominal" json:"nominal"`
	ExpiredAt     time.Time `gorm:"column:expired_at" json:"expired_at"`
	CreatedAt     time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (PayUrl) TableName() string {
	return "pay_urls"
}





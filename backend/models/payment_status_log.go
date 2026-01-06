package models

import (
	"time"
)

// PaymentStatusLog menyimpan log perubahan status pembayaran
type PaymentStatusLog struct {
	ID                uint      `gorm:"primaryKey"`
	StudentBillID     uint      `gorm:"index;column:student_bill_id"`
	StudentID         string    `gorm:"column:student_id;size:20;index"`
	OldStatus         string    `gorm:"column:old_status;size:50"`         // "unpaid", "partial"
	NewStatus         string    `gorm:"column:new_status;size:50"`         // "paid"
	OldPaidAmount     int64     `gorm:"column:old_paid_amount"`
	NewPaidAmount     int64     `gorm:"column:new_paid_amount"`
	Amount            int64     `gorm:"column:amount"`                     // Amount yang dibayar
	PaymentDate       *time.Time `gorm:"column:payment_date"`              // Tanggal pembayaran dari API
	InvoiceID         *uint      `gorm:"column:invoice_id"`                 // Invoice ID dari PNBP
	VirtualAccount    string    `gorm:"column:virtual_account;size:50"`  // Virtual Account
	Identifier        string    `gorm:"column:identifier;size:50;index"`  // NPM/Student ID yang digunakan untuk search
	TimeDifference    int64     `gorm:"column:time_difference"`            // Selisih waktu dalam detik
	Source            string    `gorm:"column:source;size:50;default:'identifier_worker'"` // Sumber perubahan
	Message           string    `gorm:"column:message;type:text"`          // Pesan/log detail
	CreatedAt         time.Time `gorm:"column:created_at"`
	UpdatedAt         time.Time `gorm:"column:updated_at"`
}

func (PaymentStatusLog) TableName() string {
	return "payment_status_logs"
}






package entity

import "time"

// PaymentConfirmation represents a manual payment confirmation uploaded by student
type PaymentConfirmation struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	StudentBillID uint      `gorm:"column:student_bill_id;index" json:"student_bill_id"`
	VaNumber      string    `gorm:"column:va_number;size:50" json:"va_number"`
	PaymentDate   string    `gorm:"column:payment_date;size:50" json:"payment_date"`
	ObjectName    string    `gorm:"column:object_name;size:255" json:"object_name"` // MinIO object name
	Message       string    `gorm:"column:message;type:text" json:"message"`
	CreatedAt     time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (PaymentConfirmation) TableName() string {
	return "payment_confirmations"
}



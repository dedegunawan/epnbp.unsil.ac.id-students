package entity

import (
	"gorm.io/datatypes"
	"time"
)

// PaymentCallback represents a payment callback from payment gateway
type PaymentCallback struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	StudentBillID *uint          `gorm:"column:student_bill_id;index" json:"student_bill_id"`
	Status        string         `gorm:"column:status;size:20" json:"status"`
	TryCount      uint           `gorm:"column:try_count;default:0" json:"try_count"`
	Request       datatypes.JSON `gorm:"type:json" json:"request"`       // entire request (body, header, url, etc)
	Response      datatypes.JSON `gorm:"type:json" json:"response"`      // our response to provider
	ResponseFrom  datatypes.JSON `gorm:"type:json" json:"response_from"` // response from callback
	LastError     string         `gorm:"column:last_error;type:text" json:"last_error"`
	CreatedAt     time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"column:updated_at" json:"updated_at"`
	LastUpdatedAt time.Time      `gorm:"column:last_updated_at" json:"last_updated_at"`
}

func (PaymentCallback) TableName() string {
	return "payment_callbacks"
}





package entity

import (
	"time"
)

type BudgetPeriod struct {
	ID         uint64    `gorm:"primaryKey" json:"id"`
	Kode       string    `gorm:"size:10;uniqueIndex" json:"kode"`
	Year       int       `json:"year"`
	Semester   int       `json:"semester"` // 1 = Ganjil, 2 = Genap
	Name       string    `json:"name"`
	StartDate  time.Time `json:"start_date"`
	EndDate    time.Time `json:"end_date"`
	FiscalYear int       `json:"fiscal_year"`

	PaymentStartDate time.Time `gorm:"not null" json:"payment_start_date"` // Presisi detik
	PaymentEndDate   time.Time `gorm:"not null" json:"payment_end_date"`

	IsActive bool `gorm:"default:false;index"` // Menandakan tahun aktif

	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	//DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

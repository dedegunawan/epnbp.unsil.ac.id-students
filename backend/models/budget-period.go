package models

import "time"

type BudgetPeriod struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	Kode             string    `gorm:"size:100" json:"kode"`
	Year             int       `json:"year"`
	Semester         int       `json:"semester"`
	Name             string    `gorm:"size:255" json:"name"`
	StartDate        time.Time `json:"start_date"`
	EndDate          time.Time `json:"end_date"`
	FiscalYear       string    `gorm:"size:20" json:"fiscal_year"`
	PaymentStartDate time.Time `json:"payment_start_date"`
	PaymentEndDate   time.Time `json:"payment_end_date"`
	IsActive         bool      `json:"is_active"`
	IsStrict         bool      `json:"is_strict"`
	Description      string    `gorm:"type:text" json:"description"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type BudgetPeriodPaymentOverride struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	BudgetPeriodID   uint      `json:"budget_period_id"`
	ScopeType        string    `gorm:"size:20" json:"scope_type"` // e.g., "fakultas", "prodi", "individual"
	ScopeID          string    `gorm:"size:100" json:"scope_id"`
	PaymentStartDate time.Time `json:"payment_start_date"`
	PaymentEndDate   time.Time `json:"payment_end_date"`
	IsActive         bool      `json:"is_active"`
	Description      string    `gorm:"type:text" json:"description"`
	CreatedBy        string    `gorm:"size:100" json:"created_by"`
	UpdatedBy        string    `gorm:"size:100" json:"updated_by"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

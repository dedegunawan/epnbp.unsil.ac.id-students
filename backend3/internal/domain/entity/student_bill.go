package entity

import "time"

type StudentBill struct {
	ID                 uint   `gorm:"primaryKey" json:"id"`
	StudentID          string `gorm:"index;size:20" json:"student_id"`
	AcademicYear       string `gorm:"size:10" json:"academic_year"` // e.g., 20241
	BillTemplateItemID uint   `gorm:"index" json:"bill_template_item_id"`
	Name               string `gorm:"size:100" json:"name"`
	Quantity           int    `gorm:"default:1" json:"quantity"`
	Amount             int64  `gorm:"default:0" json:"amount"` // nominal tagihan awal (tanpa diskon)
	Beasiswa           int64  `gorm:"default:0" json:"beasiswa"`
	PaidAmount         int64  `gorm:"default:0" json:"paid_amount"`
	Draft              bool   `gorm:"default:true" json:"draft"`
	Note               string `gorm:"size:255" json:"note"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`

	Mahasiswa *Mahasiswa `gorm:"foreignKey:MhswID" json:"mahasiswa,omitempty"`
}

// Remaining calculates the remaining amount to be paid
func (sb *StudentBill) Remaining() int64 {
	remain := sb.Amount - sb.PaidAmount
	if remain < 0 {
		return 0
	}
	return remain
}

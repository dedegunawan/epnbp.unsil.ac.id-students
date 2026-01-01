package entity

import "time"

type StudentBill struct {
	ID                 uint   `gorm:"primaryKey"`
	StudentID          string `gorm:"index;size:20"`
	AcademicYear       string `gorm:"size:10"` // e.g., 20241
	BillTemplateItemID uint   `gorm:"index"`
	Name               string `gorm:"size:100"`
	Quantity           int    `gorm:"default:1"`
	Amount             int64  `gorm:"default:0"` // nominal tagihan awal (tanpa diskon)
	Beasiswa           int64  `gorm:"default:0"`
	PaidAmount         int64  `gorm:"default:0"`
	Draft              bool   `gorm:"default:true"`
	Note               string `gorm:"size:255"`
	CreatedAt          time.Time
	UpdatedAt          time.Time

	Mahasiswa *Mahasiswa `gorm:"foreignKey:MhswID"`
}

package models

import "time"

// StatusAkademik merepresentasikan tabel status_akademiks di database PNBP
type StatusAkademik struct {
	ID        uint   `gorm:"primaryKey;column:id"`
	Kode      string `gorm:"column:kode;size:10"`      // Kode status, misalnya "A" untuk Aktif, "N" untuk Non-Aktif
	Nama      string `gorm:"column:nama;size:191"`     // Nama status
	Deskripsi string `gorm:"column:deskripsi;type:text"` // Deskripsi status (opsional)
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

// TableName memastikan nama tabel sama seperti di Laravel
func (StatusAkademik) TableName() string {
	return "status_akademiks"
}


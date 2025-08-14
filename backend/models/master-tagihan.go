package models

import (
	"time"

	"gorm.io/gorm"
)

// MasterTagihan merepresentasikan tabel `master_tagihan`
type MasterTagihan struct {
	ID uint `gorm:"primaryKey"`

	Angkatan  int    `gorm:"column:angkatan;index"` // contoh: 2024
	ProdiID   uint   `gorm:"column:prodi_id;index"`
	ProgramID uint   `gorm:"column:program_id;index"`
	BipotID   uint   `gorm:"column:bipotid"`
	Nama      string `gorm:"column:nama;size:191;not null"`

	// Relations
	Prodi *ProdiPnbp `gorm:"foreignKey:ProdiID;references:ID"`

	// Timestamps & Soft delete (opsional; hapus DeletedAt jika tidak pakai soft delete)
	CreatedAt time.Time      `gorm:"column:created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

// TableName memastikan nama tabel sama seperti di Laravel
func (MasterTagihan) TableName() string {
	return "master_tagihan"
}

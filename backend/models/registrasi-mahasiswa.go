package models

import (
	"time"
)

// RegistrasiMahasiswa merepresentasikan tabel registrasi_mahasiswa di database PNBP
type RegistrasiMahasiswa struct {
	ID      uint   `gorm:"primaryKey"`
	NPM     string `gorm:"column:npm;size:50;index"`      // NPM/NIM mahasiswa
	TahunID string `gorm:"column:tahun_id;size:10;index"` // Tahun akademik (e.g., "20251")

	// Kolom-kolom tagihan (sesuaikan dengan struktur tabel yang sebenarnya)
	Nominal     int64  `gorm:"column:nominal"`               // Nominal tagihan
	NamaTagihan string `gorm:"column:nama_tagihan;size:255"` // Nama item tagihan
	KelUKT      string `gorm:"column:kel_ukt;size:10"`       // Kelompok UKT (jika ada)

	// Kolom tambahan yang mungkin ada
	Semester int    `gorm:"column:semester"`          // Semester (jika ada)
	Status   string `gorm:"column:status;size:50"`    // Status registrasi (jika ada)
	Catatan  string `gorm:"column:catatan;type:text"` // Catatan (jika ada)

	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (RegistrasiMahasiswa) TableName() string {
	return "registrasi_mahasiswa"
}

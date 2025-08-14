package models

import (
	"gorm.io/gorm"
	"time"
)

type MahasiswaMaster struct {
	ID uint `gorm:"primaryKey"`

	// Identitas
	StudentID    string     `gorm:"column:student_id;size:50;index"` // NPM/StudentID
	NamaLengkap  string     `gorm:"column:nama_lengkap;size:191"`
	NIK          string     `gorm:"column:nik;size:32"`
	TempatLahir  string     `gorm:"column:tempat_lahir;size:100"`
	TanggalLahir *time.Time `gorm:"column:tanggal_lahir"`         // DATE/DATE-TIME
	JenisKelamin string     `gorm:"column:jenis_kelamin;size:10"` // 'L'/'P' atau teks

	// Akademik
	ProdiID             uint `gorm:"column:prodi_id"`
	ProgramID           uint `gorm:"column:program_id"`
	JenjangPendidikanID uint `gorm:"column:jenjang_pendidikan_id"`
	TahunMasuk          int  `gorm:"column:tahun_masuk"` // contoh: 2023
	SemesterMasukID     uint `gorm:"column:semester_masuk_id"`
	StatusAkademikID    uint `gorm:"column:status_akademik_id"`

	// Kontak
	Email string `gorm:"column:email;size:191"`
	NoHP  string `gorm:"column:no_hp;size:32"`

	// Alamat
	Alamat    string `gorm:"column:alamat;type:text"`
	VillageID string `gorm:"column:village_id;size:20"` // kode desa (umumnya string)

	// Referensi eksternal
	ExternalRef     string         `gorm:"column:external_ref;size:191"`
	UKT             int64          `gorm:"column:ukt"` // asumsi: nominal dalam rupiah
	MasterTagihanID uint           `gorm:"column:master_tagihan_id"`
	MasterTagihan   *MasterTagihan `gorm:"foreignKey:MasterTagihanID;references:ID"`

	// Timestamps & Soft Delete (meniru Eloquent)
	CreatedAt time.Time      `gorm:"column:created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

// TableName mengatur nama tabel agar sesuai konvensi Laravel (plural snake_case)
func (MahasiswaMaster) TableName() string {
	return "mahasiswa_masters"
}

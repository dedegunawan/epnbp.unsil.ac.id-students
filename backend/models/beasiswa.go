package models

import "time"

type Beasiswa struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement"`
	NoSK         string    `gorm:"column:no_sk;size:191;unique;not null"`
	Deskripsi    string    `gorm:"type:text;not null"`
	TanggalSK    time.Time `gorm:"column:tanggal_sk;type:date;not null"`
	FileBeasiswa *string   `gorm:"column:file_beasiswa;size:191"` // nullable
	Status       string    `gorm:"type:enum('draft','active','inactive');default:'draft';not null"`
	CreatedAt    *time.Time
	UpdatedAt    *time.Time

	// Relasi
	Details []DetailBeasiswa `gorm:"foreignKey:BeasiswaID"`
}

// TableName sets the table name for Beasiswa
func (Beasiswa) TableName() string {
	return "beasiswa"
}

type DetailBeasiswa struct {
	ID                 uint64  `gorm:"primaryKey;autoIncrement"`
	BeasiswaID         uint64  `gorm:"column:beasiswa_id;not null"`
	NPM                string  `gorm:"size:191;not null"`
	TahunID            string  `gorm:"column:tahun_id;size:191;not null"`
	KelompokUKTSaatIni string  `gorm:"column:kel_ukt_saat_ini;size:191;not null"`
	NominalUKTSaatIni  float64 `gorm:"column:nominal_ukt_saat_ini;type:decimal(15,2);not null"`
	JenisBeasiswa      string  `gorm:"size:191;not null"`
	NominalBeasiswa    float64 `gorm:"type:decimal(15,2);not null"`
	NominalYangDibayar float64 `gorm:"column:nominal_yang_dibayar;type:decimal(15,2);not null;default:0.00"`
	CreatedAt          *time.Time
	UpdatedAt          *time.Time

	// Relasi
	Beasiswa Beasiswa `gorm:"foreignKey:BeasiswaID;constraint:OnDelete:CASCADE;"`

	// Unik
	// GORM tidak bisa enforce UNIQUE (npm, tahun_id) langsung, perlu di migrasi manual
}

// TableName sets the table name for DetailBeasiswa
func (DetailBeasiswa) TableName() string {
	return "detail_beasiswa"
}

package models

import (
	"time"

	"gorm.io/gorm"
)

type ProdiPnbp struct {
	ID         uint           `gorm:"primaryKey"`
	KodeProdi  string         `gorm:"column:kode_prodi;size:50;not null"`
	NamaProdi  string         `gorm:"column:nama_prodi;size:191;not null"`
	FakultasID uint           `gorm:"column:fakultas_id;index"`
	Fakultas   *Fakultas      `gorm:"foreignKey:FakultasID;references:ID"` // Relasi belongsTo ke Fakultas
	CreatedAt  time.Time      `gorm:"column:created_at"`
	UpdatedAt  time.Time      `gorm:"column:updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (ProdiPnbp) TableName() string {
	return "prodi"
}

type FakultasPnbp struct {
	ID           uint           `gorm:"primaryKey"`
	KodeFakultas string         `gorm:"column:kode_fakultas;size:50;not null"`
	NamaFakultas string         `gorm:"column:nama_fakultas;size:191;not null"`
	CreatedAt    time.Time      `gorm:"column:created_at"`
	UpdatedAt    time.Time      `gorm:"column:updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

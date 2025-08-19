package models

import (
	"time"
)

// Tabel: mahasiswas
type MahasiswaPnbp struct {
	ID uint `gorm:"primaryKey"`
	// Primary key pakai MhswID (NPM/NIM). Jika di DB Anda masih punya kolom `id`,
	// hapus tag primaryKey di sini dan tambahkan field ID uint `gorm:"primaryKey"`
	MhswID string `gorm:"column:MhswID;primaryKey;size:32" json:"MhswID"`
	Nama   string `gorm:"column:Nama;size:255;not null" json:"Nama"`

	// Kedua kolom berikut dipetakan apa adanya sesuai DB
	ProdiID *uint `gorm:"column:ProdiID" json:"ProdiID,omitempty"`
	ProdiId *uint `gorm:"column:prodi_id" json:"prodi_id,omitempty"`

	CustomerID *uint   `gorm:"column:customer_id" json:"customer_id,omitempty"`
	Email      *string `gorm:"column:email;size:255" json:"email,omitempty"`
	FullData   *string `gorm:"column:full_data" json:"full_data,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (MahasiswaPnbp) TableName() string {
	return "mahasiswas"
}

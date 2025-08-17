package models

import (
	"time"
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
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

// TableName memastikan nama tabel sama seperti di Laravel
func (MasterTagihan) TableName() string {
	return "master_tagihan"
}

type DetailTagihan struct {
	ID              uint64  `json:"id" gorm:"primaryKey;autoIncrement"`
	MasterTagihanID uint64  `json:"master_tagihan_id" gorm:"not null;index"` // FK -> master_tagihan.id
	BipotnamaID     *uint64 `json:"bipotnama_id,omitempty" gorm:"index"`     // FK -> bipotnama.id (opsional)
	Bipot2ID        *uint64 `json:"bipot2id,omitempty" gorm:"column:bipot2id"`
	Nama            string  `json:"nama" gorm:"size:191;not null"`
	KelUKT          *string `json:"kel_ukt,omitempty" gorm:"column:kel_ukt;size:32"`
	BerapaKali      *int    `json:"berapa_kali,omitempty" gorm:"column:berapa_kali"`
	MulaiSesiBerapa *int    `json:"mulai_sesi_berapa,omitempty" gorm:"column:mulai_sesi_berapa"`
	Nominal         int64   `json:"nominal" gorm:"not null"` // uang: gunakan integer (mis. rupiah)

	// Relations
	MasterTagihan *MasterTagihan `json:"master_tagihan,omitempty" gorm:"foreignKey:MasterTagihanID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (DetailTagihan) TableName() string { return "detail_tagihan" }

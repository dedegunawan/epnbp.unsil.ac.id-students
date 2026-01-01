package entity

import "time"

type Prodi struct {
	ID         uint64 `gorm:"primaryKey;autoIncrement;not null" json:"id"`
	KodeProdi  string `gorm:"type:varchar(100);" json:"kode_prodi"`
	NamaProdi  string `gorm:"type:varchar(100);" json:"nama_prodi"`
	FakultasID uint64 `json:"fakultas_id"`

	Fakultas *Fakultas `gorm:"foreignKey:FakultasID" json:"fakultas"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

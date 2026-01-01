package entity

import "time"

type Fakultas struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement;not null" json:"id"`
	KodeFakultas string    `gorm:"type:varchar(100);" json:"kode_fakultas"`
	NamaFakultas string    `gorm:"type:varchar(100);" json:"nama_fakultas"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	Prodi []Prodi `gorm:"foreignKey:ProdiID" json:"prodi"`
}

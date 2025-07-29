package models

import "encoding/json"

// Fakultas model
type Fakultas struct {
	ID           uint    `gorm:"primaryKey;autoIncrement" json:"id"`
	KodeFakultas string  `gorm:"type:varchar(50);not null" json:"kode_fakultas"`
	NamaFakultas string  `gorm:"type:varchar(100);not null" json:"nama_fakultas"`
	Prodis       []Prodi `gorm:"foreignKey:FakultasID" json:"prodis"`
}

// Prodi model
type Prodi struct {
	ID         uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	KodeProdi  string `gorm:"type:varchar(50);not null" json:"kode_prodi"`
	NamaProdi  string `gorm:"type:varchar(100);not null" json:"nama_prodi"`
	FakultasID uint   `gorm:"not null" json:"fakultas_id"`

	Fakultas   Fakultas    `gorm:"foreignKey:FakultasID" json:"fakultas"`
	Mahasiswas []Mahasiswa `gorm:"foreignKey:ProdiID" json:"mahasiswas"`
}

// Mahasiswa model
type Mahasiswa struct {
	ID       uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	MhswID   string `gorm:"type:varchar(50);not null" json:"mhsw_id"`
	Nama     string `gorm:"type:varchar(100);not null" json:"nama"`
	ProdiID  uint   `gorm:"not null" json:"prodi_id"`
	UKT      string `gorm:"column:kel_ukt" json:"kel_ukt"`
	BIPOTID  string `gorm:"column:bipot_id"  json:"bipot_id"`
	Email    string `gorm:"type:varchar(100)" json:"email"`
	FullData string `gorm:"default:false" json:"full_data"`

	Prodi Prodi `gorm:"foreignKey:ProdiID" json:"prodi"`
}

func (m *Mahasiswa) ParseFullData() map[string]interface{} {
	var result map[string]interface{}

	if m.FullData == "" {
		return map[string]interface{}{}
	}

	if err := json.Unmarshal([]byte(m.FullData), &result); err != nil {
		return map[string]interface{}{}
	}

	return result
}

package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// RegistrasiMahasiswa merepresentasikan tabel registrasi_mahasiswa di database PNBP
type RegistrasiMahasiswa struct {
	ID                 uint       `gorm:"primaryKey;column:id"`
	NPM                string     `gorm:"column:npm;size:191;not null;index:registrasi_mahasiswa_npm_index"`
	TahunID            string     `gorm:"column:tahun_id;size:191;not null;index:registrasi_mahasiswa_tahun_id_index"`
	KelUKT             *string    `gorm:"column:kel_ukt;size:191"`
	IDUKT              *string    `gorm:"column:id_ukt;size:191"`
	NominalUKT         *float64   `gorm:"column:nominal_ukt;type:decimal(15,2)"` // decimal(15,2)
	SudahBayar         bool       `gorm:"column:sudah_bayar;default:0;index:registrasi_mahasiswa_sudah_bayar_index"`
	NominalBayar       *float64   `gorm:"column:nominal_bayar;type:decimal(15,2)"` // decimal(15,2)
	StatusStudentEPNBP *string    `gorm:"column:status_student_epnbp;size:191"`
	CallbackSintesys   *JSONB     `gorm:"column:callback_sintesys;type:longtext"` // JSON field
	CreatedAt          *time.Time `gorm:"column:created_at"`
	UpdatedAt          *time.Time `gorm:"column:updated_at"`
}

// JSONB adalah custom type untuk handle JSON field di MySQL
type JSONB map[string]interface{}

// Value implements driver.Valuer interface
func (j JSONB) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements sql.Scanner interface
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return nil
	}

	if len(bytes) == 0 {
		*j = nil
		return nil
	}

	result := make(JSONB)
	if err := json.Unmarshal(bytes, &result); err != nil {
		return err
	}
	*j = result
	return nil
}

func (RegistrasiMahasiswa) TableName() string {
	return "registrasi_mahasiswa"
}

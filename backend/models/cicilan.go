package models

import (
	"time"
)

type Cicilan struct {
	ID            uint   `gorm:"primaryKey"`
	NPM           string `gorm:"column:npm;size:20;not null"`
	TahunID       string `gorm:"column:tahun_id;size:10;not null"`
	KelUkt        string `gorm:"column:kel_ukt;size:10"`
	NominalUkt    int64  `gorm:"column:nominal_ukt"`
	JumlahCicilan int    `gorm:"column:jumlah_cicilan"`
	Catatan       string `gorm:"column:catatan;type:text"`
	File          string `gorm:"column:file;type:varchar(255)"`
	CreatedBy     uint   `gorm:"column:created_by"`
	UpdatedBy     uint   `gorm:"column:updated_by"`
	CreatedAt     time.Time
	UpdatedAt     time.Time

	Mahasiswa     *Mahasiswa      `gorm:"foreignKey:NPM;references:MhswID"`
	Creator       *User           `gorm:"foreignKey:CreatedBy"`
	Updater       *User           `gorm:"foreignKey:UpdatedBy"`
	DetailCicilan []DetailCicilan `gorm:"foreignKey:CicilanID"`
}

func (Cicilan) TableName() string {
	return "cicilans"
}

type DetailCicilan struct {
	ID         uint      `gorm:"primaryKey"`
	CicilanID  uint      `gorm:"column:cicilan_id"`
	SequenceNo int       `gorm:"column:sequence_no"`
	DueDate    time.Time `gorm:"column:due_date"`
	Amount     int64     `gorm:"column:amount"`
	Status     string    `gorm:"column:status;size:50"`
	Catatan    string    `gorm:"column:catatan;type:text"`

	Cicilan *Cicilan `gorm:"foreignKey:CicilanID"`
}

func (DetailCicilan) TableName() string {
	return "detail_cicilans"
}

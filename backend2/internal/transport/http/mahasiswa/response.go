package mahasiswa

import (
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/internal/domain/entity"
	"strconv"
)

type ProdiResponse struct {
	ID   uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Nama string `gorm:"type:varchar(100);not null" json:"nama"`
	Kode string `gorm:"type:varchar(10);not null" json:"kode"`
}

type MahasiswaResponse struct {
	ID       uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	MhswID   string `gorm:"type:varchar(50);not null" json:"mhsw_id"`
	Nama     string `gorm:"type:varchar(100);not null" json:"nama"`
	ProdiID  uint   `gorm:"not null" json:"prodi_id"`
	UKT      string `gorm:"column:kel_ukt" json:"kel_ukt"`
	BIPOTID  string `gorm:"column:bipot_id"  json:"bipot_id"`
	Email    string `gorm:"type:varchar(100)" json:"email"`
	FullData string `gorm:"default:false" json:"full_data"`

	Prodi ProdiResponse `gorm:"foreignKey:ProdiID" json:"prodi"`
}

type StudentBillResponse struct {
	Tahun               *entity.BudgetPeriod `json:"tahun"`
	IsPaid              bool                 `json:"isPaid"`
	IsGenerated         bool                 `json:"isGenerated"`
	TagihanHarusDibayar []entity.StudentBill `json:"tagihanHarusDibayar"`
	HistoryTagihan      []entity.StudentBill `json:"historyTagihan"`
}

func ConvertResponseFromMahasiswa(mahasiswa *entity.Mahasiswa) *MahasiswaResponse {
	UKTString := strconv.Itoa(int(mahasiswa.UKT))
	return &MahasiswaResponse{
		ID:       uint(mahasiswa.ID),
		MhswID:   mahasiswa.StudentID,
		Nama:     mahasiswa.NamaLengkap,
		ProdiID:  uint(mahasiswa.ProdiID),
		UKT:      UKTString,
		BIPOTID:  "",
		Email:    mahasiswa.Email,
		FullData: "",
		Prodi:    *ConvertResponseFromProdi(mahasiswa.Prodi),
	}
}

func ConvertResponseFromProdi(prodi *entity.Prodi) *ProdiResponse {
	if prodi == nil {
		return &ProdiResponse{}
	}
	return &ProdiResponse{
		ID:   uint(prodi.ID),
		Nama: prodi.NamaProdi,
		Kode: prodi.KodeProdi,
	}
}

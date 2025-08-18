package repositories

import (
	"errors"
	"github.com/dedegunawan/backend-ujian-telp-v5/models"
	"gorm.io/gorm"
	"strconv"
)

type MasterTagihanRepository struct {
	DB *gorm.DB
}

func NewMasterTagihanRepository(db *gorm.DB) *MasterTagihanRepository {
	return &MasterTagihanRepository{
		DB: db,
	}
}

func (mtr *MasterTagihanRepository) GetNominalTagihanMahasiswa(mahasiswa models.Mahasiswa) int64 {
	detailTagihan, err := mtr.FindMasterTagihanMahasiswa(mahasiswa)
	if err == nil && detailTagihan != nil {
		return detailTagihan.Nominal
	}

	return 0
}

func (mtr *MasterTagihanRepository) FindMasterTagihanMahasiswa(mahasiswa models.Mahasiswa) (*models.DetailTagihan, error) {

	//tahunIDInt :=
	tahunIDString := mahasiswa.ParseFullData()["TahunID"].(string)
	tahun := tahunIDString[:4] // Ambil 4 karakter pertama dari TahunID

	//prodiIDInt := mahasiswa.ParseFullData()["ProdiID"].(string)
	prodiIDString := mahasiswa.ParseFullData()["ProdiID"].(string)

	programID := mahasiswa.ParseFullData()["ProgramID"].(string)

	if tahunIDString == "" || prodiIDString == "" || programID == "" {
		return nil, errors.New("invalid mahasiswa data: TahunID, ProdiID, or ProgramID is missing")

	}

	var tagihan models.MasterTagihan
	err := mtr.DB.Where("Angkatan = ? and ProdiID = ? and ProgramID = ?", tahun, prodiIDString, programID).
		First(&tagihan).Error

	if err != nil {
		return nil, errors.New("invalid master tagihan data: " + err.Error())
	}

	UKTInt := mahasiswa.ParseFullData()["UKT"].(int64)
	UKTString := strconv.Itoa(int(UKTInt))

	var detailTagihan models.DetailTagihan
	err = mtr.DB.Where("MasterTagihanID = ? and kel_ukt = ?", tagihan.ID, UKTString).
		First(&detailTagihan).Error

	return &detailTagihan, err
}

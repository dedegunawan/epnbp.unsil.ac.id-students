package repositories

import (
	"errors"
	"github.com/dedegunawan/backend-ujian-telp-v5/models"
	"github.com/dedegunawan/backend-ujian-telp-v5/utils"
	"gorm.io/gorm"
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

	if err != nil {
		utils.Log.Info("GetNominalTagihanMahasiswa: error finding master tagihan for mahasiswa: ", err)
	}

	return 0
}

func (mtr *MasterTagihanRepository) FindMasterTagihanMahasiswa(mahasiswa models.Mahasiswa) (*models.DetailTagihan, error) {

	//tahunIDInt :=
	tahunIDString := utils.GetStringFromAny(mahasiswa.ParseFullData()["TahunID"])
	tahun := tahunIDString[:4] // Ambil 4 karakter pertama dari TahunID

	prodiIDString := utils.GetStringFromAny(mahasiswa.ParseFullData()["ProdiID"])

	programID := utils.GetStringFromAny(mahasiswa.ParseFullData()["ProgramID"])

	if programID == "" && (prodiIDString[:1] == "8" || mahasiswa.MhswID[:1] == "9") {
		programID = "2 - Non Reguler"
	} else if programID == "" && prodiIDString[:1] != "8" && prodiIDString[:1] != "9" {
		programID = "1 - Reg" // Default to program 1 if not specified
	}

	utils.Log.Info("Search by : ", map[string]interface{}{
		"TahunID":      tahunIDString,
		"ProdiID":      prodiIDString,
		"ProgramID":    programID,
		"RAWTahunID":   mahasiswa.ParseFullData()["TahunID"],
		"RAWProdiID":   mahasiswa.ParseFullData()["ProdiID"],
		"RAWProgramID": mahasiswa.ParseFullData()["ProgramID"],
	})

	if tahunIDString == "" || prodiIDString == "" || programID == "" {
		return nil, errors.New("invalid mahasiswa data: TahunID, ProdiID, or ProgramID is missing")

	}

	var prodi models.ProdiPnbp
	err := mtr.DB.Where("kode_prodi = ?", prodiIDString).
		First(prodi).Error
	if err != nil {
		return nil, errors.New("invalid prodi data: " + err.Error())
	}

	prodiID := prodi.ID

	var tagihan models.MasterTagihan
	err = mtr.DB.Where("angkatan = ? and prodi_id = ?", tahun, prodiID).
		First(&tagihan).Error

	if err != nil {
		return nil, errors.New("invalid master tagihan data: " + err.Error())
	}

	UKTString := utils.GetStringFromAny(mahasiswa.ParseFullData()["UKT"])

	utils.Log.Info("Querying detail tagihan for MasterTagihanID: ", tagihan.ID, " with UKT: ", UKTString)

	var detailTagihan models.DetailTagihan
	err = mtr.DB.Where("MasterTagihanID = ? and kel_ukt = ?", tagihan.ID, UKTString).
		First(&detailTagihan).Error

	return &detailTagihan, err
}

package repositories

import (
	"errors"
	"fmt"
	"github.com/dedegunawan/backend-ujian-telp-v5/database"
	"github.com/dedegunawan/backend-ujian-telp-v5/models"
	"github.com/dedegunawan/backend-ujian-telp-v5/utils"
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

	if err != nil {
		utils.Log.Info("GetNominalTagihanMahasiswa: error finding master tagihan for mahasiswa: ", err)
	}

	return 0
}

func (mtr *MasterTagihanRepository) FindMasterTagihanMahasiswa(mahasiswa models.Mahasiswa) (*models.DetailTagihan, error) {
	// Prioritas 1: Ambil langsung dari mahasiswa_masters di database PNBP
	var mhswMaster models.MahasiswaMaster
	err := database.DBPNBP.Preload("MasterTagihan").Where("student_id = ?", mahasiswa.MhswID).First(&mhswMaster).Error
	
	if err == nil && mhswMaster.MasterTagihanID != 0 {
		// Jika mahasiswa_masters ditemukan dan punya MasterTagihanID, gunakan langsung
		// Catatan: mhswMaster.UKT adalah nominal (int64), tapi kel_ukt di detail_tagihan adalah kelompok UKT (string)
		// Jadi kita perlu mencari detail_tagihan berdasarkan MasterTagihanID dan nominal UKT untuk mendapatkan kel_ukt
		var detailTagihan models.DetailTagihan
		err = mtr.DB.Where("master_tagihan_id = ? AND nominal = ?", mhswMaster.MasterTagihanID, mhswMaster.UKT).
			First(&detailTagihan).Error
		
		if err == nil && detailTagihan.KelUKT != nil {
			// Berhasil menemukan detail_tagihan dengan nominal yang sesuai
			utils.Log.Info("Using mahasiswa_masters data for tagihan lookup", map[string]interface{}{
				"StudentID":      mhswMaster.StudentID,
				"MasterTagihanID": mhswMaster.MasterTagihanID,
				"NominalUKT":     mhswMaster.UKT,
				"KelompokUKT":    *detailTagihan.KelUKT,
			})
			return &detailTagihan, nil
		}
		
		// Fallback: cari berdasarkan MasterTagihanID saja (ambil yang pertama)
		err = mtr.DB.Where("master_tagihan_id = ?", mhswMaster.MasterTagihanID).
			First(&detailTagihan).Error
		
		if err == nil && detailTagihan.KelUKT != nil {
			utils.Log.Warn("Detail tagihan ditemukan dengan fallback (tidak match nominal)", map[string]interface{}{
				"StudentID":      mhswMaster.StudentID,
				"MasterTagihanID": mhswMaster.MasterTagihanID,
				"NominalUKT":     mhswMaster.UKT,
				"KelompokUKT":    *detailTagihan.KelUKT,
			})
			return &detailTagihan, nil
		}
		
		utils.Log.Info("Detail tagihan tidak ditemukan dengan MasterTagihanID langsung, fallback ke lookup manual")
	}

	// Prioritas 2: Fallback ke metode lama menggunakan ParseFullData
	// (untuk kompatibilitas dengan data dari SIMAK atau data lama)
	tahunIDString := utils.GetStringFromAny(mahasiswa.ParseFullData()["TahunID"])
	if tahunIDString == "" {
		// Coba ambil dari TahunMasuk jika ada
		if tahunMasuk, ok := mahasiswa.ParseFullData()["TahunMasuk"].(float64); ok {
			tahunIDString = fmt.Sprintf("%.0f1", tahunMasuk)
		}
	}
	
	tahun := ""
	if len(tahunIDString) >= 4 {
		tahun = tahunIDString[:4] // Ambil 4 karakter pertama dari TahunID
	}

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

	if tahun == "" || prodiIDString == "" || programID == "" {
		return nil, errors.New("invalid mahasiswa data: TahunID, ProdiID, or ProgramID is missing")
	}

	var prodi models.ProdiPnbp
	err = mtr.DB.Where("kode_prodi = ?", prodiIDString).
		First(&prodi).Error
	if err != nil {
		return nil, errors.New("invalid prodi data: " + err.Error())
	}

	prodiID := prodi.ID

	var tagihan models.MasterTagihan
	err = mtr.DB.Where("angkatan = ? and prodi_id = ?", tahun, strconv.Itoa(int(prodiID))).
		First(&tagihan).Error

	if err != nil {
		return nil, errors.New("invalid master tagihan data: " + err.Error())
	}

	// UKT: prioritas dari mahasiswa_masters, fallback ke mahasiswa.UKT
	UKTString := utils.GetStringFromAny(mahasiswa.UKT)
	
	// Jika ada mhswMaster, coba ambil kelompok UKT dari detail_tagihan berdasarkan nominal
	if mhswMaster.UKT > 0 && mhswMaster.MasterTagihanID != 0 {
		var detailTagihanByNominal models.DetailTagihan
		errNominal := mtr.DB.Where("master_tagihan_id = ? AND nominal = ?", mhswMaster.MasterTagihanID, mhswMaster.UKT).
			First(&detailTagihanByNominal).Error
		if errNominal == nil && detailTagihanByNominal.KelUKT != nil {
			UKTString = *detailTagihanByNominal.KelUKT
			utils.Log.Info("Kelompok UKT diambil dari detail_tagihan berdasarkan nominal", "nominalUKT", mhswMaster.UKT, "kelompokUKT", UKTString)
		}
	}

	utils.Log.Info("Querying detail tagihan for MasterTagihanID: ", tagihan.ID, " with UKT: ", UKTString)

	var detailTagihan models.DetailTagihan
	err = mtr.DB.Where("master_tagihan_id = ? and kel_ukt = ?", tagihan.ID, UKTString).
		First(&detailTagihan).Error

	return &detailTagihan, err
}

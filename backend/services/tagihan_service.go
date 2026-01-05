package services

import (
	"fmt"
	"strconv"
	"time"

	"github.com/dedegunawan/backend-ujian-telp-v5/database"
	"github.com/dedegunawan/backend-ujian-telp-v5/models"
	"github.com/dedegunawan/backend-ujian-telp-v5/repositories"
	"github.com/dedegunawan/backend-ujian-telp-v5/utils"
)

type TagihanService interface {
	CreateNewTagihan(mahasiswa *models.Mahasiswa, financeYear *models.FinanceYear) error

	CreateNewTagihanPasca(mahasiswa *models.Mahasiswa, financeYear *models.FinanceYear) error
	HitungSemesterSaatIni(tahunIDAwal string, tahunIDSekarang string) (int, error)
	SavePaymentConfirmation(studentBill models.StudentBill, vaNumber string, paymentDate string, objectName string) (*models.PaymentConfirmation, error)

	CekCicilanMahasiswa(mahasiswa *models.Mahasiswa, financeYear *models.FinanceYear) bool
	CekPenangguhanMahasiswa(mahasiswa *models.Mahasiswa, financeYear *models.FinanceYear) bool
	CekBeasiswaMahasiswa(mahasiswa *models.Mahasiswa, financeYear *models.FinanceYear) bool
	CekDepositMahasiswa(mahasiswa *models.Mahasiswa, financeYear *models.FinanceYear) bool
	IsNominalDibayarLebihKecilSeharusnya(mahasiswa *models.Mahasiswa, financeYear *models.FinanceYear) (bool, int64, int64)
	CreateNewTagihanSekurangnya(mahasiswa *models.Mahasiswa, financeYear *models.FinanceYear, tagihanKurang int64) error
}

type tagihanService struct {
	repo                    repositories.TagihanRepository
	masterTagihanRepository repositories.MasterTagihanRepository
}

func NewTagihanService(repo repositories.TagihanRepository, masterTagihanRepository repositories.MasterTagihanRepository) TagihanService {
	return &tagihanService{repo: repo, masterTagihanRepository: masterTagihanRepository}
}

func (r *tagihanService) GetNominalBeasiswa(studentId string, academicYear string) int64 {
	var total int64

	dbEpnbp := database.DBPNBP

	err := dbEpnbp.Table("detail_beasiswa").
		Joins("JOIN beasiswa ON beasiswa.id = detail_beasiswa.beasiswa_id").
		Select("COALESCE(CAST(SUM(detail_beasiswa.nominal_beasiswa) AS SIGNED), 0)").
		Where("beasiswa.status = ?", "active").
		Where("detail_beasiswa.tahun_id = ?", academicYear).
		Where("detail_beasiswa.npm = ?", studentId).
		Scan(&total).Error

	if err != nil {
		utils.Log.Info("Error saat ambil total nominal_beasiswa:", err)
		return 0
	}

	return total

}

func (r *tagihanService) CheckDepositMahasiswa(mahasiswa *models.Mahasiswa, financeYear *models.FinanceYear) {
	//dbPnbp := database.DBPNBP

	// cek apakah sudah ada deposit yang digunakan di tahun tersebut
	//var deposit models.DepositLedgerEntry
	//dbPnbp.Where("student_id = ? AND academic_year = ? AND status = ?",)

	// jika sudah ada kembalikan hasilnya & sukses, kecuali masih ada kekurangan, buatkan tagihan baru nya

	// jika belum ada & masih punya deposit, buatkan tagihan deposit baru untuk mahasiswa tersebut

	// jika tidak punya deposit kembalikan hasil kosong & lanjutkan
}

func (r *tagihanService) GenerateCicilanMahasiswa(mahasiswa *models.Mahasiswa, financeYear *models.FinanceYear) bool {
	mhswID := string(mahasiswa.MhswID)
	financeCode := financeYear.Code
	dbEpnbp := database.DBPNBP

	var cicilanJatuhTempo []models.DetailCicilan
	today := time.Now().Format("2006-01-02") // Format YYYY-MM-DD

	err := dbEpnbp.Preload("Cicilan").
		Joins("JOIN cicilans ON cicilans.id = detail_cicilans.cicilan_id").
		Where("detail_cicilans.due_date <= ?", today).
		Where("cicilans.tahun_id = ? AND cicilans.npm = ?", financeCode, mhswID).
		Find(&cicilanJatuhTempo).Error

	if err == nil && len(cicilanJatuhTempo) > 0 {
		for _, data := range cicilanJatuhTempo {
			dt := models.StudentBill{
				StudentID:          string(mahasiswa.MhswID),
				AcademicYear:       financeYear.AcademicYear,
				BillTemplateItemID: 0,
				Name:               "Cicilan UKT",
				Amount:             data.Amount,
				PaidAmount:         0,
				CreatedAt:          time.Now(),
				UpdatedAt:          time.Now(),
			}
			r.repo.DB.Create(&dt)
		}
		return true
	}
	return false
}

func (r *tagihanService) HasCicilanMahasiswa(mahasiswa *models.Mahasiswa, financeYear *models.FinanceYear) bool {
	mhswID := string(mahasiswa.MhswID)
	financeCode := financeYear.Code
	dbEpnbp := database.DBPNBP

	var hasCicilanCount int64

	err := dbEpnbp.Preload("Cicilan").
		Joins("JOIN cicilans ON cicilans.id = detail_cicilans.cicilan_id").
		Where("cicilans.tahun_id = ? AND cicilans.npm = ?", financeCode, mhswID).
		Count(&hasCicilanCount).Error

	if err == nil && hasCicilanCount > 0 {
		return true
	}
	return false
}

// getUKTFromMahasiswaMasters mengambil kelompok UKT (kel_ukt) dari mahasiswa_masters di database PNBP
// Catatan: mhswMaster.UKT adalah nominal (int64), tapi yang dibutuhkan adalah kelompok UKT (string: "1"-"7")
func (r *tagihanService) getUKTFromMahasiswaMasters(mhswID string) (string, error) {
	var mhswMaster models.MahasiswaMaster
	err := database.DBPNBP.Where("student_id = ?", mhswID).First(&mhswMaster).Error
	if err != nil {
		return "", err
	}

	// Ambil kelompok UKT dari detail_tagihan berdasarkan MasterTagihanID dan nominal UKT
	if mhswMaster.MasterTagihanID != 0 {
		var detailTagihan models.DetailTagihan
		errDetail := database.DBPNBP.Where("master_tagihan_id = ? AND nominal = ?", mhswMaster.MasterTagihanID, mhswMaster.UKT).
			First(&detailTagihan).Error
		if errDetail == nil && detailTagihan.KelUKT != nil {
			return *detailTagihan.KelUKT, nil
		}

		// Fallback: cari berdasarkan MasterTagihanID saja
		errFallback := database.DBPNBP.Where("master_tagihan_id = ?", mhswMaster.MasterTagihanID).
			First(&detailTagihan).Error
		if errFallback == nil && detailTagihan.KelUKT != nil {
			return *detailTagihan.KelUKT, nil
		}
	}

	// Fallback terakhir: gunakan nominal sebagai string (untuk kompatibilitas)
	return strconv.Itoa(int(mhswMaster.UKT)), nil
}

// getBIPOTIDFromMahasiswaMasters mengambil BIPOTID langsung dari mahasiswa_masters -> master_tagihan di database PNBP
func (r *tagihanService) getBIPOTIDFromMahasiswaMasters(mhswID string) (string, error) {
	var mhswMaster models.MahasiswaMaster
	err := database.DBPNBP.Preload("MasterTagihan").Where("student_id = ?", mhswID).First(&mhswMaster).Error
	if err != nil {
		utils.Log.Warn("Mahasiswa tidak ditemukan di mahasiswa_masters", "mhswID", mhswID, "error", err)
		return "", fmt.Errorf("mahasiswa tidak ditemukan di mahasiswa_masters: %w", err)
	}

	if mhswMaster.MasterTagihanID == 0 {
		utils.Log.Warn("MasterTagihanID = 0 untuk mahasiswa", "mhswID", mhswID)
		return "", fmt.Errorf("master_tagihan_id tidak ditemukan untuk mahasiswa %s (MasterTagihanID=0)", mhswID)
	}

	if mhswMaster.MasterTagihan == nil {
		utils.Log.Warn("MasterTagihan nil untuk mahasiswa, mencoba load manual", "mhswID", mhswID, "MasterTagihanID", mhswMaster.MasterTagihanID)
		// Coba load manual jika Preload gagal
		var masterTagihan models.MasterTagihan
		errLoad := database.DBPNBP.Where("id = ?", mhswMaster.MasterTagihanID).First(&masterTagihan).Error
		if errLoad != nil {
			return "", fmt.Errorf("gagal load master_tagihan untuk mahasiswa %s: %w", mhswID, errLoad)
		}
		mhswMaster.MasterTagihan = &masterTagihan
	}

	if mhswMaster.MasterTagihan.BipotID == 0 {
		utils.Log.Warn("BipotID = 0 di master_tagihan", "mhswID", mhswID, "MasterTagihanID", mhswMaster.MasterTagihanID)
		return "", fmt.Errorf("bipotid tidak ditemukan di master_tagihan untuk mahasiswa %s (BipotID=0)", mhswID)
	}

	BIPOTID := strconv.Itoa(int(mhswMaster.MasterTagihan.BipotID))
	utils.Log.Info("BIPOTID diambil dari mahasiswa_masters", "mhswID", mhswID, "BIPOTID", BIPOTID, "MasterTagihanID", mhswMaster.MasterTagihanID)
	return BIPOTID, nil
}

func (r *tagihanService) CreateNewTagihan(mahasiswa *models.Mahasiswa, financeYear *models.FinanceYear) error {
	utils.Log.Info("CreateNewTagihan dimulai", map[string]interface{}{
		"mhswID":       mahasiswa.MhswID,
		"nama":         mahasiswa.Nama,
		"BIPOTID":      mahasiswa.BIPOTID,
		"UKT":          mahasiswa.UKT,
		"academicYear": financeYear.AcademicYear,
	})

	// interception: jika mahasiswa memiliki data cicilan generate dari cicilan tersebut
	hasCicilan := r.GenerateCicilanMahasiswa(mahasiswa, financeYear)
	if hasCicilan {
		utils.Log.Info("Mahasiswa memiliki cicilan, generate dari cicilan", "mhswID", mahasiswa.MhswID)
		return nil
	}

	// Ambil data dari mahasiswa_masters -> master_tagihan -> detail_tagihan
	// JANGAN gunakan mahasiswa.BIPOTID atau mahasiswa.UKT, ambil langsung dari mahasiswa_masters
	utils.Log.Info("Mengambil data dari mahasiswa_masters", "mhswID", mahasiswa.MhswID)

	var mhswMaster models.MahasiswaMaster
	err := database.DBPNBP.Preload("MasterTagihan").Where("student_id = ?", mahasiswa.MhswID).First(&mhswMaster).Error
	if err != nil {
		utils.Log.Error("Mahasiswa tidak ditemukan di mahasiswa_masters", "mhswID", mahasiswa.MhswID, "error", err)
		return fmt.Errorf("mahasiswa tidak ditemukan di mahasiswa_masters untuk %s: %w", mahasiswa.MhswID, err)
	}

	if mhswMaster.MasterTagihanID == 0 {
		utils.Log.Error("MasterTagihanID = 0 untuk mahasiswa", "mhswID", mahasiswa.MhswID)
		return fmt.Errorf("master_tagihan_id tidak ditemukan untuk mahasiswa %s (MasterTagihanID=0)", mahasiswa.MhswID)
	}

	// Load master_tagihan jika belum ter-load
	if mhswMaster.MasterTagihan == nil {
		utils.Log.Info("MasterTagihan nil, mencoba load manual", "mhswID", mahasiswa.MhswID, "MasterTagihanID", mhswMaster.MasterTagihanID)
		var masterTagihan models.MasterTagihan
		errLoad := database.DBPNBP.Where("id = ?", mhswMaster.MasterTagihanID).First(&masterTagihan).Error
		if errLoad != nil {
			utils.Log.Error("Gagal load master_tagihan", "mhswID", mahasiswa.MhswID, "MasterTagihanID", mhswMaster.MasterTagihanID, "error", errLoad)
			return fmt.Errorf("gagal load master_tagihan untuk mahasiswa %s: %w", mahasiswa.MhswID, errLoad)
		}
		mhswMaster.MasterTagihan = &masterTagihan
	}

	// Catatan: BIPOTID tidak diperlukan lagi karena kita tidak menggunakan bill_template
	// Langsung menggunakan detail_tagihan dari master_tagihan_id
	if mhswMaster.MasterTagihan.BipotID == 0 {
		utils.Log.Warn("BipotID = 0 di master_tagihan, lanjut tanpa BIPOTID (tidak diperlukan)", map[string]interface{}{
			"mhswID":          mahasiswa.MhswID,
			"MasterTagihanID": mhswMaster.MasterTagihanID,
		})
	} else {
		BIPOTID := strconv.Itoa(int(mhswMaster.MasterTagihan.BipotID))
		utils.Log.Info("BIPOTID diambil dari master_tagihan (untuk referensi)", map[string]interface{}{
			"mhswID":          mahasiswa.MhswID,
			"BIPOTID":         BIPOTID,
			"MasterTagihanID": mhswMaster.MasterTagihanID,
		})
	}

	// Ambil detail_tagihan untuk mendapatkan kel_ukt berdasarkan master_tagihan_id dan UKT nominal
	var detailTagihan models.DetailTagihan
	errDetail := database.DBPNBP.Where("master_tagihan_id = ? AND nominal = ?", mhswMaster.MasterTagihanID, mhswMaster.UKT).
		First(&detailTagihan).Error

	var UKT string // Kelompok UKT (kel_ukt) dari detail_tagihan
	if errDetail == nil && detailTagihan.KelUKT != nil {
		UKT = *detailTagihan.KelUKT
		utils.Log.Info("Kelompok UKT ditemukan dari detail_tagihan", map[string]interface{}{
			"mhswID":          mahasiswa.MhswID,
			"kelompokUKT":     UKT,
			"nominalUKT":      mhswMaster.UKT,
			"masterTagihanID": mhswMaster.MasterTagihanID,
		})
	} else {
		// Fallback: cari berdasarkan master_tagihan_id saja (ambil yang pertama)
		utils.Log.Warn("Detail tagihan tidak ditemukan dengan nominal, mencoba fallback", map[string]interface{}{
			"mhswID":          mahasiswa.MhswID,
			"nominalUKT":      mhswMaster.UKT,
			"masterTagihanID": mhswMaster.MasterTagihanID,
			"error":           errDetail,
		})
		errFallback := database.DBPNBP.Where("master_tagihan_id = ?", mhswMaster.MasterTagihanID).
			First(&detailTagihan).Error
		if errFallback == nil && detailTagihan.KelUKT != nil {
			UKT = *detailTagihan.KelUKT
			utils.Log.Warn("Kelompok UKT diambil dari detail_tagihan fallback (tidak match nominal)", map[string]interface{}{
				"mhswID":          mahasiswa.MhswID,
				"kelompokUKT":     UKT,
				"nominalUKT":      mhswMaster.UKT,
				"masterTagihanID": mhswMaster.MasterTagihanID,
			})
		} else {
			// Fallback terakhir: gunakan nominal sebagai string
			UKT = strconv.Itoa(int(mhswMaster.UKT))
			utils.Log.Warn("Kelompok UKT tidak ditemukan, menggunakan nominal sebagai string", map[string]interface{}{
				"mhswID":     mahasiswa.MhswID,
				"UKT":        UKT,
				"nominalUKT": mhswMaster.UKT,
			})
		}
	}

	// Ambil semua detail_tagihan dari master_tagihan_id yang sesuai dengan UKT nominal
	// JANGAN gunakan bill_template atau bill_template_items, langsung dari detail_tagihan
	utils.Log.Info("Mencari detail_tagihan dari master_tagihan", map[string]interface{}{
		"masterTagihanID": mhswMaster.MasterTagihanID,
		"nominalUKT":      mhswMaster.UKT,
		"kelompokUKT":     UKT,
	})

	var detailTagihans []models.DetailTagihan
	// Ambil semua detail_tagihan yang sesuai dengan master_tagihan_id dan kel_ukt
	errDetailList := database.DBPNBP.Where("master_tagihan_id = ? AND kel_ukt = ?", mhswMaster.MasterTagihanID, UKT).
		Find(&detailTagihans).Error

	if errDetailList != nil {
		utils.Log.Error("Gagal query detail_tagihan", map[string]interface{}{
			"masterTagihanID": mhswMaster.MasterTagihanID,
			"kelompokUKT":     UKT,
			"error":           errDetailList.Error(),
		})
		return fmt.Errorf("gagal query detail_tagihan untuk master_tagihan_id %d dan kel_ukt %s: %w", mhswMaster.MasterTagihanID, UKT, errDetailList)
	}

	if len(detailTagihans) == 0 {
		// Fallback: cari berdasarkan master_tagihan_id saja (tanpa filter kel_ukt)
		utils.Log.Warn("Detail tagihan tidak ditemukan dengan kel_ukt, mencoba tanpa filter kel_ukt", map[string]interface{}{
			"masterTagihanID": mhswMaster.MasterTagihanID,
			"kelompokUKT":     UKT,
		})
		errDetailFallback := database.DBPNBP.Where("master_tagihan_id = ?", mhswMaster.MasterTagihanID).
			Find(&detailTagihans).Error
		if errDetailFallback != nil {
			utils.Log.Error("Gagal query detail_tagihan (fallback)", map[string]interface{}{
				"masterTagihanID": mhswMaster.MasterTagihanID,
				"error":           errDetailFallback.Error(),
			})
			return fmt.Errorf("gagal query detail_tagihan untuk master_tagihan_id %d: %w", mhswMaster.MasterTagihanID, errDetailFallback)
		}
	}

	if len(detailTagihans) == 0 {
		utils.Log.Error("Tidak ada detail_tagihan yang ditemukan", map[string]interface{}{
			"masterTagihanID": mhswMaster.MasterTagihanID,
			"kelompokUKT":     UKT,
			"mhswID":          mahasiswa.MhswID,
		})
		return fmt.Errorf("tidak ada detail_tagihan yang ditemukan untuk master_tagihan_id %d dan kel_ukt %s (mahasiswa %s)", mhswMaster.MasterTagihanID, UKT, mahasiswa.MhswID)
	}

	utils.Log.Info("Detail tagihan ditemukan", "count", len(detailTagihans), "masterTagihanID", mhswMaster.MasterTagihanID, "kelompokUKT", UKT)

	nominalBeasiswa := r.GetNominalBeasiswa(string(mahasiswa.MhswID), financeYear.AcademicYear)
	utils.Log.Info("nominalBeasiswa:", nominalBeasiswa)

	sisaBeasiswa := nominalBeasiswa
	// Generate StudentBill langsung dari detail_tagihan
	for _, dt := range detailTagihans {
		nominalBeasiswaSaatIni := int64(0)
		nominalTagihan := dt.Nominal
		if sisaBeasiswa > 0 && sisaBeasiswa >= dt.Nominal {
			sisaBeasiswa = sisaBeasiswa - dt.Nominal
			nominalBeasiswaSaatIni = dt.Nominal
			nominalTagihan = 0
		} else if sisaBeasiswa > 0 {
			nominalBeasiswaSaatIni = sisaBeasiswa
			nominalTagihan = dt.Nominal - nominalBeasiswaSaatIni
		}

		bill := models.StudentBill{
			StudentID:          string(mahasiswa.MhswID),
			AcademicYear:       financeYear.AcademicYear,
			BillTemplateItemID: 0, // Tidak menggunakan bill_template_item lagi
			Name:               dt.Nama,
			Amount:             nominalTagihan,
			Beasiswa:           nominalBeasiswaSaatIni,
			PaidAmount:         0,
			CreatedAt:          time.Now(),
			UpdatedAt:          time.Now(),
		}

		if err := r.repo.DB.Create(&bill).Error; err != nil {
			utils.Log.Error("Gagal membuat StudentBill", map[string]interface{}{
				"mhswID":          mahasiswa.MhswID,
				"detailTagihanID": dt.ID,
				"nama":            dt.Nama,
				"nominal":         dt.Nominal,
				"error":           err.Error(),
			})
			return fmt.Errorf("gagal membuat tagihan mahasiswa dari detail_tagihan ID %d: %w", dt.ID, err)
		}

		utils.Log.Info("StudentBill berhasil dibuat dari detail_tagihan", map[string]interface{}{
			"mhswID":          mahasiswa.MhswID,
			"detailTagihanID": dt.ID,
			"nama":            dt.Nama,
			"nominal":         dt.Nominal,
			"amount":          nominalTagihan,
			"beasiswa":        nominalBeasiswaSaatIni,
		})
	}

	return nil
}
func (r *tagihanService) CreateNewTagihanPasca(mahasiswa *models.Mahasiswa, financeYear *models.FinanceYear) error {
	var template models.BillTemplate

	// Ambil bill_template berdasarkan BIPOTID mahasiswa
	if err := r.repo.DB.
		Where("code = ?", mahasiswa.BIPOTID).
		First(&template).Error; err != nil {
		return fmt.Errorf("bill template not found for BIPOTID %s: %w", mahasiswa.BIPOTID, err)
	}

	// Ambil semua item UKT yang cocok
	var items []models.BillTemplateItem
	if err := r.repo.DB.
		Where(`bill_template_id = ?`, template.ID).
		Find(&items).Error; err != nil {
		return fmt.Errorf("bill_template_items not found for UKT %s: %w", mahasiswa.UKT, err)
	}

	if len(items) == 0 {
		utils.Log.Info("Last query : ", `bill_template_id = ?`, template.ID, mahasiswa.UKT)
		return fmt.Errorf("tidak ada item tagihan yang cocok untuk UKT %s", mahasiswa.UKT)
	}

	mhswID := mahasiswa.MhswID
	// Ambil SemesterMasukID dari mahasiswa_masters di database PNBP
	// SemesterMasukID adalah referensi ke budget_periods.id, bukan semester masuk (1/2)
	var mhswMaster models.MahasiswaMaster
	errMhswMaster := database.DBPNBP.Where("student_id = ?", mhswID).First(&mhswMaster).Error

	var tahunIDAwal string

	if errMhswMaster == nil && mhswMaster.SemesterMasukID > 0 {
		// SemesterMasukID adalah referensi ke budget_periods.id
		// Ambil budget_periods.kode sebagai tahunIDAwal
		var budgetPeriod models.BudgetPeriod
		errBudgetPeriod := database.DBPNBP.Where("id = ?", mhswMaster.SemesterMasukID).First(&budgetPeriod).Error
		if errBudgetPeriod == nil && budgetPeriod.Kode != "" {
			tahunIDAwal = budgetPeriod.Kode
			utils.Log.Info("CreateNewTagihanPasca: TahunID awal diambil dari budget_periods berdasarkan SemesterMasukID", map[string]interface{}{
				"mhswID":           mhswID,
				"SemesterMasukID":  mhswMaster.SemesterMasukID,
				"budgetPeriodID":   budgetPeriod.ID,
				"budgetPeriodKode": budgetPeriod.Kode,
				"tahunIDAwal":      tahunIDAwal,
			})
		} else {
			utils.Log.Warn("CreateNewTagihanPasca: Budget period tidak ditemukan berdasarkan SemesterMasukID, fallback ke TahunMasuk", map[string]interface{}{
				"mhswID":          mhswID,
				"SemesterMasukID": mhswMaster.SemesterMasukID,
				"error":           errBudgetPeriod,
			})
			// Fallback: gunakan TahunMasuk dengan default semester 1
			if mhswMaster.TahunMasuk > 0 {
				tahunIDAwal = fmt.Sprintf("%d1", mhswMaster.TahunMasuk)
			} else {
				return fmt.Errorf("tahun masuk tidak ditemukan untuk mahasiswa %s", mhswID)
			}
		}
	} else {
		// Fallback: ambil dari FullData
		if tahunIDData, ok := mahasiswa.ParseFullData()["TahunID"].(string); ok && tahunIDData != "" {
			tahunIDAwal = tahunIDData
			utils.Log.Info("CreateNewTagihanPasca: TahunID awal diambil dari FullData", "mhswID", mhswID, "tahunIDAwal", tahunIDAwal)
		} else if tahunMasukData, ok := mahasiswa.ParseFullData()["TahunMasuk"].(float64); ok {
			// Fallback: buat dari TahunMasuk dengan default semester 1
			tahunIDAwal = fmt.Sprintf("%.0f1", tahunMasukData)
			utils.Log.Info("CreateNewTagihanPasca: TahunID awal dibuat dari TahunMasuk di FullData", "mhswID", mhswID, "tahunIDAwal", tahunIDAwal)
		} else {
			// Fallback terakhir: estimasi dari NPM
			if len(mhswID) >= 2 {
				tahunMasukStr := "20" + mhswID[0:2] + "1"
				tahunIDAwal = tahunMasukStr
				utils.Log.Info("CreateNewTagihanPasca: TahunID awal diestimasi dari NPM", "mhswID", mhswID, "tahunIDAwal", tahunIDAwal)
			} else {
				return fmt.Errorf("tahun masuk tidak ditemukan untuk mahasiswa %s", mhswID)
			}
		}
	}

	financeCode := financeYear.Code
	semesterSaatIni, err := r.HitungSemesterSaatIni(tahunIDAwal, financeCode)
	if err != nil {
		return err
	}

	// Generate StudentBill berdasarkan item
	for _, item := range items {
		endSesi := item.MulaiSesi + item.KaliSesi - 1
		utils.Log.Info(" mulai Sesi, ", item.MulaiSesi, "endSesi: ", endSesi, "semester saat ini ", semesterSaatIni)
		matchSesi := int64(item.MulaiSesi) <= int64(semesterSaatIni) && int64(semesterSaatIni) <= endSesi
		broadSesi := item.MulaiSesi > 0 && item.KaliSesi == 0 && int64(item.MulaiSesi) <= int64(semesterSaatIni)
		if matchSesi || broadSesi {
			bill := models.StudentBill{
				StudentID:          string(mahasiswa.MhswID),
				AcademicYear:       financeYear.AcademicYear,
				BillTemplateItemID: item.BillTemplateID,
				Name:               item.AdditionalName,
				Amount:             item.Amount,
				PaidAmount:         0,
				CreatedAt:          time.Now(),
				UpdatedAt:          time.Now(),
			}
			if err := r.repo.DB.Create(&bill).Error; err != nil {
				return fmt.Errorf("gagal membuat tagihan mahasiswa: %w", err)
			}
		}

	}

	return nil
}

// getTahunIDFromMahasiswaMasters mengambil TahunID langsung dari mahasiswa_masters di database PNBP
func getTahunIDFromMahasiswaMasters(mhswID string) string {
	var mhswMaster models.MahasiswaMaster
	err := database.DBPNBP.Where("student_id = ?", mhswID).First(&mhswMaster).Error
	if err != nil {
		return ""
	}

	// TahunMasuk adalah int (contoh: 2023)
	// SemesterMasukID adalah uint (1 = Ganjil, 2 = Genap, atau sesuai enum)
	// Format TahunID: YYYYS (tahun + semester)
	// Jika SemesterMasukID tidak ada, default ke semester 1 (Ganjil)
	semesterMasuk := 1
	if mhswMaster.SemesterMasukID > 0 {
		semesterMasuk = int(mhswMaster.SemesterMasukID)
		// Pastikan semester hanya 1 atau 2
		if semesterMasuk > 2 {
			semesterMasuk = 1
		}
	}

	if mhswMaster.TahunMasuk > 0 {
		TahunID := fmt.Sprintf("%d%d", mhswMaster.TahunMasuk, semesterMasuk)
		utils.Log.Info("TahunID diambil dari mahasiswa_masters", "mhswID", mhswID, "TahunMasuk", mhswMaster.TahunMasuk, "SemesterMasukID", mhswMaster.SemesterMasukID, "TahunID", TahunID)
		return TahunID
	}

	return ""
}

func getTahunIDFormParsed(mahasiswa *models.Mahasiswa) string {
	data := mahasiswa.ParseFullData()

	// Coba ambil TahunID langsung
	tahunRaw, exists := data["TahunID"]
	if exists {
		var TahunID string
		switch v := tahunRaw.(type) {
		case string:
			TahunID = v
		case float64:
			TahunID = fmt.Sprintf("%.0f", v)
		case int:
			TahunID = strconv.Itoa(v)
		default:
			utils.Log.Info("TahunID ditemukan tapi tipe tidak dikenali", "value", tahunRaw)
			return ""
		}
		if TahunID != "" {
			return TahunID
		}
	}

	// Fallback: coba ambil dari TahunMasuk jika ada
	if tahunMasuk, ok := data["TahunMasuk"].(float64); ok {
		TahunID := fmt.Sprintf("%.0f1", tahunMasuk) // Default semester 1
		utils.Log.Info("TahunID dibuat dari TahunMasuk", "TahunMasuk", tahunMasuk, "TahunID", TahunID)
		return TahunID
	}

	utils.Log.Info("Field TahunID tidak ditemukan pada data mahasiswa", "data", data)
	return ""

}

// HitungSemesterSaatIni menghitung semester saat ini berdasarkan tahun masuk (angkatan) dan budget_periods.kode
// Rumus: semester = (tahun sekarang - tahun masuk) * 2 + semester sekarang - semester masuk + 1
// Contoh: budget_periods.kode = 20252, angkatan = 2024, semester masuk = 1 → semester = (2025 - 2024) * 2 + 2 - 1 + 1 = 4
// Contoh: budget_periods.kode = 20252, angkatan = 2025, semester masuk = 1 → semester = (2025 - 2025) * 2 + 2 - 1 + 1 = 2
func (r *tagihanService) HitungSemesterSaatIni(tahunIDAwal string, tahunIDSekarang string) (int, error) {
	utils.Log.Info("HitungSemesterSaatIni", map[string]interface{}{
		"tahunIDAwal":     tahunIDAwal,
		"tahunIDSekarang": tahunIDSekarang,
	})

	if len(tahunIDSekarang) != 5 {
		return 0, fmt.Errorf("format tahunIDSekarang tidak valid, harus 5 digit seperti 20251")
	}

	// Parsing tahun dan semester dari budget_periods.kode (tahunIDSekarang)
	tahunSekarang, err1 := strconv.Atoi(tahunIDSekarang[:4])
	semesterSekarang, err2 := strconv.Atoi(tahunIDSekarang[4:])

	if err1 != nil || err2 != nil {
		return 0, fmt.Errorf("gagal parsing tahun atau semester dari budget_periods.kode: %v, %v", err1, err2)
	}

	// Ambil tahun masuk (angkatan) dan semester masuk dari tahunIDAwal
	var tahunMasuk, semesterMasuk int
	if len(tahunIDAwal) == 5 {
		// Jika tahunIDAwal format YYYYS, ambil tahun dan semester masuknya
		tahunMasuk, err1 = strconv.Atoi(tahunIDAwal[:4])
		semesterMasuk, err2 = strconv.Atoi(tahunIDAwal[4:])
		if err1 != nil || err2 != nil {
			return 0, fmt.Errorf("gagal parsing tahun atau semester masuk dari tahunIDAwal: %v, %v", err1, err2)
		}
	} else {
		// Jika tahunIDAwal hanya tahun (4 digit), default semester masuk = 1
		tahunMasuk, err1 = strconv.Atoi(tahunIDAwal)
		semesterMasuk = 1 // Default semester masuk = 1 (Ganjil)
		if err1 != nil {
			return 0, fmt.Errorf("gagal parsing tahun masuk: %v", err1)
		}
	}

	// Rumus: semester = (tahun sekarang - tahun masuk) * 2 + semester sekarang - semester masuk + 1
	selisihTahun := tahunSekarang - tahunMasuk
	semester := (selisihTahun * 2) + semesterSekarang - semesterMasuk + 1

	utils.Log.Info("Perhitungan semester", map[string]interface{}{
		"tahunMasuk":       tahunMasuk,
		"semesterMasuk":    semesterMasuk,
		"tahunSekarang":    tahunSekarang,
		"semesterSekarang": semesterSekarang,
		"selisihTahun":     selisihTahun,
		"semester":         semester,
	})

	return semester, nil
}

func (r *tagihanService) SavePaymentConfirmation(studentBill models.StudentBill, vaNumber string, paymentDate string, objectName string) (*models.PaymentConfirmation, error) {
	paymentConfirmation := models.PaymentConfirmation{
		StudentBillID: studentBill.ID,
		VaNumber:      vaNumber,
		PaymentDate:   paymentDate,
		ObjectName:    objectName,
		Message:       "",
	}
	r.repo.DB.Save(&paymentConfirmation)

	// check all payment id is success or not
	payUrls, err := r.repo.GetAllPayUrlByStudentBillID(studentBill.ID)
	if err != nil {
		return nil, err
	}

	epnbpRepo := repositories.NewEpnbpRepository(r.repo.DB)
	eService := NewEpnbpService(epnbpRepo)

	var realPaymentDate *time.Time
	isPaid := false
	invoiceIds := []string{}
	for _, payUrl := range payUrls {
		invoiceId := strconv.FormatUint(uint64(payUrl.InvoiceID), 10)
		isPaid, realPaymentDate = eService.CheckStatusPaidByInvoiceID(invoiceId)
		invoiceIds = append(invoiceIds, invoiceId)
		if isPaid {
			break
		}
	}
	if !isPaid {
		isPaid, realPaymentDate = eService.CheckStatusPaidByVirtualAccount(vaNumber, invoiceIds)
	}

	if isPaid {
		r.savePaidStudentBill(studentBill, studentBill.Amount, *realPaymentDate, vaNumber, objectName)
		return &paymentConfirmation, nil
	}

	return nil, nil
}

func (r *tagihanService) savePaidStudentBill(studentBill models.StudentBill, amount int64, realPaymentDate time.Time, vaNumber string, objectName string) bool {
	studentBill.PaidAmount = amount
	r.repo.DB.Save(&studentBill)

	studentPayment := models.StudentPayment{
		StudentID:    string(studentBill.StudentID),
		AcademicYear: studentBill.AcademicYear,
		PaymentRef:   vaNumber,
		Amount:       amount,
		Bank:         "",
		Method:       "VA",
		Note:         objectName,
		Date:         realPaymentDate,
	}
	r.repo.DB.Save(&studentPayment)

	studentPaymentAllocation := models.StudentPaymentAllocation{
		StudentPaymentID: studentPayment.ID,
		StudentBillID:    studentBill.ID,
		Amount:           amount,
	}
	r.repo.DB.Save(&studentPaymentAllocation)

	return true

}

func (r *tagihanService) CekCicilanMahasiswa(mahasiswa *models.Mahasiswa, financeYear *models.FinanceYear) bool {
	mhswID := string(mahasiswa.MhswID)
	financeCode := financeYear.AcademicYear
	var hasCicilanCount int64
	dbEpnbp := database.DBPNBP
	_ = dbEpnbp.Where("npm = ? AND tahun_id = ?", mhswID, financeCode).Model(&models.Cicilan{}).Count(&hasCicilanCount).Error

	if hasCicilanCount > 0 {
		return true
	}

	return false
}

func (r *tagihanService) CekPenangguhanMahasiswa(mahasiswa *models.Mahasiswa, financeYear *models.FinanceYear) bool {

	mhswID := string(mahasiswa.MhswID)
	financeCode := financeYear.AcademicYear
	var hasDepositDebitCount int64
	dbEpnbp := database.DBPNBP
	err := dbEpnbp.Where("npm = ? AND tahun_id = ? and direction = ?", mhswID, financeCode, "debit").
		Model(&models.DepositLedgerEntry{}).Count(&hasDepositDebitCount).Error

	if err != nil {
		utils.Log.Error("Error checking deposit debit count:", err)
		return false

	}

	utils.Log.Info("Has Deposit Debit Count for Mahasiswa:", mhswID, "Finance Year:", financeCode, "Count:", hasDepositDebitCount)

	if hasDepositDebitCount > 0 {
		return true
	}

	return false
}

func (r *tagihanService) CekBeasiswaMahasiswa(mahasiswa *models.Mahasiswa, financeYear *models.FinanceYear) bool {

	mhswID := string(mahasiswa.MhswID)
	financeCode := financeYear.AcademicYear
	var hasBeasiswaCount int64
	dbEpnbp := database.DBPNBP
	_ = dbEpnbp.Where("npm = ? AND tahun_id = ?", mhswID, financeCode).
		Model(&models.DetailBeasiswa{}).Count(&hasBeasiswaCount).Error

	if hasBeasiswaCount > 0 {
		return true
	}

	return false
}

func (r *tagihanService) CekDepositMahasiswa(mahasiswa *models.Mahasiswa, financeYear *models.FinanceYear) bool {
	return false
}

func (r *tagihanService) IsNominalDibayarLebihKecilSeharusnya(mahasiswa *models.Mahasiswa, financeYear *models.FinanceYear) (bool, int64, int64) {
	// seharusnya diambil dari BillTemplateItem
	tagihanSeharusnya := r.masterTagihanRepository.GetNominalTagihanMahasiswa(*mahasiswa)

	// ambil nominal tagihan yang sudah dibayar oleh mahasiswa
	totalTagihanDibayar := r.repo.GetTotalStudentBill(mahasiswa.MhswID, financeYear.AcademicYear)
	utils.Log.Info("Tagihan seharusnya:", tagihanSeharusnya, " Total tagihan dibayar:", totalTagihanDibayar)

	return totalTagihanDibayar < tagihanSeharusnya, tagihanSeharusnya, totalTagihanDibayar
}

func (r *tagihanService) CreateNewTagihanSekurangnya(mahasiswa *models.Mahasiswa, financeYear *models.FinanceYear, tagihanKurang int64) error {
	studentBill := models.StudentBill{
		StudentID:          string(mahasiswa.MhswID),
		AcademicYear:       financeYear.AcademicYear,
		BillTemplateItemID: 0, // Asumsikan tidak ada item template yang
		Name:               "UKT",
		Amount:             tagihanKurang,
		PaidAmount:         0,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	utils.Log.Info("Membuat tagihan mahasiswa dengan nominal kurang:", tagihanKurang)

	if err := r.repo.DB.Create(&studentBill).Error; err != nil {
		return fmt.Errorf("gagal membuat tagihan mahasiswa: %w", err)
	}

	return nil
}
